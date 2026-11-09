package ask

import (
	"context"
	"math"
	"strings"
	"time"
)

// Why sessions are in the answer, as /api/ask reports it.
const (
	ShownForSessionQuestion = "question_about_sessions"
	ShownForTitleMatch      = "title_match"
	SessionsHidden          = "none"
	ShownIfTitleMatches     = "if_title_matches"
)

const (
	SessionUpcoming = "upcoming"
	SessionRunning  = "running"
	SessionEnded    = "ended"
)

const (
	DistanceSameLevel     = "same_level"
	DistanceSameBuilding  = "same_building"
	DistanceOtherBuilding = "other_building"
	DistanceUnknown       = "unknown"
)

const (
	windowBefore = 5 * time.Minute
	windowAfter  = 2 * time.Hour
	restNeed     = "rest"
	restKind     = "quiet-room"
)

var (
	sessionQuestionWords = []string{"talk", "talks", "session", "sessions", "keynote", "keynotes", "workshop", "workshops", "panel", "panels", "speaker", "speakers", "lightning", "sign language", "asl", "captions", "captioned", "captioning", "interpreter", "interpreters"}
	restWords            = []string{"quiet", "overwhelmed", "sensory", "break", "rest"}
	pastWords            = []string{"earlier", "just got out", "missed", "miss", "this morning", "ended", "was on"}
)

// SessionsShown decides, before the store is asked, whether sessions belong in the answer and why.
// Rest comes first: a person asking for quiet is never handed talks.
// ShownIfTitleMatches means the sessions are fetched and kept only when the top title matches the question.
func SessionsShown(question, kind string, need *Need) string {
	q := wordsOf(question)
	switch {
	case containsAny(q, restWords) || kind == restKind || (need != nil && need.ID == restNeed):
		return SessionsHidden
	case containsAny(q, sessionQuestionWords):
		return ShownForSessionQuestion
	case kind != "" || need != nil:
		return SessionsHidden
	default:
		return ShownIfTitleMatches
	}
}

// TitleMatches reports whether the title shares at least two of the question's long words and at least half of them.
func TitleMatches(question, title string) bool {
	title = wordsOf(title)
	var long, shared int
	for _, w := range strings.Fields(wordsOf(question)) {
		if len(w) < 4 {
			continue
		}
		long++
		if strings.Contains(title, " "+w+" ") {
			shared++
		}
	}
	return shared >= 2 && shared*2 >= long
}

// Window is the time span sessions are searched in: running now or starting within two hours,
// widened back to the start of today when the question speaks of the past.
func Window(now time.Time, question string) (from, to time.Time) {
	from = now.Add(-windowBefore)
	if containsAny(wordsOf(question), pastWords) {
		local := now.In(mst)
		from = time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, mst)
	}
	return from, now.Add(windowAfter)
}

// Describe stamps each session with its state and start relative to now, and its distance tier from here.
func Describe(sessions []Session, now time.Time, here Here) []Session {
	out := make([]Session, 0, len(sessions))
	for _, s := range sessions {
		switch {
		case now.Before(s.Start):
			s.State = SessionUpcoming
			s.StartsInMinutes = int(math.Round(s.Start.Sub(now).Minutes()))
		case now.Before(s.End):
			s.State = SessionRunning
		default:
			s.State = SessionEnded
		}
		s.Distance = distanceTier(s.Venue, s.Level, here)
		out = append(out, s)
	}
	return out
}

func distanceTier(venue, level string, here Here) string {
	switch {
	case here.Venue == "":
		return DistanceUnknown
	case venue == here.Venue && level != "" && level == here.Level:
		return DistanceSameLevel
	case venue == here.Venue:
		return DistanceSameBuilding
	default:
		return DistanceOtherBuilding
	}
}

// InWindow keeps the sessions that have not ended by from and, unless to is zero, start by to, in the order given.
func InWindow(sessions []Session, from, to time.Time) []Session {
	out := make([]Session, 0, len(sessions))
	for _, s := range sessions {
		if s.End.Before(from) || (!to.IsZero() && s.Start.After(to)) {
			continue
		}
		out = append(out, s)
	}
	return out
}

// sessionsFor asks the store for the sessions the question is about, relevance first, then keeps the ones in the time window.
// When nothing relevant is on in the window and the question is not about the past, the next relevant sessions are kept instead,
// so "sign language" on Monday still answers Wednesday's crash course rather than five unrelated talks that happen to be on now.
func (s *Service) sessionsFor(ctx context.Context, ans *Answer, question string) error {
	reason := SessionsShown(question, KindOf(question), ans.Need)
	ans.SessionsShownBecause = reason
	ans.Sessions = []Session{}
	if reason == SessionsHidden {
		return nil
	}
	sessions, err := step(ctx, s.tracer(), "ask.FindSessions", func(ctx context.Context) ([]Session, error) { return s.Store.FindSessions(ctx, question) })
	if err != nil {
		return err
	}
	if reason == ShownIfTitleMatches {
		if len(sessions) == 0 || !TitleMatches(question, sessions[0].Title) {
			ans.SessionsShownBecause = SessionsHidden
			return nil
		}
		ans.SessionsShownBecause = ShownForTitleMatch
	}
	now := s.now()
	from, to := Window(now, question)
	kept := InWindow(sessions, from, to)
	if len(kept) == 0 && !from.Before(now.Add(-windowBefore)) {
		kept = InWindow(sessions, now, time.Time{})
	}
	ans.Sessions = Describe(kept, now, ans.Here)
	return nil
}

func containsAny(words string, phrases []string) bool {
	for _, p := range phrases {
		if strings.Contains(words, " "+p+" ") {
			return true
		}
	}
	return false
}
