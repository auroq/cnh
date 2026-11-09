package ask

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/auroq/cnh/pkg/model"
)

type fakeStatus struct {
	statuses map[string]model.AttendeeStatus
	err      error
	asked    []string
}

func (f *fakeStatus) StatusFor(_ context.Context, ids []string) (map[string]model.AttendeeStatus, error) {
	f.asked = ids
	return f.statuses, f.err
}

var reportedAt = time.Date(2026, 11, 9, 21, 52, 0, 0, time.UTC)

func TestAttendeeSentence(t *testing.T) {
	cases := []struct {
		name   string
		status *model.AttendeeStatus
		want   string
	}{
		{"nobody reported", nil, ""},
		{"one attendee reported a problem", &model.AttendeeStatus{Status: model.StatusProblem, Count: 1, LatestNote: "ramp is there but the door is locked", LatestAt: reportedAt}, "Reported 2:52 pm by one attendee: ramp is there but the door is locked."},
		{"two attendees reported problems", &model.AttendeeStatus{Status: model.StatusProblem, Count: 2, LatestNote: "out of order", LatestAt: reportedAt}, "Reported 2:52 pm by 2 attendees: out of order."},
		{"one attendee disputed the description", &model.AttendeeStatus{Status: model.StatusDisputed, Count: 1, LatestNote: "it is a freight elevator", LatestAt: reportedAt}, "Disputed by one attendee at 2:52 pm: it is a freight elevator."},
		{"two attendees agree it works", &model.AttendeeStatus{Status: model.StatusVerified, Count: 2, LatestAt: reportedAt}, "Verified by attendees (2)."},
		{"one attendee says it works", &model.AttendeeStatus{Status: model.StatusNone, Count: 1, LatestAt: reportedAt}, "One attendee says it works; a second confirmation makes it verified."},
	}
	for _, c := range cases {
		t.Run("when "+c.name, func(t *testing.T) {
			got := AttendeeSentence(Place{Attendee: c.status})
			t.Run("it should say so in one sentence", func(t *testing.T) {
				if got != c.want {
					t.Fatalf("got %q, want %q", got, c.want)
				}
			})
		})
	}
}

func TestAttendeeStatusInAnswers(t *testing.T) {
	restroom := Place{ID: "restroom-sp-l2-255", Kind: "all-gender-restroom", Name: "All-gender restroom outside 255", Venue: "Salt Palace", Level: "Level 2", Verified: true}
	elevator := Place{ID: "elevators-salt-palace", Kind: "elevator", Name: "Elevators", Venue: "Salt Palace", Verified: false}

	t.Run("when a place in the answer has a problem report", func(t *testing.T) {
		status := &fakeStatus{statuses: map[string]model.AttendeeStatus{
			"restroom-sp-l2-255": {Status: model.StatusProblem, Count: 1, LatestNote: "out of order", LatestAt: reportedAt},
		}}
		svc := &Service{Store: &fakeStore{places: []Place{restroom, elevator}}, Status: status}
		ans, err := svc.Ask(context.Background(), "nearest all-gender restroom", "")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should ask the status source about every place in the answer", func(t *testing.T) {
			if len(status.asked) != 2 || status.asked[0] != "restroom-sp-l2-255" || status.asked[1] != "elevators-salt-palace" {
				t.Fatalf("asked %v", status.asked)
			}
		})
		t.Run("it should fold the status into that place", func(t *testing.T) {
			if ans.Places[0].Attendee == nil || ans.Places[0].Attendee.Status != model.StatusProblem {
				t.Fatalf("place %+v", ans.Places[0])
			}
		})
		t.Run("it should leave a place nobody reported on without a status", func(t *testing.T) {
			if ans.Places[1].Attendee != nil {
				t.Fatalf("place %+v", ans.Places[1])
			}
		})
		t.Run("it should say the report right after saying the organiser published it", func(t *testing.T) {
			if !strings.Contains(ans.Summary, "Published by the organiser. Reported 2:52 pm by one attendee: out of order.") {
				t.Fatalf("summary %q", ans.Summary)
			}
		})
	})

	t.Run("when the status source fails", func(t *testing.T) {
		svc := &Service{Store: &fakeStore{places: []Place{restroom}}, Status: &fakeStatus{err: errors.New("down")}}
		ans, err := svc.Ask(context.Background(), "nearest all-gender restroom", "")
		t.Run("it should still answer", func(t *testing.T) {
			if err != nil || len(ans.Places) != 1 {
				t.Fatalf("answer %+v, %v", ans, err)
			}
		})
	})

	t.Run("when there is no status source", func(t *testing.T) {
		svc := &Service{Store: &fakeStore{places: []Place{restroom}}}
		ans, err := svc.Ask(context.Background(), "nearest all-gender restroom", "")
		t.Run("it should answer as before", func(t *testing.T) {
			if err != nil || ans.Places[0].Attendee != nil {
				t.Fatalf("answer %+v, %v", ans, err)
			}
		})
	})
}
