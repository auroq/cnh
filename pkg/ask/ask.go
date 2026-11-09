// Package ask answers "where is the nearest X from here" and free-text questions.
// It speaks in places, sessions and rooms; how they are searched is the Store's business.
package ask

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/auroq/cnh/pkg/model"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Store is what ask needs from a search backend, in the app's own terms.
type Store interface {
	// Ping reports whether the backend answers at all; it is the readiness probe's question and is never traced.
	Ping(ctx context.Context) error
	// LocateRoom resolves a room named exactly as the schedule names it, or returns an empty Here.
	LocateRoom(ctx context.Context, name string) (Here, error)
	// LocateRoomFuzzy resolves a misspelt or partial room name, or returns an empty Here.
	LocateRoomFuzzy(ctx context.Context, asked string) (Here, error)
	// LocateAlias resolves a phrase people use for a place ("where I registered", "the 100 S side"), or returns an empty Here.
	LocateAlias(ctx context.Context, phrase string) (Here, error)
	// LocateSession resolves a session title to the room it is in, preferring one that ended just before now, or returns an empty Here.
	LocateSession(ctx context.Context, title string, now time.Time) (Here, error)
	// LocateNearby resolves a free description to the place whose content it most resembles, or returns an empty Here when nothing is close.
	LocateNearby(ctx context.Context, phrase string) (Here, error)
	// FindPlaces returns access features matching the question, best first, never rooms;
	// when kind is not empty, only places of that kind.
	FindPlaces(ctx context.Context, question, kind string) ([]Place, error)
	// FindSessions returns schedule sessions matching the question, best first.
	FindSessions(ctx context.Context, question string) ([]Session, error)
	// FindNeeds returns the needs whose phrases the question means, best first, with the score the use case thresholds on.
	FindNeeds(ctx context.Context, question string) ([]NeedMatch, error)
}

// The domain types live in pkg/model; these names stay so callers read naturally.
type (
	Place   = model.Place
	Need    = model.Need
	Session = model.Session
	Here    = model.Here
	Answer  = model.Answer
)

// Service answers questions. Needs is the vocabulary that turns what people say into kinds of place.
// Now is the clock sessions and doors are timed against; nil means time.Now, tests fix it.
type Service struct {
	Store  Store
	Needs  []Need
	Status Status
	Router Router
	Tracer trace.TracerProvider
	Meter  metric.MeterProvider
	Now    func() time.Time

	once sync.Once
	inst *instruments
}

func (s *Service) now() time.Time {
	if s.Now == nil {
		return time.Now()
	}
	return s.Now()
}

const maxPlaces = 5

// Ready reports whether the store answers; it records no span, because probes are not traffic.
func (s *Service) Ready(ctx context.Context) error {
	return s.Store.Ping(ctx)
}

// Locate resolves what the attendee said about where they are to a known place, trying each rule in order.
func (s *Service) Locate(ctx context.Context, asked string) (Here, error) {
	asked = strings.TrimSpace(asked)
	if asked == "" {
		return Here{}, nil
	}
	for _, r := range s.rules(asked) {
		here, err := step(ctx, s.tracer(), "ask."+r.span, r.find)
		if err != nil {
			return Here{Asked: asked}, err
		}
		if here.Venue != "" {
			here.Asked, here.Via = asked, r.via
			return here, nil
		}
	}
	return Here{Asked: asked}, nil
}

// Ask finds places and sessions for the question and ranks places by nearness to here.
// It is also where the question is observed: one span with the question's shape and the answer's,
// and the metrics that say how often attendees are told something nobody has verified.
func (s *Service) Ask(ctx context.Context, question, from string) (Answer, error) {
	return s.ask(ctx, question, from, nil)
}

func (s *Service) ask(ctx context.Context, question, from string, prefs Preferences) (Answer, error) {
	start := time.Now()
	question = strings.TrimSpace(question)
	ctx, span := s.tracer().Start(ctx, "ask.Ask", trace.WithAttributes(attrQuestionLength.Int(len(question))))
	defer span.End()

	ans, err := s.answer(ctx, question, from, prefs)
	kind, verified := topKind(ans)
	span.SetAttributes(
		attrNeed.String(needID(ans)),
		attrHereVenue.String(ans.Here.Venue),
		attrHereLevel.String(ans.Here.Level),
		attrPlacesCount.Int(len(ans.Places)),
		attrSessionsCount.Int(len(ans.Sessions)),
		attrAnswerVerified.Bool(verified),
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	if question != "" {
		in := s.instruments()
		elapsed := float64(time.Since(start)) / float64(time.Millisecond)
		in.duration.Record(ctx, elapsed, metric.WithAttributes(attrHasLocation.Bool(ans.Here.Venue != ""), attrAnswered.Bool(err == nil && kind != "none")))
		if err == nil {
			in.answers.Add(ctx, 1, metric.WithAttributes(attrKind.String(kind), attrVerified.Bool(verified)))
			if len(ans.Places) > 0 && !verified {
				in.unverifiedServed.Add(ctx, 1, metric.WithAttributes(attrKind.String(kind)))
			}
		}
	}
	return ans, err
}

func (s *Service) answer(ctx context.Context, question, from string, prefs Preferences) (Answer, error) {
	here, err := s.Locate(ctx, from)
	if err != nil {
		return Answer{}, err
	}
	ans := Answer{Question: question, Here: here}
	if question == "" {
		ans.Summary = "Ask where something is, such as the nearest all-gender restroom or quiet room, or search for a session."
		return ans, nil
	}

	places, err := s.placesFor(ctx, &ans, question, prefs)
	if err != nil {
		return ans, err
	}
	ans.Places = RankByNearness(places, here)
	applyMusts(&ans, question, prefs)
	if len(ans.Places) > maxPlaces {
		ans.Places = ans.Places[:maxPlaces]
	}
	s.attach(ctx, ans.Places)
	s.routeFor(&ans, prefs)

	if err := s.sessionsFor(ctx, &ans, question); err != nil {
		return ans, err
	}
	ans.Summary = Summarise(ans)
	shapeSummary(&ans, prefs)
	ans.Learned = Learned(question, ans.Need)
	return ans, nil
}

// RankByNearness keeps the search order but moves places on the attendee's level, then venue, to the front.
// Without coordinates this is the honest ordering: same level beats same building beats a different building.
func RankByNearness(places []Place, here Here) []Place {
	out := append([]Place(nil), places...)
	if here.Venue == "" {
		return out
	}
	tier := func(p Place) int {
		switch {
		case here.Side != "" && p.Side == here.Side:
			return -1
		case p.Venue == here.Venue && p.Level != "" && p.Level == here.Level:
			return 0
		case p.Venue == here.Venue:
			return 1
		default:
			return 2
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return tier(out[i]) < tier(out[j]) })
	return out
}

// Summarise writes the one-sentence answer a screen reader hears first.
// When the question was understood as a need, it names the need before the place and ends with the need's note.
func Summarise(a Answer) string {
	sentence := summarise(a)
	if a.Need == nil {
		return sentence
	}
	sentence = "You asked for " + a.Need.Name + ". " + sentence
	if a.Need.Note != "" && len(a.Places) > 0 {
		sentence += " " + a.Need.Note
	}
	return sentence
}

func summarise(a Answer) string {
	if len(a.Places) == 0 && len(a.Sessions) == 0 {
		return "Nothing matched. Try other words, or ask staff at registration."
	}
	if len(a.Places) == 0 {
		return fmt.Sprintf("No place matched, but %d sessions did.", len(a.Sessions))
	}
	if len(a.Sessions) > 0 && (looksLikeSessionQuestion(a.Question) || len(a.Places) == 0) {
		s := a.Sessions[0]
		return fmt.Sprintf("%d sessions match. First: %s, %s, %s.", len(a.Sessions), s.Title, s.Start.In(mst).Format("Monday 3:04 pm"), strings.ReplaceAll(s.Location, " | ", ", "))
	}
	p := a.Places[0]
	where := p.Venue
	if p.Level != "" {
		where += ", " + p.Level
	}
	if p.Side != "" {
		where += ", " + p.Side + " side"
	}
	if p.Near != "" {
		where += ", " + strings.TrimRight(p.Near, ".")
	}
	sentence := fmt.Sprintf("%s: %s.", p.Name, where)
	switch {
	case a.Here.Side != "" && p.Side != "" && p.Side == a.Here.Side:
		sentence += " Same side of the building as you."
	case a.Here.Side != "" && p.Side != "":
		sentence += fmt.Sprintf(" You are on the %s side, this door is on the %s side.", a.Here.Side, p.Side)
	case a.Here.Venue == "" || a.Here.Level == "":
	case p.Venue == a.Here.Venue && p.Level == a.Here.Level:
		sentence += " Same level as you, no stairs or elevator."
	case p.Venue == a.Here.Venue:
		sentence += " " + levelChange(a.Here.Level, p.Level)
	default:
		sentence += fmt.Sprintf(" It is in the %s, you are in the %s.", p.Venue, a.Here.Venue)
	}
	for _, fact := range p.Route {
		sentence += " " + fact
	}
	if p.Verified {
		sentence += " Published by the organiser."
	} else {
		sentence += " Unverified: read off a plan, nobody has checked it yet."
	}
	if attendee := AttendeeSentence(p); attendee != "" {
		sentence += " " + attendee
	}
	return sentence
}

var mst = time.FixedZone("MST", -7*3600)

var sessionWords = []string{"talk", "session", "keynote", "speaker", "workshop", "panel", "lightning", "about", "presentation"}

func looksLikeSessionQuestion(q string) bool {
	q = strings.ToLower(q)
	for _, w := range sessionWords {
		if strings.Contains(q, w) {
			return true
		}
	}
	return false
}

// levelChange is the route fact the app can state from the data alone: how many levels, which way, and that it means an elevator or stairs.
// It never estimates a time, because a time assumes a pace and a body (ADR 0012).
func levelChange(from, to string) string {
	if to == "" {
		return "Same building, we do not know which level it is on."
	}
	f, okFrom := levelNumber(from)
	t, okTo := levelNumber(to)
	if !okFrom || !okTo {
		return fmt.Sprintf("Same building, you are on %s and it is on %s.", from, to)
	}
	n, direction := t-f, "up"
	if n < 0 {
		n, direction = -n, "down"
	}
	return fmt.Sprintf("Same building, %s %s, by elevator or stairs.", levels(n), direction)
}

func levelNumber(level string) (int, bool) {
	n, err := strconv.Atoi(strings.TrimPrefix(level, "Level "))
	return n, err == nil
}

func levels(n int) string {
	words := []string{"", "one level", "two levels", "three levels", "four levels", "five levels"}
	if n < len(words) {
		return words[n]
	}
	return fmt.Sprintf("%d levels", n)
}
