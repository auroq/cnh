package acceptance

import (
	"strings"
	"testing"

	"github.com/auroq/cnh/testing/helpers"
)

type answer struct {
	Summary string `json:"summary"`
	Here    struct {
		Venue string `json:"venue"`
	} `json:"here"`
	Places []struct {
		ID    string `json:"id"`
		Level string `json:"level"`
	} `json:"places"`
	Sessions []struct {
		Title string `json:"title"`
	} `json:"sessions"`
}

func ask(t *testing.T, question, from string) answer {
	t.Helper()
	var a answer
	helpers.GetJSON(t, "/api/ask", helpers.Query("q", question, "from", from), &a)
	return a
}

func TestAsk(t *testing.T) {
	t.Parallel()

	t.Run("when I ask for the nearest all-gender restroom", func(t *testing.T) {
		t.Parallel()
		t.Run("and I am in 255 BC", func(t *testing.T) {
			a := ask(t, "nearest all-gender restroom", "255 BC")
			t.Run("it should answer the restroom outside 255", func(t *testing.T) {
				if len(a.Places) == 0 || a.Places[0].ID != "agr-sp-l2-255" {
					t.Fatalf("places = %+v", a.Places)
				}
			})
			t.Run("it should say it is on the same level", func(t *testing.T) {
				if !strings.Contains(a.Summary, "Same level as you") {
					t.Fatalf("summary = %q", a.Summary)
				}
			})
		})
	})

	t.Run("when I ask without saying where I am", func(t *testing.T) {
		t.Parallel()
		a := ask(t, "nearest all-gender restroom", "")
		t.Run("it should still answer", func(t *testing.T) {
			if len(a.Places) == 0 {
				t.Fatalf("summary = %q, no places", a.Summary)
			}
		})
		t.Run("it should leave the location unresolved", func(t *testing.T) {
			if a.Here.Venue != "" {
				t.Fatalf("here.venue = %q", a.Here.Venue)
			}
		})
	})

	t.Run("when I ask about sign language", func(t *testing.T) {
		t.Parallel()
		a := ask(t, "sign language", "")
		t.Run("it should list the Sign Language Crash Course among sessions", func(t *testing.T) {
			for _, s := range a.Sessions {
				if s.Title == "Sign Language Crash Course" {
					return
				}
			}
			t.Fatalf("sessions = %+v", a.Sessions)
		})
	})
}
