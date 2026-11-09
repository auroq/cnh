package qr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/auroq/cnh/pkg/venue"
)

func TestURL(t *testing.T) {
	cases := []struct {
		name string
		base string
		id   string
		want string
	}{
		{"a production base", "https://widedoor.cnh.auroq.com", "room-salt-palace-level-1-grand-ballroom-j", "https://widedoor.cnh.auroq.com/?from=room-salt-palace-level-1-grand-ballroom-j"},
		{"a base with a trailing slash", "https://widedoor.cnh.auroq.com/", "quiet-room-260a", "https://widedoor.cnh.auroq.com/?from=quiet-room-260a"},
		{"an id that needs escaping", "http://localhost:8080", "a b&c", "http://localhost:8080/?from=a+b%26c"},
	}
	for _, c := range cases {
		t.Run("when the base is "+c.name, func(t *testing.T) {
			got := Codes{BaseURL: c.base}.URL(c.id)
			t.Run("it should point at the home page with from set to the id", func(t *testing.T) {
				if got != c.want {
					t.Fatalf("\n got %q\nwant %q", got, c.want)
				}
			})
		})
	}
}

func TestWrite(t *testing.T) {
	grandJ := venue.Place{ID: "room-salt-palace-level-1-grand-ballroom-j", Kind: "room", Name: "Grand Ballroom J", Venue: "Salt Palace", Level: "Level 1", Verified: true, Source: "x"}
	quiet := venue.Place{ID: "quiet-room-260a", Kind: "quiet-room", Name: "All-attendee quiet room", Venue: "Salt Palace", Level: "Level 2", Verified: true, Source: "x"}
	elevator := venue.Place{ID: "elevators-salt-palace", Kind: "elevator", Name: "Elevators", Venue: "Salt Palace", Source: "x"}

	t.Run("when codes are written for a room, a quiet room and an elevator", func(t *testing.T) {
		dir := t.TempDir()
		codes := Codes{BaseURL: "https://widedoor.cnh.auroq.com"}
		written, err := codes.Write(dir, []venue.Place{grandJ, quiet, elevator}, true)
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should report the two places that get a code", func(t *testing.T) {
			if len(written) != 2 || written[0].ID != grandJ.ID || written[1].ID != quiet.ID {
				t.Fatalf("written %+v", written)
			}
		})
		for _, p := range []venue.Place{grandJ, quiet} {
			for _, ext := range []string{".svg", ".png"} {
				t.Run("and "+p.ID+ext+" is looked for", func(t *testing.T) {
					t.Run("it should exist", func(t *testing.T) {
						if _, err := os.Stat(filepath.Join(dir, p.ID+ext)); err != nil {
							t.Fatal(err)
						}
					})
				})
			}
		}
		t.Run("and the elevator's code is looked for", func(t *testing.T) {
			t.Run("it should not exist, because nobody stands at an elevator to scan it", func(t *testing.T) {
				if _, err := os.Stat(filepath.Join(dir, elevator.ID+".svg")); !os.IsNotExist(err) {
					t.Fatalf("stat = %v", err)
				}
			})
		})
		t.Run("and the SVG is read", func(t *testing.T) {
			svg, err := os.ReadFile(filepath.Join(dir, grandJ.ID+".svg"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("it should be a scalable drawing with a viewBox", func(t *testing.T) {
				if !strings.HasPrefix(string(svg), "<svg") || !strings.Contains(string(svg), "viewBox=") {
					t.Fatalf("svg = %.80s", svg)
				}
			})
			t.Run("it should name the place for a screen reader", func(t *testing.T) {
				if !strings.Contains(string(svg), "<title>Wide Door code for Grand Ballroom J</title>") {
					t.Fatalf("svg = %.200s", svg)
				}
			})
		})
		t.Run("and the sheet is read", func(t *testing.T) {
			sheet, err := os.ReadFile(filepath.Join(dir, "sheet.html"))
			if err != nil {
				t.Fatal(err)
			}
			s := string(sheet)
			checks := map[string]string{
				"print the place name in large type":   `<h1 class="name">Grand Ballroom J</h1>`,
				"print the URL as text under the code": codes.URL(grandJ.ID) + "</p>",
				"say the building and level":           "Salt Palace, Level 1",
				"have one page per place":              `class="page"`,
				"set the name size to at least 48pt":   "font-size: 48pt",
			}
			for name, want := range checks {
				t.Run("it should "+name, func(t *testing.T) {
					if !strings.Contains(s, want) {
						t.Fatalf("%q missing in %s", want, s)
					}
				})
			}
			t.Run("it should hold one page per place written", func(t *testing.T) {
				if n := strings.Count(s, `class="page"`); n != 2 {
					t.Fatalf("%d pages", n)
				}
			})
		})
	})

	t.Run("when codes are written without a sheet", func(t *testing.T) {
		dir := t.TempDir()
		if _, err := (Codes{BaseURL: "https://widedoor.cnh.auroq.com"}).Write(dir, []venue.Place{grandJ}, false); err != nil {
			t.Fatal(err)
		}
		t.Run("it should not write sheet.html", func(t *testing.T) {
			if _, err := os.Stat(filepath.Join(dir, "sheet.html")); !os.IsNotExist(err) {
				t.Fatalf("stat = %v", err)
			}
		})
	})
}
