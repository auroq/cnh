package ask

import (
	"context"
	"errors"
	"testing"
	"time"
)

type scheduledStore struct {
	*fakeStore
	byTitle map[string]Session
	byID    map[string]Session
	matched []string
	wanted  []string
}

func (f *scheduledStore) MatchSession(_ context.Context, title string) (Session, bool, error) {
	f.matched = append(f.matched, title)
	s, ok := f.byTitle[title]
	return s, ok, f.err
}

func (f *scheduledStore) SessionsByID(_ context.Context, ids []string) ([]Session, error) {
	f.wanted = append(f.wanted, ids...)
	var out []Session
	for _, id := range ids {
		if s, ok := f.byID[id]; ok {
			out = append(out, s)
		}
	}
	return out, f.err
}

func TestSchedule(t *testing.T) {
	ctx := context.Background()
	day := time.Date(2026, 11, 11, 0, 0, 0, 0, time.UTC)
	at := func(h, m int) time.Time { return day.Add(time.Duration(h)*time.Hour + time.Duration(m)*time.Minute) }
	crash := Session{ID: "crash", Title: "Sign Language Crash Course", Start: at(16, 0), End: at(16, 35), Room: "Grand Ballroom J", Venue: "Salt Palace", Level: "Level 1"}
	keynote := Session{ID: "keynote", Title: "Keynote: Welcome", Start: at(17, 0), End: at(17, 35), Room: "Grand Ballroom A", Venue: "Salt Palace", Level: "Level 1"}
	quiet := Place{ID: "quiet-room-260a", Kind: "quiet-room", Name: "All-attendee quiet room", Venue: "Salt Palace", Level: "Level 2", Verified: true}
	newStore := func() *scheduledStore {
		return &scheduledStore{
			fakeStore: &fakeStore{room: Here{Room: "Grand Ballroom J", Venue: "Salt Palace", Level: "Level 1"}, places: []Place{quiet}},
			byTitle:   map[string]Session{"Sign Language Crash Course": crash, "Keynote: Welcome": keynote},
			byID:      map[string]Session{"crash": crash, "keynote": keynote},
		}
	}

	t.Run("when three titles are pasted and one is made up", func(t *testing.T) {
		store := newStore()
		matches, err := (&Service{Store: store}).MatchSessions(ctx, []string{"Sign Language Crash Course", "A talk nobody gave", "Keynote: Welcome"})
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should ask the store about every title in order", func(t *testing.T) {
			if len(store.matched) != 3 || store.matched[1] != "A talk nobody gave" {
				t.Fatalf("matched %q", store.matched)
			}
		})
		t.Run("it should pair the real titles with their sessions", func(t *testing.T) {
			if matches[0].Session == nil || matches[0].Session.ID != "crash" || matches[2].Session == nil || matches[2].Session.ID != "keynote" {
				t.Fatalf("matches %+v", matches)
			}
		})
		t.Run("it should keep the made-up title with no session", func(t *testing.T) {
			if matches[1].Title != "A talk nobody gave" || matches[1].Session != nil {
				t.Fatalf("matches %+v", matches)
			}
		})
	})

	t.Run("when the store cannot look up sessions", func(t *testing.T) {
		_, err := (&Service{Store: &fakeStore{}}).MatchSessions(ctx, []string{"Keynote: Welcome"})
		t.Run("it should say so", func(t *testing.T) {
			if !errors.Is(err, ErrNoSessionFinder) {
				t.Fatalf("err %v", err)
			}
		})
	})

	t.Run("when a question is asked ten minutes after a kept session ended and no location is given", func(t *testing.T) {
		store := newStore()
		svc := &Service{Store: store, Now: func() time.Time { return at(16, 45) }}
		ans, err := svc.AskScheduled(ctx, "nearest quiet room", "", nil, []string{"crash", "keynote"})
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should load exactly the kept sessions", func(t *testing.T) {
			if len(store.wanted) != 2 || store.wanted[0] != "crash" {
				t.Fatalf("wanted %q", store.wanted)
			}
		})
		t.Run("it should assume the attendee is still near that room", func(t *testing.T) {
			if ans.Schedule == nil || ans.Schedule.AssumedLocation != "Grand Ballroom J" {
				t.Fatalf("schedule %+v", ans.Schedule)
			}
		})
		t.Run("it should resolve that room as here", func(t *testing.T) {
			if ans.Here.Room != "Grand Ballroom J" || ans.Here.Venue != "Salt Palace" {
				t.Fatalf("here %+v", ans.Here)
			}
		})
		t.Run("it should say the location came from the schedule", func(t *testing.T) {
			if ans.Here.Via != ViaSchedule {
				t.Fatalf("via %q", ans.Here.Via)
			}
		})
		t.Run("it should name the next session", func(t *testing.T) {
			if ans.Schedule.Next == nil || ans.Schedule.Next.ID != "keynote" {
				t.Fatalf("schedule %+v", ans.Schedule)
			}
		})
		t.Run("it should name the session just left", func(t *testing.T) {
			if ans.Schedule.JustLeft == nil || ans.Schedule.JustLeft.ID != "crash" {
				t.Fatalf("schedule %+v", ans.Schedule)
			}
		})
	})

	t.Run("when a location is given", func(t *testing.T) {
		store := newStore()
		store.room = Here{Room: "255 BC", Venue: "Salt Palace", Level: "Level 2"}
		svc := &Service{Store: store, Now: func() time.Time { return at(16, 45) }}
		ans, err := svc.AskScheduled(ctx, "nearest quiet room", "255 BC", nil, []string{"crash"})
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should not assume anything", func(t *testing.T) {
			if ans.Schedule.AssumedLocation != "" || ans.Here.Via != ViaRoom {
				t.Fatalf("schedule %+v here %+v", ans.Schedule, ans.Here)
			}
		})
	})

	t.Run("when no session has ended recently", func(t *testing.T) {
		store := newStore()
		svc := &Service{Store: store, Now: func() time.Time { return at(12, 0) }}
		ans, err := svc.AskScheduled(ctx, "nearest quiet room", "", nil, []string{"crash", "keynote"})
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should leave the location unknown", func(t *testing.T) {
			if ans.Here.Venue != "" || ans.Schedule.AssumedLocation != "" {
				t.Fatalf("here %+v schedule %+v", ans.Here, ans.Schedule)
			}
		})
	})

	t.Run("when no schedule is kept", func(t *testing.T) {
		store := newStore()
		ans, err := (&Service{Store: store}).AskScheduled(ctx, "nearest quiet room", "", nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should answer exactly as Ask does", func(t *testing.T) {
			if ans.Schedule != nil || len(store.wanted) != 0 {
				t.Fatalf("schedule %+v wanted %q", ans.Schedule, store.wanted)
			}
		})
	})

	t.Run("when the agenda alone is asked for three sessions in a row", func(t *testing.T) {
		store := newStore()
		third := Session{ID: "third", Start: at(17, 40), End: at(18, 15), Room: "Grand Ballroom B"}
		store.byID["third"] = third
		store.byID["crash"] = Session{ID: "crash", Start: at(16, 20), End: at(16, 55), Room: "Grand Ballroom J"}
		sched, err := (&Service{Store: store, Now: func() time.Time { return at(12, 0) }}).Agenda(ctx, []string{"crash", "keynote", "third"})
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should flag them as back to back", func(t *testing.T) {
			if !sched.BackToBack {
				t.Fatalf("schedule %+v", sched)
			}
		})
	})
}
