package venue

import (
	"os"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("when loading the committed places.yaml", func(t *testing.T) {
		f, err := os.Open("../../data/venue/places.yaml")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		places, err := Load(f)
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		t.Run("it should load every place with a verified flag and a source", func(t *testing.T) {
			for _, p := range places {
				if p.Source == "" {
					t.Fatalf("%s has no source", p.ID)
				}
			}
		})
		t.Run("it should mark the elevators unverified", func(t *testing.T) {
			for _, p := range places {
				if p.Kind == "elevator" && p.Verified {
					t.Fatalf("%s is an elevator marked verified; the organiser does not publish them", p.ID)
				}
			}
		})
	})

	t.Run("when a place lists route facts", func(t *testing.T) {
		places, err := Load(strings.NewReader("places:\n- {id: a, kind: k, name: n, venue: v, source: s, route: [\"Connected at Level 1.\", \"Step-free.\"]}\n"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should keep them in order", func(t *testing.T) {
			if len(places) != 1 || strings.Join(places[0].Route, "|") != "Connected at Level 1.|Step-free." {
				t.Fatalf("got %+v", places)
			}
		})
	})

	t.Run("when two places share an id", func(t *testing.T) {
		_, err := Load(strings.NewReader("places:\n- {id: a, kind: k, name: n, venue: v, source: s}\n- {id: a, kind: k, name: n, venue: v, source: s}\n"))
		t.Run("it should fail", func(t *testing.T) {
			if err == nil {
				t.Fatal("expected a duplicate id error")
			}
		})
	})
}

func TestSplitLocation(t *testing.T) {
	cases := []struct {
		name, in   string
		venue, lvl string
		room       string
	}{
		{"three parts", "Salt Palace | Level 2 | 255 BC", "Salt Palace", "Level 2", "255 BC"},
		{"five parts", "Salt Palace | Level 1 | Solutions Showcase | Halls 5 | Cosmic Canyon Zone", "Salt Palace", "Level 1", "Solutions Showcase, Halls 5, Cosmic Canyon Zone"},
		{"no level", "Salt Palace | 200 South Entrance (South)", "Salt Palace", "", "200 South Entrance (South)"},
		{"venue only", "Off-Site Location", "Off-Site Location", "", ""},
	}
	for _, c := range cases {
		t.Run("when the location has "+c.name, func(t *testing.T) {
			r, ok := SplitLocation(c.in)
			if !ok {
				t.Fatal("not ok")
			}
			t.Run("it should split venue, level and room", func(t *testing.T) {
				if r.Venue != c.venue || r.Level != c.lvl || r.Room != c.room {
					t.Fatalf("got %+v", r)
				}
			})
		})
	}
}

func TestRoomsFromLocations(t *testing.T) {
	rooms := RoomsFromLocations([]string{
		"Salt Palace | Level 2 | 255 BC",
		"Salt Palace | Level 2 | 255 BC",
		"Off-Site Location",
	})
	t.Run("when locations repeat or have no room", func(t *testing.T) {
		t.Run("it should return one verified room per distinct room", func(t *testing.T) {
			if len(rooms) != 1 || !rooms[0].Verified || rooms[0].Name != "255 BC" {
				t.Fatalf("got %+v", rooms)
			}
		})
	})
}
