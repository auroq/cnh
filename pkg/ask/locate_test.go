package ask

import (
	"context"
	"testing"
)

type fakeIDStore struct {
	*fakeStore
	byID     map[string]Here
	lookedUp []string
}

func (f *fakeIDStore) LocatePlace(_ context.Context, id string) (Here, error) {
	f.lookedUp = append(f.lookedUp, id)
	return f.byID[id], nil
}

func TestLocateByID(t *testing.T) {
	ctx := context.Background()
	grandJ := Here{Room: "Grand Ballroom J", Venue: "Salt Palace", Level: "Level 1"}
	fuzzy := Here{Room: "255 BC", Venue: "Salt Palace", Level: "Level 2"}

	t.Run("when the location is a place id the store knows", func(t *testing.T) {
		store := &fakeIDStore{fakeStore: &fakeStore{room: fuzzy}, byID: map[string]Here{"room-salt-palace-level-1-grand-ballroom-j": grandJ}}
		here, err := (&Service{Store: store}).Locate(ctx, " room-salt-palace-level-1-grand-ballroom-j ")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should resolve it by id", func(t *testing.T) {
			if here.Room != grandJ.Room || here.Level != grandJ.Level {
				t.Fatalf("here %+v", here)
			}
		})
		t.Run("it should keep what was asked", func(t *testing.T) {
			if here.Asked != "room-salt-palace-level-1-grand-ballroom-j" {
				t.Fatalf("asked %q", here.Asked)
			}
		})
		t.Run("it should not run the room search", func(t *testing.T) {
			if len(store.located) != 0 {
				t.Fatalf("located %q", store.located)
			}
		})
	})

	t.Run("when the location looks like an id nobody knows", func(t *testing.T) {
		store := &fakeIDStore{fakeStore: &fakeStore{room: fuzzy}, byID: map[string]Here{}}
		here, err := (&Service{Store: store}).Locate(ctx, "room-nowhere")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should fall back to the room search", func(t *testing.T) {
			if len(store.located) != 1 || here.Room != fuzzy.Room {
				t.Fatalf("located %q, here %+v", store.located, here)
			}
		})
	})

	notIDs := []string{"255 BC", "Grand Ballroom J", "Zion", "260A"}
	for _, asked := range notIDs {
		t.Run("when the location is "+asked, func(t *testing.T) {
			store := &fakeIDStore{fakeStore: &fakeStore{room: fuzzy}, byID: map[string]Here{}}
			if _, err := (&Service{Store: store}).Locate(ctx, asked); err != nil {
				t.Fatal(err)
			}
			t.Run("it should go straight to the room search", func(t *testing.T) {
				if len(store.lookedUp) != 0 || len(store.located) != 1 {
					t.Fatalf("looked up %q, located %q", store.lookedUp, store.located)
				}
			})
		})
	}

	t.Run("when the store cannot look places up by id", func(t *testing.T) {
		store := &fakeStore{room: fuzzy}
		here, err := (&Service{Store: store}).Locate(ctx, "room-salt-palace-level-2-255-bc")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should use the room search", func(t *testing.T) {
			if len(store.located) != 1 || here.Room != fuzzy.Room {
				t.Fatalf("located %q, here %+v", store.located, here)
			}
		})
	})
}
