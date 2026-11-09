// Package ask answers "where is the nearest X from here" and free-text questions.
// It speaks in places, sessions and rooms; how they are searched is the Store's business.
package ask

import (
	"context"
	"fmt"
	"sort"
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
	// LocateRoom resolves what an attendee typed as their location to a known room, or returns an empty Here.
	LocateRoom(ctx context.Context, asked string) (Here, error)
	// FindPlaces returns access features matching the question, best first, never rooms.
	FindPlaces(ctx context.Context, question string) ([]Place, error)
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
type Service struct {
	Store  Store
	Needs  []Need
	Tracer trace.TracerProvider
	Meter  metric.MeterProvider

	once sync.Once
	inst *instruments
}

const maxPlaces = 5

// Ready reports whether the store answers; it records no span, because probes are not traffic.
func (s *Service) Ready(ctx context.Context) error {
	return s.Store.Ping(ctx)
}

// Locate resolves what the attendee typed as their location to a known room.
func (s *Service) Locate(ctx context.Context, asked string) (Here, error) {
	asked = strings.TrimSpace(asked)
	if asked == "" {
		return Here{}, nil
	}
	here, err := step(ctx, s.tracer(), "ask.LocateRoom", func(ctx context.Context) (Here, error) { return s.Store.LocateRoom(ctx, asked) })
	here.Asked = asked
	return here, err
}

// Ask finds places and sessions for the question and ranks places by nearness to here.
// It is also where the question is observed: one span with the question's shape and the answer's,
// and the metrics that say how often attendees are told something nobody has verified.
func (s *Service) Ask(ctx context.Context, question, from string) (Answer, error) {
	start := time.Now()
	question = strings.TrimSpace(question)
	ctx, span := s.tracer().Start(ctx, "ask.Ask", trace.WithAttributes(attrQuestionLength.Int(len(question))))
	defer span.End()

	ans, err := s.answer(ctx, question, from)
	kind, verified := topKind(ans)
	span.SetAttributes(
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

func (s *Service) answer(ctx context.Context, question, from string) (Answer, error) {
	here, err := s.Locate(ctx, from)
	if err != nil {
		return Answer{}, err
	}
	ans := Answer{Question: question, Here: here}
	if question == "" {
		ans.Summary = "Ask where something is, such as the nearest all-gender restroom or quiet room, or search for a session."
		return ans, nil
	}

	places, err := step(ctx, s.tracer(), "ask.FindPlaces", func(ctx context.Context) ([]Place, error) { return s.Store.FindPlaces(ctx, question) })
	if err != nil {
		return ans, err
	}
	ans.Places = RankByNearness(places, here)
	if len(ans.Places) > maxPlaces {
		ans.Places = ans.Places[:maxPlaces]
	}

	ans.Sessions, err = step(ctx, s.tracer(), "ask.FindSessions", func(ctx context.Context) ([]Session, error) { return s.Store.FindSessions(ctx, question) })
	if err != nil {
		return ans, err
	}
	ans.Summary = Summarise(ans)
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
func Summarise(a Answer) string {
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
	if p.Near != "" {
		where += ", " + strings.TrimRight(p.Near, ".")
	}
	sentence := fmt.Sprintf("%s: %s.", p.Name, where)
	switch {
	case a.Here.Venue == "" || a.Here.Level == "":
	case p.Venue == a.Here.Venue && p.Level == a.Here.Level:
		sentence += " Same level as you, no stairs or elevator."
	case p.Venue == a.Here.Venue:
		sentence += fmt.Sprintf(" Same building, you are on %s and it is on %s.", a.Here.Level, orUnknown(p.Level))
	default:
		sentence += fmt.Sprintf(" It is in the %s, you are in the %s.", p.Venue, a.Here.Venue)
	}
	if p.Verified {
		sentence += " Published by the organiser."
	} else {
		sentence += " Unverified: read off a plan, nobody has checked it yet."
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

func orUnknown(s string) string {
	if s == "" {
		return "an unlisted level"
	}
	return s
}
