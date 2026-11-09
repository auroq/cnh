package ask

import (
	"context"
	"fmt"

	"github.com/auroq/cnh/pkg/model"
)

// Status is where the attendee reports on places come from; nil means nobody can report yet.
type Status interface {
	StatusFor(ctx context.Context, placeIDs []string) (map[string]model.AttendeeStatus, error)
}

// attach folds what attendees reported into each place; a failing status source never fails the answer.
func (s *Service) attach(ctx context.Context, places []Place) {
	if s.Status == nil || len(places) == 0 {
		return
	}
	ids := make([]string, 0, len(places))
	for _, p := range places {
		ids = append(ids, p.ID)
	}
	statuses, err := step(ctx, s.tracer(), "ask.AttendeeStatus", func(ctx context.Context) (map[string]model.AttendeeStatus, error) {
		return s.Status.StatusFor(ctx, ids)
	})
	if err != nil {
		return
	}
	for i := range places {
		if st, ok := statuses[places[i].ID]; ok {
			places[i].Attendee = &st
		}
	}
}

// AttendeeSentence is what a screen reader hears about the attendee reports on a place, in the same breath as verified or unverified.
func AttendeeSentence(p Place) string {
	st := p.Attendee
	if st == nil || st.Count == 0 {
		return ""
	}
	when := st.LatestAt.In(mst).Format("3:04 pm")
	switch st.Status {
	case model.StatusProblem:
		return fmt.Sprintf("Reported %s by %s: %s.", when, attendees(st.Count), st.LatestNote)
	case model.StatusDisputed:
		return fmt.Sprintf("Disputed by %s at %s: %s.", attendees(st.Count), when, st.LatestNote)
	case model.StatusVerified:
		return fmt.Sprintf("Verified by attendees (%d).", st.Count)
	default:
		return "One attendee says it works; a second confirmation makes it verified."
	}
}

func attendees(n int) string {
	if n == 1 {
		return "one attendee"
	}
	return fmt.Sprintf("%d attendees", n)
}
