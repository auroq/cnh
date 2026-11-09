package acceptance

import (
	"strings"
	"testing"

	"github.com/auroq/cnh/testing/helpers"
)

type answer struct {
	Summary string `json:"summary"`
	Here    here   `json:"here"`
	Places  []struct {
		ID       string `json:"id"`
		Kind     string `json:"kind"`
		Verified bool   `json:"verified"`
		Side     string `json:"side"`
	} `json:"places"`
	Sessions []struct {
		Title string `json:"title"`
	} `json:"sessions"`
}

type here struct {
	Asked string `json:"asked"`
	Room  string `json:"room"`
	Venue string `json:"venue"`
	Level string `json:"level"`
	Side  string `json:"side"`
	Via   string `json:"via"`
}

func ask(t *testing.T, question, from string) answer {
	t.Helper()
	var a answer
	helpers.GetJSON(t, "/api/ask", helpers.Query("q", question, "from", from), &a)
	return a
}

var distanceWords = map[distance]string{
	sameLevel:     "Same level as you",
	sameBuilding:  "Same building,",
	otherBuilding: "It is in the",
}

func saysDistance(summary string, d distance) bool {
	if d == noDistance {
		for _, w := range distanceWords {
			if strings.Contains(summary, w) {
				return false
			}
		}
		return true
	}
	return strings.Contains(summary, distanceWords[d])
}

func whereIAm(from string) string {
	if from == "" {
		return "and I do not say where I am"
	}
	return "and I am in " + from
}

func verifiedWording(verified bool) string {
	if verified {
		return "it should say the organiser published it"
	}
	return "it should say it is unverified"
}

func TestAsk(t *testing.T) {
	t.Parallel()

	for _, q := range cases {
		t.Run("when I ask for "+q.Ask, func(t *testing.T) {
			t.Parallel()
			for _, row := range q.Rows {
				t.Run(whereIAm(row.From), func(t *testing.T) {
					t.Parallel()
					a := ask(t, q.Ask, row.From)
					t.Run("it should answer "+row.Place+" first", func(t *testing.T) {
						if len(a.Places) == 0 || a.Places[0].ID != row.Place {
							t.Fatalf("places = %+v", a.Places)
						}
					})
					t.Run("it should say "+row.Distance.String(), func(t *testing.T) {
						if !saysDistance(a.Summary, row.Distance) {
							t.Fatalf("summary = %q", a.Summary)
						}
					})
					t.Run(verifiedWording(row.Verified), func(t *testing.T) {
						if len(a.Places) == 0 || a.Places[0].Verified != row.Verified {
							t.Fatalf("summary = %q, places = %+v", a.Summary, a.Places)
						}
					})
				})
			}
		})
	}

	t.Run("when I ask from a room nobody has heard of", func(t *testing.T) {
		t.Parallel()
		a := ask(t, "nearest all-gender restroom", "Narnia")
		t.Run("it should echo what I typed", func(t *testing.T) {
			if a.Here.Asked != "Narnia" {
				t.Fatalf("here = %+v", a.Here)
			}
		})
		t.Run("it should leave the room unresolved", func(t *testing.T) {
			if a.Here.Room != "" || a.Here.Venue != "" || a.Here.Level != "" {
				t.Fatalf("here = %+v", a.Here)
			}
		})
	})

	for _, c := range needCases {
		t.Run("when I say "+c.Say, func(t *testing.T) {
			t.Parallel()
			a := ask(t, c.Say, "")
			t.Run("it should understand I need "+c.Need, func(t *testing.T) {
				if !strings.HasPrefix(a.Summary, "You asked for "+c.Need+". ") {
					t.Fatalf("summary = %q", a.Summary)
				}
			})
			t.Run("it should offer a "+c.Kind+" first", func(t *testing.T) {
				if len(a.Places) == 0 || a.Places[0].Kind != c.Kind {
					t.Fatalf("places = %+v", a.Places)
				}
			})
			if c.Note != "" {
				t.Run("it should say why the place is a stand-in", func(t *testing.T) {
					if !strings.Contains(a.Summary, c.Note) {
						t.Fatalf("summary = %q", a.Summary)
					}
				})
			}
		})
	}

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
