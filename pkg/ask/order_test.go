package ask

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestLocateOrder(t *testing.T) {
	ctx := context.Background()
	registration := Here{Room: "Registration and badge pick-up", Venue: "Salt Palace"}
	door := Here{Room: "100 South entrance", Venue: "Salt Palace", Side: "100 South"}
	grandJ := Here{Room: "Grand Ballroom J", Venue: "Salt Palace", Level: "Level 1"}
	all := []string{ViaRoom, ViaAlias, ViaSession, ViaFuzzyRoom, ViaNearby}

	cases := []struct {
		name  string
		store *fakeStore
		via   string
		rules []string
		here  Here
	}{
		{"the exact room name", &fakeStore{room: grandJ, alias: door, session: grandJ}, ViaRoom, all[:1], grandJ},
		{"a phrase people use for a place", &fakeStore{alias: registration, session: grandJ}, ViaAlias, all[:2], registration},
		{"the side of the building they stand on", &fakeStore{alias: door, fuzzy: grandJ}, ViaAlias, all[:2], door},
		{"the title of the session they just left", &fakeStore{session: grandJ, fuzzy: grandJ}, ViaSession, all[:3], grandJ},
		{"a misspelt room", &fakeStore{fuzzy: grandJ, nearby: registration}, ViaFuzzyRoom, all[:4], grandJ},
		{"something only the semantic match understands", &fakeStore{nearby: registration}, ViaNearby, all, registration},
		{"nothing anyone knows", &fakeStore{}, "", all, Here{}},
	}
	for _, c := range cases {
		t.Run("when the attendee says "+c.name, func(t *testing.T) {
			got, err := (&Service{Store: c.store}).Locate(ctx, "somewhere")
			if err != nil {
				t.Fatal(err)
			}
			t.Run("it should stop at the rule that matched", func(t *testing.T) {
				if strings.Join(c.store.rules, ",") != strings.Join(c.rules, ",") {
					t.Fatalf("rules %q, want %q", c.store.rules, c.rules)
				}
			})
			t.Run("it should say which rule matched", func(t *testing.T) {
				if got.Via != c.via {
					t.Fatalf("via %q, want %q", got.Via, c.via)
				}
			})
			t.Run("it should carry the place with its side beside what was asked", func(t *testing.T) {
				want := c.here
				want.Asked, want.Via = "somewhere", c.via
				if got != want {
					t.Fatalf("\n got %+v\nwant %+v", got, want)
				}
			})
		})
	}

	t.Run("when the session rule runs", func(t *testing.T) {
		now := time.Date(2026, 11, 10, 10, 0, 0, 0, time.UTC)
		store := &fakeStore{}
		if _, err := (&Service{Store: store, Now: func() time.Time { return now }}).Locate(ctx, "the flux keynote"); err != nil {
			t.Fatal(err)
		}
		t.Run("it should pass the service's clock so recency is testable", func(t *testing.T) {
			if !store.sessionNow.Equal(now) {
				t.Fatalf("now %v", store.sessionNow)
			}
		})
	})

	t.Run("when a rule fails", func(t *testing.T) {
		boom := errors.New("elasticsearch is down")
		_, err := (&Service{Store: &fakeStore{err: boom}}).Locate(ctx, "255 BC")
		t.Run("it should return the store's error", func(t *testing.T) {
			if !errors.Is(err, boom) {
				t.Fatalf("got %v", err)
			}
		})
	})
}

func TestOutside(t *testing.T) {
	ctx := context.Background()
	south200 := Place{ID: "step-free-200-south", Kind: "step-free-entrance", Name: "Step-free entrance, 200 South", Venue: "Salt Palace", Side: "200 South", Source: "floor plan"}
	south100 := Place{ID: "step-free-100-south", Kind: "step-free-entrance", Name: "Step-free entrance, 100 South", Venue: "Salt Palace", Side: "100 South", Source: "floor plan"}
	door := Here{Room: "100 South entrance", Venue: "Salt Palace", Side: "100 South"}

	t.Run("when the attendee is outside on the 100 South side", func(t *testing.T) {
		t.Run("and asks for a way in without stairs", func(t *testing.T) {
			store := &fakeStore{alias: door, places: []Place{south200, south100}}
			ans, err := (&Service{Store: store}).Ask(ctx, "how do I get in without stairs", "the 100 S side")
			if err != nil {
				t.Fatal(err)
			}
			t.Run("it should ask the store for step-free entrances", func(t *testing.T) {
				if len(store.kinds) == 0 || store.kinds[0] != "step-free-entrance" {
					t.Fatalf("kinds %q", store.kinds)
				}
			})
			t.Run("it should put the door on their side first", func(t *testing.T) {
				if ans.Places[0].ID != south100.ID {
					t.Fatalf("first %s", ans.Places[0].ID)
				}
			})
			t.Run("it should say the door is on their side", func(t *testing.T) {
				want := "Step-free entrance, 100 South: Salt Palace, 100 South side. Same side of the building as you. Unverified: read off a plan, nobody has checked it yet."
				if ans.Summary != want {
					t.Fatalf("\n got %q\nwant %q", ans.Summary, want)
				}
			})
		})

		t.Run("and the only step-free door is on another side", func(t *testing.T) {
			store := &fakeStore{alias: door, places: []Place{south200}}
			ans, err := (&Service{Store: store}).Ask(ctx, "without stairs", "the 100 S side")
			if err != nil {
				t.Fatal(err)
			}
			t.Run("it should name both sides", func(t *testing.T) {
				want := "Step-free entrance, 200 South: Salt Palace, 200 South side. You are on the 100 South side, this door is on the 200 South side. Unverified: read off a plan, nobody has checked it yet."
				if ans.Summary != want {
					t.Fatalf("\n got %q\nwant %q", ans.Summary, want)
				}
			})
		})

		t.Run("and asks for something that names its own kind", func(t *testing.T) {
			store := &fakeStore{alias: door, places: []Place{south200}}
			if _, err := (&Service{Store: store}).Ask(ctx, "nearest elevator", "the 100 S side"); err != nil {
				t.Fatal(err)
			}
			t.Run("it should ask for that kind, not for doors", func(t *testing.T) {
				if store.kinds[0] != "elevator" {
					t.Fatalf("kinds %q", store.kinds)
				}
			})
		})
	})

	t.Run("when the attendee is inside and asks without stairs", func(t *testing.T) {
		store := &fakeStore{room: Here{Room: "255 BC", Venue: "Salt Palace", Level: "Level 2"}, places: []Place{south200}}
		if _, err := (&Service{Store: store}).Ask(ctx, "restroom without stairs", "255 BC"); err != nil {
			t.Fatal(err)
		}
		t.Run("it should not turn the question into one about doors", func(t *testing.T) {
			if store.kinds[0] == "step-free-entrance" {
				t.Fatalf("kinds %q", store.kinds)
			}
		})
	})
}
