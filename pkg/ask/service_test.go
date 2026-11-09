package ask

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeStore struct {
	room          Here
	fuzzy         Here
	alias         Here
	session       Here
	nearby        Here
	places        []Place
	sessions      []Session
	err           error
	located       []string
	rules         []string
	sessionNow    time.Time
	searched      []string
	kinds         []string
	noneOfAnyKind bool
}

func (f *fakeStore) Ping(context.Context) error { return f.err }

func (f *fakeStore) LocateRoom(_ context.Context, asked string) (Here, error) {
	f.located = append(f.located, asked)
	f.rules = append(f.rules, ViaRoom)
	return f.room, f.err
}

func (f *fakeStore) LocateRoomFuzzy(context.Context, string) (Here, error) {
	f.rules = append(f.rules, ViaFuzzyRoom)
	return f.fuzzy, f.err
}

func (f *fakeStore) LocateAlias(context.Context, string) (Here, error) {
	f.rules = append(f.rules, ViaAlias)
	return f.alias, f.err
}

func (f *fakeStore) LocateSession(_ context.Context, _ string, now time.Time) (Here, error) {
	f.rules = append(f.rules, ViaSession)
	f.sessionNow = now
	return f.session, f.err
}

func (f *fakeStore) LocateNearby(context.Context, string) (Here, error) {
	f.rules = append(f.rules, ViaNearby)
	return f.nearby, f.err
}

func (f *fakeStore) FindPlaces(_ context.Context, question, kind string) ([]Place, error) {
	f.searched = append(f.searched, question)
	f.kinds = append(f.kinds, kind)
	if kind != "" && f.noneOfAnyKind {
		return nil, f.err
	}
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
		t.Run("it should ask the store for all-gender restrooms only", func(t *testing.T) {
			if len(store.kinds) != 1 || store.kinds[0] != "all-gender-restroom" {
				t.Fatalf("kinds %q", store.kinds)
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

	t.Run("when the attendee in 255 BC asks for the nearest elevator", func(t *testing.T) {
		elevator := Place{ID: "elevators-salt-palace", Kind: "elevator", Name: "Elevators", Venue: "Salt Palace"}
		store := &fakeStore{
			room:   Here{Room: "255 BC", Venue: "Salt Palace", Level: "Level 2"},
			places: []Place{elevator},
		}
		ans, err := (&Service{Store: store}).Ask(ctx, "nearest elevator", "255 BC")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should ask the store for elevators only", func(t *testing.T) {
			if len(store.kinds) != 1 || store.kinds[0] != "elevator" {
				t.Fatalf("kinds %q", store.kinds)
			}
		})
		t.Run("it should answer the elevator", func(t *testing.T) {
			if ans.Places[0].ID != elevator.ID {
				t.Fatalf("first place %s", ans.Places[0].ID)
			}
		})
		t.Run("it should say it is in the same building on an unlisted level", func(t *testing.T) {
			want := "Elevators: Salt Palace. Same building, you are on Level 2 and it is on an unlisted level. Unverified: read off a plan, nobody has checked it yet."
			if ans.Summary != want {
				t.Fatalf("\n got %q\nwant %q", ans.Summary, want)
			}
		})
	})

	t.Run("when the question names a kind the store has no place for", func(t *testing.T) {
		store := &fakeStore{places: []Place{outside255}, noneOfAnyKind: true}
		ans, err := (&Service{Store: store}).Ask(ctx, "nearest elevator", "")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should fall back to places of any kind", func(t *testing.T) {
			if len(store.kinds) != 2 || store.kinds[0] != "elevator" || store.kinds[1] != "" {
				t.Fatalf("kinds %q", store.kinds)
			}
		})
		t.Run("it should answer the best place of any kind", func(t *testing.T) {
			if len(ans.Places) == 0 || ans.Places[0].ID != outside255.ID {
				t.Fatalf("places %+v", ans.Places)
			}
		})
	})

	t.Run("when the question names no kind", func(t *testing.T) {
		store := &fakeStore{places: []Place{outside255}}
		_, err := (&Service{Store: store}).Ask(ctx, "sign language", "")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should ask the store for places of any kind", func(t *testing.T) {
			if len(store.kinds) != 1 || store.kinds[0] != "" {
				t.Fatalf("kinds %q", store.kinds)
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
		{"a known room", " Grand Ballroom J ", Here{Room: "Grand Ballroom J", Venue: "Salt Palace", Level: "Level 1"}, Here{Asked: "Grand Ballroom J", Room: "Grand Ballroom J", Venue: "Salt Palace", Level: "Level 1", Via: ViaRoom}},
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
