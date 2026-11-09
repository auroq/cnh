package ask

import (
	"context"
	"errors"
	"testing"
)

type fakeStore struct {
	room     Here
	places   []Place
	sessions []Session
	err      error
	located  []string
	searched []string
}

func (f *fakeStore) Ping(context.Context) error { return f.err }

func (f *fakeStore) LocateRoom(_ context.Context, asked string) (Here, error) {
	f.located = append(f.located, asked)
	return f.room, f.err
}

func (f *fakeStore) FindPlaces(_ context.Context, question string) ([]Place, error) {
	f.searched = append(f.searched, question)
	return f.places, f.err
}

func (f *fakeStore) FindSessions(_ context.Context, question string) ([]Session, error) {
	return f.sessions, f.err
}

func TestAsk(t *testing.T) {
	ctx := context.Background()
	hyattRestroom := Place{ID: "hyatt-l1-restroom", Kind: "all-gender-restroom", Name: "All-gender restroom, Hyatt lobby", Venue: "Hyatt Regency", Level: "Level 1", Verified: true}
	outside255 := Place{ID: "sp-l2-restroom", Kind: "all-gender-restroom", Name: "All-gender restroom outside 255", Venue: "Salt Palace", Level: "Level 2", Near: "Outside room 255", Verified: true}

	t.Run("when the attendee in 255 BC asks for the nearest all-gender restroom", func(t *testing.T) {
		store := &fakeStore{
			room:   Here{Room: "255 BC", Venue: "Salt Palace", Level: "Level 2"},
			places: []Place{hyattRestroom, outside255},
		}
		ans, err := (&Service{Store: store}).Ask(ctx, " nearest all-gender restroom ", " 255 BC ")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should resolve the trimmed room name through the store", func(t *testing.T) {
			if len(store.located) != 1 || store.located[0] != "255 BC" {
				t.Fatalf("located %q", store.located)
			}
		})
		t.Run("it should search places with the trimmed question", func(t *testing.T) {
			if len(store.searched) != 1 || store.searched[0] != "nearest all-gender restroom" {
				t.Fatalf("searched %q", store.searched)
			}
		})
		t.Run("it should put the restroom on the attendee's level first", func(t *testing.T) {
			if ans.Places[0].ID != outside255.ID {
				t.Fatalf("first place %s", ans.Places[0].ID)
			}
		})
		t.Run("it should say the answer is on the same level", func(t *testing.T) {
			want := "All-gender restroom outside 255: Salt Palace, Level 2, Outside room 255. Same level as you, no stairs or elevator. Published by the organiser."
			if ans.Summary != want {
				t.Fatalf("\n got %q\nwant %q", ans.Summary, want)
			}
		})
	})

	t.Run("when no room is given", func(t *testing.T) {
		store := &fakeStore{places: []Place{hyattRestroom, outside255}}
		ans, err := (&Service{Store: store}).Ask(ctx, "all-gender restroom", "")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should not ask the store to locate a room", func(t *testing.T) {
			if len(store.located) != 0 {
				t.Fatalf("located %q", store.located)
			}
		})
		t.Run("it should keep the store's order", func(t *testing.T) {
			if ans.Places[0].ID != hyattRestroom.ID {
				t.Fatalf("first place %s", ans.Places[0].ID)
			}
		})
	})

	t.Run("when the store returns more than five places", func(t *testing.T) {
		store := &fakeStore{places: make([]Place, 8)}
		ans, _ := (&Service{Store: store}).Ask(ctx, "elevator", "")
		t.Run("it should keep five", func(t *testing.T) {
			if len(ans.Places) != 5 {
				t.Fatalf("got %d places", len(ans.Places))
			}
		})
	})

	t.Run("when the question is empty", func(t *testing.T) {
		store := &fakeStore{places: []Place{outside255}}
		ans, err := (&Service{Store: store}).Ask(ctx, "  ", "")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should not search", func(t *testing.T) {
			if len(store.searched) != 0 {
				t.Fatalf("searched %q", store.searched)
			}
		})
		t.Run("it should invite a question", func(t *testing.T) {
			if ans.Summary != "Ask where something is, such as the nearest all-gender restroom or quiet room, or search for a session." {
				t.Fatalf("summary %q", ans.Summary)
			}
		})
	})

	t.Run("when the store fails", func(t *testing.T) {
		boom := errors.New("elasticsearch is down")
		_, err := (&Service{Store: &fakeStore{err: boom}}).Ask(ctx, "quiet room", "255")
		t.Run("it should return the store's error", func(t *testing.T) {
			if !errors.Is(err, boom) {
				t.Fatalf("got %v", err)
			}
		})
	})
}

func TestLocate(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name  string
		asked string
		room  Here
		want  Here
	}{
		{"a known room", " Grand Ballroom J ", Here{Room: "Grand Ballroom J", Venue: "Salt Palace", Level: "Level 1"}, Here{Asked: "Grand Ballroom J", Room: "Grand Ballroom J", Venue: "Salt Palace", Level: "Level 1"}},
		{"an unknown room", "the moon", Here{}, Here{Asked: "the moon"}},
		{"nothing", "   ", Here{}, Here{}},
	}
	for _, c := range cases {
		t.Run("when the attendee names "+c.name, func(t *testing.T) {
			got, err := (&Service{Store: &fakeStore{room: c.room}}).Locate(ctx, c.asked)
			if err != nil {
				t.Fatal(err)
			}
			t.Run("it should keep what was asked beside what was found", func(t *testing.T) {
				if got != c.want {
					t.Fatalf("\n got %+v\nwant %+v", got, c.want)
				}
			})
		})
	}
}
