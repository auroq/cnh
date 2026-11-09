package ingest

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/auroq/cnh/pkg/schedule"
	"github.com/auroq/cnh/pkg/venue"
)

type fakeSource struct {
	sessions  []schedule.Session
	err       error
	described bool
}

func (f *fakeSource) Index(context.Context) ([]schedule.Session, error) { return f.sessions, f.err }
func (f *fakeSource) Describe(_ context.Context, sessions []schedule.Session, report func(string, error)) int {
	f.described = true
	report(sessions[0].ID, errors.New("timeout"))
	return 1
}

type fakePlaces struct {
	places []venue.Place
	err    error
}

func (f fakePlaces) Places() ([]venue.Place, error) { return f.places, f.err }

type fakeIndexer struct {
	calls    []string
	sessions []schedule.Session
	places   []venue.Place
	failed   int
}

func (f *fakeIndexer) EnsureIndexes(context.Context) error {
	f.calls = append(f.calls, "ensure")
	return nil
}
func (f *fakeIndexer) IndexSessions(_ context.Context, s []schedule.Session) (Result, error) {
	f.calls = append(f.calls, "sessions")
	f.sessions = s
	return Result{Sent: len(s), Failed: f.failed, Count: len(s)}, nil
}
func (f *fakeIndexer) IndexPlaces(_ context.Context, p []venue.Place) (Result, error) {
	f.calls = append(f.calls, "places")
	f.places = p
	return Result{Sent: len(p), Count: len(p)}, nil
}

var quiet = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestRun(t *testing.T) {
	ctx := context.Background()
	sessions := []schedule.Session{
		{ID: "1", Title: "Keynote", Location: "Salt Palace | Level 1 | Hall 1"},
		{ID: "2", Title: "Signs", Location: "Salt Palace | Level 1 | Hall 1"},
		{ID: "3", Title: "Coat check", Location: "Salt Palace | Level 2 | East Registration"},
	}
	curated := []venue.Place{{ID: "quiet-260a", Kind: "quiet-room", Name: "260A", Venue: "Salt Palace", Source: "organiser"}}

	t.Run("when the schedule and the places both load", func(t *testing.T) {
		src := &fakeSource{sessions: sessions}
		idx := &fakeIndexer{}
		rep, err := (&Runner{Source: src, Places: fakePlaces{places: curated}, Indexer: idx, Log: quiet}).Run(ctx)
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should ensure the indexes before writing", func(t *testing.T) {
			if len(idx.calls) != 3 || idx.calls[0] != "ensure" || idx.calls[1] != "sessions" || idx.calls[2] != "places" {
				t.Fatalf("calls %v", idx.calls)
			}
		})
		t.Run("it should fetch the session details", func(t *testing.T) {
			if !src.described || rep.DescribeFailed != 1 {
				t.Fatalf("described=%v failed=%d", src.described, rep.DescribeFailed)
			}
		})
		t.Run("it should index every session", func(t *testing.T) {
			if len(idx.sessions) != 3 {
				t.Fatalf("indexed %d sessions", len(idx.sessions))
			}
		})
		t.Run("it should index the curated places plus one room per distinct location", func(t *testing.T) {
			if rep.Curated != 1 || rep.Rooms != 2 || len(idx.places) != 3 {
				t.Fatalf("curated=%d rooms=%d indexed=%d", rep.Curated, rep.Rooms, len(idx.places))
			}
		})
		t.Run("it should report success", func(t *testing.T) {
			if rep.Failed() {
				t.Fatalf("report %+v", rep)
			}
		})
	})

	t.Run("when details are skipped", func(t *testing.T) {
		src := &fakeSource{sessions: sessions}
		_, err := (&Runner{Source: src, Places: fakePlaces{}, Indexer: &fakeIndexer{}, SkipDetails: true, Log: quiet}).Run(ctx)
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should not fetch session pages", func(t *testing.T) {
			if src.described {
				t.Fatal("described")
			}
		})
	})

	t.Run("when the store rejects a document", func(t *testing.T) {
		rep, _ := (&Runner{Source: &fakeSource{sessions: sessions}, Places: fakePlaces{}, Indexer: &fakeIndexer{failed: 1}, SkipDetails: true, Log: quiet}).Run(ctx)
		t.Run("it should report the run as failed", func(t *testing.T) {
			if !rep.Failed() {
				t.Fatalf("report %+v", rep)
			}
		})
	})

	cases := []struct {
		name   string
		runner *Runner
		want   string
	}{
		{"the schedule cannot be fetched", &Runner{Source: &fakeSource{err: errors.New("503")}, Places: fakePlaces{}, Indexer: &fakeIndexer{}}, "fetch schedule"},
		{"the places file is broken", &Runner{Source: &fakeSource{sessions: sessions}, Places: fakePlaces{err: errors.New("yaml")}, Indexer: &fakeIndexer{}}, "load places"},
	}
	for _, c := range cases {
		t.Run("when "+c.name, func(t *testing.T) {
			c.runner.SkipDetails = true
			c.runner.Log = quiet
			_, err := c.runner.Run(ctx)
			t.Run("it should say which step failed", func(t *testing.T) {
				if err == nil || !contains(err.Error(), c.want) {
					t.Fatalf("got %v", err)
				}
			})
		})
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
