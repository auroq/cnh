// Package reports is the use case behind "Were you here?": it validates what an attendee says about a place,
// rate limits it, stores it, and turns the live reports on a place into the status the answer says out loud.
package reports

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/auroq/cnh/pkg/model"
	"go.opentelemetry.io/otel/metric"
)

// Store is what reports needs from a backend, in the app's own terms.
type Store interface {
	// Save writes one report so the next answer sees it.
	Save(ctx context.Context, r model.Report) error
	// Since returns every report on the given places made at or after since, hidden ones included.
	Since(ctx context.Context, placeIDs []string, since time.Time) ([]model.Report, error)
	// PlaceKind names the kind of a known place, or returns "" when the id is unknown.
	PlaceKind(ctx context.Context, id string) (string, error)
}

// Submission is the form as the attendee sent it; IP and Token are only for the rate limits.
type Submission struct {
	PlaceID  string
	Verdict  model.Verdict
	Note     string
	Token    string
	IP       string
	Honeypot string
}

// Outcome is what an accepted submission became.
type Outcome struct {
	ID   string
	Kind string
}

var (
	ErrUnknownPlace = errors.New("unknown place")
	ErrVerdict      = errors.New("verdict must be works, problem or not-as-described")
	ErrNoteRequired = errors.New("note required")
	ErrNoteLength   = errors.New("note must be 10 to 280 characters")
	ErrNoteURL      = errors.New("note may not contain a link")
	ErrNoteBlocked  = errors.New("note contains a blocked word")
	ErrHoneypot     = errors.New("honeypot filled")
	ErrToken        = errors.New("token missing")
	ErrRateLimited  = errors.New("rate limited")
)

const (
	MinNote        = 10
	MaxNote        = 280
	Expiry         = 6 * time.Hour
	TokenLimit     = 3
	IPLimit        = 20
	LimitWindow    = 10 * time.Minute
	SourceAttendee = "attendee"
)

var urlPattern = regexp.MustCompile(`(?i)(https?://|www\.)`)

// Service validates, rate limits, stores and derives. Now is injectable for tests.
type Service struct {
	Store        Store
	BlockedWords []string
	Meter        metric.MeterProvider
	Now          func() time.Time

	once    sync.Once
	byToken *Limiter
	byIP    *Limiter
	inst    *instruments
}

func (s *Service) now() time.Time {
	if s.Now == nil {
		return time.Now().UTC()
	}
	return s.Now()
}

func (s *Service) limiters() (*Limiter, *Limiter) {
	s.once.Do(func() {
		s.byToken = NewLimiter(TokenLimit, LimitWindow)
		s.byIP = NewLimiter(IPLimit, LimitWindow)
	})
	return s.byToken, s.byIP
}

// Submit checks a submission against every guard rail and stores it when it passes.
func (s *Service) Submit(ctx context.Context, sub Submission) (Outcome, error) {
	now := s.now()
	out, err := s.submit(ctx, sub, now)
	if err != nil {
		s.rejected(ctx, err)
		return out, err
	}
	s.received(ctx, sub.Verdict, out.Kind)
	return out, nil
}

func (s *Service) submit(ctx context.Context, sub Submission, now time.Time) (Outcome, error) {
	if sub.Honeypot != "" {
		return Outcome{}, ErrHoneypot
	}
	if sub.Token == "" {
		return Outcome{}, ErrToken
	}
	note := strings.TrimSpace(sub.Note)
	if err := s.validate(sub.Verdict, note); err != nil {
		return Outcome{}, err
	}
	kind, err := s.Store.PlaceKind(ctx, sub.PlaceID)
	if err != nil {
		return Outcome{}, fmt.Errorf("look up place %q: %w", sub.PlaceID, err)
	}
	if kind == "" {
		return Outcome{}, ErrUnknownPlace
	}
	byToken, byIP := s.limiters()
	if !byIP.Allow(sub.IP, now) || !byToken.Allow(sub.Token, now) {
		return Outcome{}, ErrRateLimited
	}
	r := model.Report{ID: newID(), PlaceID: sub.PlaceID, Verdict: sub.Verdict, Note: note, At: now, Token: sub.Token, Source: SourceAttendee}
	if err := s.Store.Save(ctx, r); err != nil {
		return Outcome{}, fmt.Errorf("save report: %w", err)
	}
	return Outcome{ID: r.ID, Kind: kind}, nil
}

func (s *Service) validate(verdict model.Verdict, note string) error {
	switch verdict {
	case model.VerdictWorks, model.VerdictProblem, model.VerdictNotAsDescribed:
	default:
		return ErrVerdict
	}
	if note == "" {
		if verdict != model.VerdictWorks {
			return ErrNoteRequired
		}
		return nil
	}
	if n := len([]rune(note)); n < MinNote || n > MaxNote {
		return ErrNoteLength
	}
	if urlPattern.MatchString(note) {
		return ErrNoteURL
	}
	lower := strings.ToLower(note)
	for _, w := range s.BlockedWords {
		if w = strings.ToLower(strings.TrimSpace(w)); w != "" && strings.Contains(lower, w) {
			return ErrNoteBlocked
		}
	}
	return nil
}

// StatusFor derives the attendee status of each place that has a live report; places without one are absent.
func (s *Service) StatusFor(ctx context.Context, placeIDs []string) (map[string]model.AttendeeStatus, error) {
	out := map[string]model.AttendeeStatus{}
	if len(placeIDs) == 0 {
		return out, nil
	}
	now := s.now()
	reports, err := s.Store.Since(ctx, placeIDs, now.Add(-Expiry))
	if err != nil {
		return nil, fmt.Errorf("reports since: %w", err)
	}
	byPlace := map[string][]model.Report{}
	for _, r := range reports {
		byPlace[r.PlaceID] = append(byPlace[r.PlaceID], r)
	}
	for id, rs := range byPlace {
		st := Derive(rs, now)
		if st.Count > 0 {
			out[id] = st
		}
	}
	return out, nil
}

// Derive is the trust rule: a problem is shown at once, a dispute is shown at once,
// and "works" counts as verified only when two different attendees said so within six hours.
func Derive(reports []model.Report, now time.Time) model.AttendeeStatus {
	cutoff := now.Add(-Expiry)
	var problems, disputes, works []model.Report
	for _, r := range reports {
		if r.Hidden || r.At.Before(cutoff) {
			continue
		}
		switch r.Verdict {
		case model.VerdictProblem:
			problems = append(problems, r)
		case model.VerdictNotAsDescribed:
			disputes = append(disputes, r)
		case model.VerdictWorks:
			works = append(works, r)
		}
	}
	switch {
	case len(problems) > 0:
		return statusOf(model.StatusProblem, problems)
	case len(disputes) > 0:
		return statusOf(model.StatusDisputed, disputes)
	case len(works) == 0:
		return model.AttendeeStatus{Status: model.StatusNone}
	}
	st := statusOf(model.StatusNone, works)
	if st.Count >= 2 {
		st.Status = model.StatusVerified
	}
	return st
}

func statusOf(status string, rs []model.Report) model.AttendeeStatus {
	tokens := map[string]bool{}
	st := model.AttendeeStatus{Status: status}
	for _, r := range rs {
		tokens[r.Token] = true
		if !r.At.Before(st.LatestAt) {
			st.LatestAt, st.LatestNote = r.At, r.Note
		}
	}
	st.Count = len(tokens)
	return st
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// NewToken is the anonymous per-browser id: 128 random bits, hex.
func NewToken() string { return newID() }

// Limiter is a sliding-window counter per key, in memory, per process.
type Limiter struct {
	max    int
	window time.Duration
	mu     sync.Mutex
	seen   map[string][]time.Time
}

func NewLimiter(max int, window time.Duration) *Limiter {
	return &Limiter{max: max, window: window, seen: map[string][]time.Time{}}
}

// Allow records an attempt at now and reports whether it is within the limit.
func (l *Limiter) Allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	cutoff := now.Add(-l.window)
	kept := l.seen[key][:0]
	for _, t := range l.seen[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= l.max {
		l.seen[key] = kept
		return false
	}
	l.seen[key] = append(kept, now)
	return true
}
