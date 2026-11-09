package ask

import (
	"context"
	"errors"

	"github.com/auroq/cnh/pkg/agenda"
	"github.com/auroq/cnh/pkg/model"
)

// SessionFinder is the optional half of a Store that matches pasted titles and loads the sessions an attendee kept.
// A Store without it still answers questions; the schedule is then not offered.
type SessionFinder interface {
	MatchSession(ctx context.Context, title string) (Session, bool, error)
	SessionsByID(ctx context.Context, ids []string) ([]Session, error)
}

type (
	Match    = model.Match
	Schedule = model.Schedule
)

// ViaSchedule says the location was not given but assumed from the session the attendee just left.
const ViaSchedule = "schedule"

// ErrNoSessionFinder is returned when the store cannot look sessions up by title or id.
var ErrNoSessionFinder = errors.New("the store cannot look up sessions")

// MatchSessions pairs each pasted title with the session it names, keeping the titles that matched nothing.
func (s *Service) MatchSessions(ctx context.Context, titles []string) ([]Match, error) {
	finder, ok := s.Store.(SessionFinder)
	if !ok {
		return nil, ErrNoSessionFinder
	}
	out := make([]Match, 0, len(titles))
	for _, title := range titles {
		session, err := step(ctx, s.tracer(), "ask.MatchSession", func(ctx context.Context) (*Session, error) {
			session, found, err := finder.MatchSession(ctx, title)
			if err != nil || !found {
				return nil, err
			}
			return &session, nil
		})
		if err != nil {
			return out, err
		}
		out = append(out, Match{Title: title, Session: session})
	}
	return out, nil
}

// Agenda says what the kept sessions mean right now: the next one, the one just left, and whether three run back to back.
func (s *Service) Agenda(ctx context.Context, ids []string) (Schedule, error) {
	sched, _, err := s.agenda(ctx, ids)
	return sched, err
}

func (s *Service) agenda(ctx context.Context, ids []string) (Schedule, []Session, error) {
	if len(ids) == 0 {
		return Schedule{}, nil, nil
	}
	finder, ok := s.Store.(SessionFinder)
	if !ok {
		return Schedule{}, nil, ErrNoSessionFinder
	}
	sessions, err := step(ctx, s.tracer(), "ask.SessionsByID", func(ctx context.Context) ([]Session, error) { return finder.SessionsByID(ctx, ids) })
	if err != nil {
		return Schedule{}, nil, err
	}
	now := s.now()
	return Schedule{
		Sessions:   agenda.ByStart(sessions),
		Next:       agenda.Next(sessions, now),
		JustLeft:   agenda.JustLeft(sessions, now),
		BackToBack: agenda.BackToBack(sessions),
	}, sessions, nil
}

// AskScheduled is AskFor with the attendee's schedule: when no location is given and a session of theirs ended
// in the last half hour, its room is assumed to be where they are, and the answer says so.
func (s *Service) AskScheduled(ctx context.Context, question, from string, prefs Preferences, ids []string) (Answer, error) {
	sched, _, err := s.agenda(ctx, ids)
	if err != nil {
		return Answer{}, err
	}
	if from == "" && sched.JustLeft != nil && sched.JustLeft.Room != "" {
		from = sched.JustLeft.Room
		sched.AssumedLocation = from
	}
	ans, err := s.ask(ctx, question, from, prefs)
	if sched.AssumedLocation != "" && ans.Here.Venue != "" {
		ans.Here.Via = ViaSchedule
	}
	if len(ids) > 0 {
		ans.Schedule = &sched
	}
	return ans, err
}

// ParseTitles reads what the attendee pasted, an iCal export or titles one per line, into distinct titles.
func ParseTitles(text string) []string { return agenda.Parse(text) }

// ParseScheduleIDs reads the sched parameter, and JoinScheduleIDs writes it.
func ParseScheduleIDs(s string) []string { return agenda.ParseIDs(s) }

func JoinScheduleIDs(ids []string) string { return agenda.JoinIDs(ids) }
