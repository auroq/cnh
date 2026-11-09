package web

import (
	"strings"
	"testing"
)

func TestProfileKnownMarker(t *testing.T) {
	cases := []struct {
		name    string
		target  string
		cookies map[string]string
		known   bool
	}{
		{"the home page is opened with nothing about the attendee", "/", nil, false},
		{"a question is asked with nothing about the attendee", "/ask?q=coat+check", nil, false},
		{"the page is opened with the profile in the query", "/?prefs=step-free", nil, true},
		{"the page is opened with an empty profile in the query", "/?prefs=", nil, true},
		{"the page is opened with the profile cookie", "/", map[string]string{prefsCookie: "step-free"}, true},
		{"the page is opened with only the do-not-re-add cookie", "/", map[string]string{declinedCookie: "step-free"}, true},
		{"the page is reopened after a chip was removed", "/profile?removed=step-free&q=nearest+elevator", nil, true},
		{"the page is reopened after the attendee was forgotten", "/profile?forgot=1", nil, true},
	}
	for _, c := range cases {
		t.Run("when "+c.name, func(t *testing.T) {
			_, body := newClient(t, fakeAsker{}).getWith(c.target, c.cookies)
			want := "it should leave the browser's own copy of the profile alone"
			if c.known {
				want = "it should tell the browser the server knows the profile"
			}
			t.Run(want, func(t *testing.T) {
				if strings.Contains(body, `data-prefs-known="1"`) != c.known {
					t.Fatalf("known marker present: %v in %s", !c.known, body[:strings.Index(body, "<main")])
				}
			})
		})
	}
}
