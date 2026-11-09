package sessionize

import (
	"strings"
	"testing"
	"time"
)

const index = `# KubeCon + CloudNativeCon North America 2026

- **Timezone:** Mountain Standard Time (UTC-07)

## Sessions

### Monday, November 9, 2026

- [Coat + Bag Check](https://sessionize.com/api/v2/uo948i5p/view/markdown/session/1292207)
  7:00 am - 7:00 pm · Salt Palace | Level 2 | East Registration
- [Agent Rush Hackathon Hosted by Vultr | SOLD OUT](https://sessionize.com/api/v2/uo948i5p/view/markdown/session/1344383)
  9:00 am - 6:00 pm · Off-Site Location
- [Project Lightning Talk - Project Overview](https://sessionize.com/api/v2/uo948i5p/view/markdown/session/1289969)
  Jorge Castro, Katherine Druckman · 9:37 am - 9:52 am · Hyatt Regency | Level 4 | Regency Ballroom B-D

### Tuesday, November 10, 2026

- [Keynote: Welcome](https://sessionize.com/api/v2/uo948i5p/view/markdown/session/1000001)
  Someone Else · 9:00 am - 9:10 am · Salt Palace | Level 1 | Hall 1

## Speakers

- [Jorge Castro](https://sessionize.com/api/v2/uo948i5p/view/markdown/speaker/abc)
`

const detail = `# Coat + Bag Check

KubeCon · [All sessions](https://sessionize.com/api/v2/uo948i5p/view/markdown)

- **Session ID:** 1292207
- **When:** Monday, November 9, 2026 7:00 am - 7:00 pm (Mountain Standard Time, UTC-07)
- **Room:** Salt Palace | Level 2 | East Registration

## Description

Leave your coat here.
Pick it up by 7 pm.

## Speakers

- Nobody
`

func TestParseIndex(t *testing.T) {
	sessions, err := ParseIndex(strings.NewReader(index))
	if err != nil {
		t.Fatalf("ParseIndex: %v", err)
	}
	byID := map[string]Session{}
	for _, s := range sessions {
		byID[s.ID] = s
	}
	mst := time.FixedZone("MST", -7*3600)

	t.Run("when the index lists sessions under day headings", func(t *testing.T) {
		t.Run("it should return one session per bullet", func(t *testing.T) {
			if len(sessions) != 4 {
				t.Fatalf("got %d sessions, want 4", len(sessions))
			}
		})
		t.Run("it should stop before the speakers section", func(t *testing.T) {
			if _, ok := byID["abc"]; ok {
				t.Fatal("a speaker was parsed as a session")
			}
		})
	})

	t.Run("when a session has a venue, level and room", func(t *testing.T) {
		s := byID["1292207"]
		cases := []struct{ name, got, want string }{
			{"title", s.Title, "Coat + Bag Check"},
			{"venue", s.Venue, "Salt Palace"},
			{"level", s.Level, "Level 2"},
			{"room", s.Room, "East Registration"},
			{"location", s.Location, "Salt Palace | Level 2 | East Registration"},
		}
		for _, c := range cases {
			t.Run("it should parse the "+c.name, func(t *testing.T) {
				if c.got != c.want {
					t.Fatalf("%s = %q, want %q", c.name, c.got, c.want)
				}
			})
		}
		t.Run("it should parse the start time in Mountain time", func(t *testing.T) {
			want := time.Date(2026, 11, 9, 7, 0, 0, 0, mst)
			if !s.Start.Equal(want) {
				t.Fatalf("start = %v, want %v", s.Start, want)
			}
		})
		t.Run("it should parse the end time on the same day", func(t *testing.T) {
			want := time.Date(2026, 11, 9, 19, 0, 0, 0, mst)
			if !s.End.Equal(want) {
				t.Fatalf("end = %v, want %v", s.End, want)
			}
		})
	})

	t.Run("when a session has speakers before the time", func(t *testing.T) {
		s := byID["1289969"]
		t.Run("it should split the speakers on commas", func(t *testing.T) {
			if len(s.Speakers) != 2 || s.Speakers[0] != "Jorge Castro" || s.Speakers[1] != "Katherine Druckman" {
				t.Fatalf("speakers = %q", s.Speakers)
			}
		})
		t.Run("it should still parse the room", func(t *testing.T) {
			if s.Room != "Regency Ballroom B-D" {
				t.Fatalf("room = %q", s.Room)
			}
		})
	})

	t.Run("when a session has a location with more than three parts", func(t *testing.T) {
		sessions, _ := ParseIndex(strings.NewReader("## Sessions\n\n### Monday, November 9, 2026\n\n- [X](https://s/session/9)\n  9:00 am - 9:30 am · Salt Palace | Level 1 | Solutions Showcase | Halls 5 | Cosmic Canyon Zone\n"))
		t.Run("it should keep the level and join the rest as the room", func(t *testing.T) {
			if sessions[0].Level != "Level 1" || sessions[0].Room != "Solutions Showcase, Halls 5, Cosmic Canyon Zone" {
				t.Fatalf("level=%q room=%q", sessions[0].Level, sessions[0].Room)
			}
		})
	})

	t.Run("when a session is off-site with no level or room", func(t *testing.T) {
		s := byID["1344383"]
		t.Run("it should keep the location as the venue", func(t *testing.T) {
			if s.Venue != "Off-Site Location" || s.Level != "" || s.Room != "" {
				t.Fatalf("venue=%q level=%q room=%q", s.Venue, s.Level, s.Room)
			}
		})
	})

	t.Run("when a session is on a later day heading", func(t *testing.T) {
		s := byID["1000001"]
		t.Run("it should take its date from that heading", func(t *testing.T) {
			if s.Start.Day() != 10 {
				t.Fatalf("start = %v, want the 10th", s.Start)
			}
		})
	})
}

func TestParseDetail(t *testing.T) {
	d, err := ParseDetail(strings.NewReader(detail))
	if err != nil {
		t.Fatalf("ParseDetail: %v", err)
	}
	t.Run("when the detail has a description section", func(t *testing.T) {
		t.Run("it should return the description text only", func(t *testing.T) {
			want := "Leave your coat here.\nPick it up by 7 pm."
			if d.Description != want {
				t.Fatalf("description = %q, want %q", d.Description, want)
			}
		})
		t.Run("it should return the session id", func(t *testing.T) {
			if d.ID != "1292207" {
				t.Fatalf("id = %q", d.ID)
			}
		})
	})
}
