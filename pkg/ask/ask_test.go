package ask

import (
	"strings"
	"testing"
	"time"
)

func TestRankByNearness(t *testing.T) {
	places := []Place{
		{ID: "hyatt-l2", Venue: "Hyatt Regency", Level: "Level 2"},
		{ID: "sp-l1", Venue: "Salt Palace", Level: "Level 1"},
		{ID: "sp-l2-a", Venue: "Salt Palace", Level: "Level 2"},
		{ID: "sp-l2-b", Venue: "Salt Palace", Level: "Level 2"},
	}
	ids := func(ps []Place) string {
		var out []string
		for _, p := range ps {
			out = append(out, p.ID)
		}
		return strings.Join(out, ",")
	}

	t.Run("when the attendee is on Salt Palace Level 2", func(t *testing.T) {
		got := RankByNearness(places, Here{Venue: "Salt Palace", Level: "Level 2"})
		t.Run("it should put same-level places first in search order, then same building, then others", func(t *testing.T) {
			if ids(got) != "sp-l2-a,sp-l2-b,sp-l1,hyatt-l2" {
				t.Fatalf("got %s", ids(got))
			}
		})
	})

	t.Run("when the attendee's location is unknown", func(t *testing.T) {
		got := RankByNearness(places, Here{})
		t.Run("it should keep the search order", func(t *testing.T) {
			if ids(got) != ids(places) {
				t.Fatalf("got %s", ids(got))
			}
		})
	})
}

func TestSummarise(t *testing.T) {
	restroom := Place{Name: "All-gender restroom outside 255", Venue: "Salt Palace", Level: "Level 2", Near: "Outside room 255", Verified: true}
	elevator := Place{Name: "Elevators", Venue: "Salt Palace", Verified: false}
	cases := []struct {
		name string
		ans  Answer
		want string
	}{
		{"same level", Answer{Here: Here{Venue: "Salt Palace", Level: "Level 2"}, Places: []Place{restroom}},
			"All-gender restroom outside 255: Salt Palace, Level 2, Outside room 255. Same level as you, no stairs or elevator. Published by the organiser."},
		{"other building", Answer{Here: Here{Venue: "Hyatt Regency", Level: "Level 4"}, Places: []Place{restroom}},
			"All-gender restroom outside 255: Salt Palace, Level 2, Outside room 255. It is in the Salt Palace, you are in the Hyatt Regency. Published by the organiser."},
		{"unverified", Answer{Places: []Place{elevator}},
			"Elevators: Salt Palace. Unverified: read off a plan, nobody has checked it yet."},
		{"nothing", Answer{}, "Nothing matched. Try other words, or ask staff at registration."},
		{"a question about talks", Answer{Question: "talks about accessibility", Places: []Place{restroom}, Sessions: []Session{{Title: "Sign Language Crash Course", Start: time.Date(2026, 11, 11, 20, 30, 0, 0, time.UTC), Location: "Salt Palace | Level 1 | Grand Ballroom J"}}},
			"1 sessions match. First: Sign Language Crash Course, Wednesday 1:30 pm, Salt Palace, Level 1, Grand Ballroom J."},
	}
	for _, c := range cases {
		t.Run("when the answer is "+c.name, func(t *testing.T) {
			t.Run("it should say where, how far, and whether it is verified", func(t *testing.T) {
				if got := Summarise(c.ans); got != c.want {
					t.Fatalf("\n got %q\nwant %q", got, c.want)
				}
			})
		})
	}
}
