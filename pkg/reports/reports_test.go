package reports

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/auroq/cnh/pkg/model"
)

var now = time.Date(2026, 11, 9, 14, 52, 0, 0, time.UTC)

func at(minutesAgo int) time.Time { return now.Add(-time.Duration(minutesAgo) * time.Minute) }

func TestDerive(t *testing.T) {
	cases := []struct {
		name    string
		reports []model.Report
		want    model.AttendeeStatus
	}{
		{"no reports", nil, model.AttendeeStatus{Status: model.StatusNone}},
		{"one attendee says it works", []model.Report{
			{Verdict: model.VerdictWorks, Token: "a", At: at(5)},
		}, model.AttendeeStatus{Status: model.StatusNone, Count: 1, LatestAt: at(5)}},
		{"the same attendee says it works twice", []model.Report{
			{Verdict: model.VerdictWorks, Token: "a", At: at(5)},
			{Verdict: model.VerdictWorks, Token: "a", At: at(3)},
		}, model.AttendeeStatus{Status: model.StatusNone, Count: 1, LatestAt: at(3)}},
		{"two attendees agree it works", []model.Report{
			{Verdict: model.VerdictWorks, Token: "a", At: at(5)},
			{Verdict: model.VerdictWorks, Token: "b", At: at(3)},
		}, model.AttendeeStatus{Status: model.StatusVerified, Count: 2, LatestAt: at(3)}},
		{"one attendee reports a problem", []model.Report{
			{Verdict: model.VerdictProblem, Token: "a", At: at(1), Note: "ramp is there but the door is locked"},
		}, model.AttendeeStatus{Status: model.StatusProblem, Count: 1, LatestNote: "ramp is there but the door is locked", LatestAt: at(1)}},
		{"a problem outranks two agreeing attendees", []model.Report{
			{Verdict: model.VerdictWorks, Token: "a", At: at(50)},
			{Verdict: model.VerdictWorks, Token: "b", At: at(40)},
			{Verdict: model.VerdictProblem, Token: "c", At: at(1), Note: "out of order"},
		}, model.AttendeeStatus{Status: model.StatusProblem, Count: 1, LatestNote: "out of order", LatestAt: at(1)}},
		{"one attendee says it is not as described", []model.Report{
			{Verdict: model.VerdictNotAsDescribed, Token: "a", At: at(2), Note: "it is a freight elevator"},
		}, model.AttendeeStatus{Status: model.StatusDisputed, Count: 1, LatestNote: "it is a freight elevator", LatestAt: at(2)}},
		{"the latest note wins", []model.Report{
			{Verdict: model.VerdictProblem, Token: "a", At: at(30), Note: "out of order"},
			{Verdict: model.VerdictProblem, Token: "b", At: at(2), Note: "blocked by a booth"},
		}, model.AttendeeStatus{Status: model.StatusProblem, Count: 2, LatestNote: "blocked by a booth", LatestAt: at(2)}},
		{"a report older than six hours is forgotten", []model.Report{
			{Verdict: model.VerdictProblem, Token: "a", At: at(6*60 + 1), Note: "door locked at nine"},
		}, model.AttendeeStatus{Status: model.StatusNone}},
		{"a report exactly six hours old still counts", []model.Report{
			{Verdict: model.VerdictWorks, Token: "a", At: at(6 * 60)},
			{Verdict: model.VerdictWorks, Token: "b", At: at(1)},
		}, model.AttendeeStatus{Status: model.StatusVerified, Count: 2, LatestAt: at(1)}},
		{"a hidden report does not count", []model.Report{
			{Verdict: model.VerdictProblem, Token: "a", At: at(1), Note: "spam", Hidden: true},
			{Verdict: model.VerdictWorks, Token: "b", At: at(2)},
		}, model.AttendeeStatus{Status: model.StatusNone, Count: 1, LatestAt: at(2)}},
	}
	for _, c := range cases {
		t.Run("when "+c.name, func(t *testing.T) {
			got := Derive(c.reports, now)
			t.Run("it should derive "+c.want.Status, func(t *testing.T) {
				if got != c.want {
					t.Fatalf("got %+v, want %+v", got, c.want)
				}
			})
		})
	}
}

func TestLimiter(t *testing.T) {
	t.Run("when a key has made three requests in ten minutes", func(t *testing.T) {
		l := NewLimiter(3, 10*time.Minute)
		for i := range 3 {
			l.Allow("a", at(9-i))
		}
		t.Run("it should refuse the fourth", func(t *testing.T) {
			if l.Allow("a", now) {
				t.Fatal("allowed a fourth request")
			}
		})
		t.Run("it should still allow another key", func(t *testing.T) {
			if !l.Allow("b", now) {
				t.Fatal("refused a different key")
			}
		})
		t.Run("and the oldest has aged out", func(t *testing.T) {
			t.Run("it should allow again", func(t *testing.T) {
				if !l.Allow("a", at(9).Add(10*time.Minute+time.Second)) {
					t.Fatal("refused after the window passed")
				}
			})
		})
	})
}

type fakeStore struct {
	saved   []model.Report
	kinds   map[string]string
	reports []model.Report
	err     error
}

func (f *fakeStore) Save(_ context.Context, r model.Report) error {
	f.saved = append(f.saved, r)
	return f.err
}

func (f *fakeStore) Since(_ context.Context, placeIDs []string, _ time.Time) ([]model.Report, error) {
	var out []model.Report
	for _, r := range f.reports {
		for _, id := range placeIDs {
			if r.PlaceID == id {
				out = append(out, r)
			}
		}
	}
	return out, f.err
}

func (f *fakeStore) PlaceKind(_ context.Context, id string) (string, error) {
	return f.kinds[id], f.err
}

func newService(store *fakeStore) *Service {
	return &Service{Store: store, BlockedWords: []string{"casino", "kill yourself"}, Now: func() time.Time { return now }}
}

func works(token string) Submission {
	return Submission{PlaceID: "elevators-hyatt", Verdict: model.VerdictWorks, Token: token, IP: "10.0.0.1"}
}

func TestSubmit(t *testing.T) {
	ctx := context.Background()

	t.Run("when a problem report with a good note arrives", func(t *testing.T) {
		store := &fakeStore{kinds: map[string]string{"step-free-200-south": "step-free-entrance"}}
		svc := newService(store)
		got, err := svc.Submit(ctx, Submission{PlaceID: "step-free-200-south", Verdict: model.VerdictProblem, Note: "  ramp is there but the door is locked ", Token: "tok", IP: "10.0.0.1"})
		t.Run("it should accept it", func(t *testing.T) {
			if err != nil {
				t.Fatal(err)
			}
		})
		t.Run("it should store it with the note trimmed and the time from the clock", func(t *testing.T) {
			want := model.Report{ID: got.ID, PlaceID: "step-free-200-south", Verdict: model.VerdictProblem, Note: "ramp is there but the door is locked", At: now, Token: "tok", Source: "attendee"}
			if len(store.saved) != 1 || store.saved[0] != want {
				t.Fatalf("saved %+v, want %+v", store.saved, want)
			}
		})
		t.Run("it should give the report an id", func(t *testing.T) {
			if got.ID == "" {
				t.Fatal("empty id")
			}
		})
		t.Run("it should name the kind of the place for telemetry", func(t *testing.T) {
			if got.Kind != "step-free-entrance" {
				t.Fatalf("kind = %q", got.Kind)
			}
		})
	})

	rejected := []struct {
		name string
		sub  Submission
		want error
	}{
		{"the place is unknown", Submission{PlaceID: "narnia", Verdict: model.VerdictWorks, Token: "t", IP: "1"}, ErrUnknownPlace},
		{"the verdict is not one of the three", Submission{PlaceID: "elevators-hyatt", Verdict: "meh", Token: "t", IP: "1"}, ErrVerdict},
		{"a problem has no note", Submission{PlaceID: "elevators-hyatt", Verdict: model.VerdictProblem, Token: "t", IP: "1"}, ErrNoteRequired},
		{"not as described has no note", Submission{PlaceID: "elevators-hyatt", Verdict: model.VerdictNotAsDescribed, Token: "t", IP: "1"}, ErrNoteRequired},
		{"the note is five characters", Submission{PlaceID: "elevators-hyatt", Verdict: model.VerdictProblem, Note: "stuck", Token: "t", IP: "1"}, ErrNoteLength},
		{"the note is 281 characters", Submission{PlaceID: "elevators-hyatt", Verdict: model.VerdictProblem, Note: strings.Repeat("x", 281), Token: "t", IP: "1"}, ErrNoteLength},
		{"the note has a URL", Submission{PlaceID: "elevators-hyatt", Verdict: model.VerdictProblem, Note: "see https://example.com for details", Token: "t", IP: "1"}, ErrNoteURL},
		{"the note has a www link", Submission{PlaceID: "elevators-hyatt", Verdict: model.VerdictProblem, Note: "look at www.example.com now", Token: "t", IP: "1"}, ErrNoteURL},
		{"the note has a blocked word in any case", Submission{PlaceID: "elevators-hyatt", Verdict: model.VerdictProblem, Note: "best CASINO in town here", Token: "t", IP: "1"}, ErrNoteBlocked},
		{"the note has a blocked phrase", Submission{PlaceID: "elevators-hyatt", Verdict: model.VerdictProblem, Note: "go kill yourself please", Token: "t", IP: "1"}, ErrNoteBlocked},
		{"the honeypot is filled", Submission{PlaceID: "elevators-hyatt", Verdict: model.VerdictWorks, Honeypot: "bot", Token: "t", IP: "1"}, ErrHoneypot},
		{"the token is missing", Submission{PlaceID: "elevators-hyatt", Verdict: model.VerdictWorks, IP: "1"}, ErrToken},
	}
	for _, c := range rejected {
		t.Run("when "+c.name, func(t *testing.T) {
			store := &fakeStore{kinds: map[string]string{"elevators-hyatt": "elevator"}}
			svc := newService(store)
			_, err := svc.Submit(ctx, c.sub)
			t.Run("it should reject it as "+c.want.Error(), func(t *testing.T) {
				if !errors.Is(err, c.want) {
					t.Fatalf("err = %v", err)
				}
			})
			t.Run("it should store nothing", func(t *testing.T) {
				if len(store.saved) != 0 {
					t.Fatalf("saved %+v", store.saved)
				}
			})
		})
	}

	t.Run("when a works report has an optional note", func(t *testing.T) {
		store := &fakeStore{kinds: map[string]string{"elevators-hyatt": "elevator"}}
		s := works("t")
		s.Note = "both cars running"
		_, err := newService(store).Submit(ctx, s)
		t.Run("it should accept it", func(t *testing.T) {
			if err != nil {
				t.Fatal(err)
			}
		})
	})

	t.Run("when one token has reported three times in ten minutes", func(t *testing.T) {
		store := &fakeStore{kinds: map[string]string{"elevators-hyatt": "elevator"}}
		svc := newService(store)
		for range 3 {
			if _, err := svc.Submit(ctx, works("same")); err != nil {
				t.Fatal(err)
			}
		}
		_, err := svc.Submit(ctx, works("same"))
		t.Run("it should reject the fourth as rate limited", func(t *testing.T) {
			if !errors.Is(err, ErrRateLimited) {
				t.Fatalf("err = %v", err)
			}
		})
		t.Run("it should store only three", func(t *testing.T) {
			if len(store.saved) != 3 {
				t.Fatalf("saved %d", len(store.saved))
			}
		})
	})

	t.Run("when one IP has reported twenty times in ten minutes", func(t *testing.T) {
		store := &fakeStore{kinds: map[string]string{"elevators-hyatt": "elevator"}}
		svc := newService(store)
		for i := range 20 {
			if _, err := svc.Submit(ctx, works(fmt.Sprint("token", i))); err != nil {
				t.Fatal(err)
			}
		}
		_, err := svc.Submit(ctx, works("token-new"))
		t.Run("it should reject the twenty-first even from a new token", func(t *testing.T) {
			if !errors.Is(err, ErrRateLimited) {
				t.Fatalf("err = %v", err)
			}
		})
	})

	t.Run("when the store fails", func(t *testing.T) {
		store := &fakeStore{kinds: map[string]string{"elevators-hyatt": "elevator"}, err: errors.New("down")}
		_, err := newService(store).Submit(ctx, works("t"))
		t.Run("it should return the error with context", func(t *testing.T) {
			if err == nil || !strings.Contains(err.Error(), "down") {
				t.Fatalf("err = %v", err)
			}
		})
	})
}

func TestStatusFor(t *testing.T) {
	ctx := context.Background()
	store := &fakeStore{reports: []model.Report{
		{PlaceID: "elevators-hyatt", Verdict: model.VerdictWorks, Token: "a", At: at(5)},
		{PlaceID: "elevators-hyatt", Verdict: model.VerdictWorks, Token: "b", At: at(3)},
		{PlaceID: "step-free-200-south", Verdict: model.VerdictProblem, Token: "c", At: at(1), Note: "door locked"},
		{PlaceID: "quiet-260a", Verdict: model.VerdictWorks, Token: "d", At: at(1)},
	}}
	svc := newService(store)

	t.Run("when status is asked for several places", func(t *testing.T) {
		got, err := svc.StatusFor(ctx, []string{"elevators-hyatt", "step-free-200-south", "quiet-260a", "nothing-reported"})
		if err != nil {
			t.Fatal(err)
		}
		cases := map[string]model.AttendeeStatus{
			"elevators-hyatt":     {Status: model.StatusVerified, Count: 2, LatestAt: at(3)},
			"step-free-200-south": {Status: model.StatusProblem, Count: 1, LatestNote: "door locked", LatestAt: at(1)},
			"quiet-260a":          {Status: model.StatusNone, Count: 1, LatestAt: at(1)},
		}
		for id, want := range cases {
			t.Run("it should derive "+want.Status+" for "+id, func(t *testing.T) {
				if got[id] != want {
					t.Fatalf("got %+v, want %+v", got[id], want)
				}
			})
		}
		t.Run("it should leave a place nobody reported on out of the map", func(t *testing.T) {
			if _, ok := got["nothing-reported"]; ok {
				t.Fatalf("got %+v", got["nothing-reported"])
			}
		})
	})

	t.Run("when no places are asked for", func(t *testing.T) {
		got, err := svc.StatusFor(ctx, nil)
		t.Run("it should return an empty map without asking the store", func(t *testing.T) {
			if err != nil || len(got) != 0 {
				t.Fatalf("got %+v, %v", got, err)
			}
		})
	})
}
