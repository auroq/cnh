package web

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/auroq/cnh/pkg/model"
)

func TestContractV1(t *testing.T) {
	recorded, err := os.ReadFile("testdata/ask_response.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(string(recorded), "\n"), "\n")

	cases := []struct {
		name   string
		answer model.Answer
		want   string
	}{
		{"a full answer", model.Answer{
			Question: "nearest all-gender restroom",
			Here:     model.Here{Asked: "255 BC", Room: "255 BC", Venue: "Salt Palace", Level: "Level 2"},
			Summary:  "All-gender restroom outside 255: Salt Palace, Level 2, Outside room 255. Same level as you, no stairs or elevator. Published by the organiser.",
			Places: []model.Place{
				{ID: "restroom-sp-l2-255", Kind: "all-gender-restroom", Name: "All-gender restroom outside 255", Venue: "Salt Palace", Level: "Level 2", Near: "Outside room 255", Notes: "Single stalls.", For: []string{"everyone"}, Verified: true, Source: "https://example.org/inclusion"},
				{ID: "elevators-salt-palace", Kind: "elevator", Name: "Elevators", Venue: "Salt Palace", Verified: false, Source: "floor plan"},
			},
			Sessions: []model.Session{
				{ID: "1", Title: "Sign Language Crash Course", URL: "https://sessionize.com/s/1", Speakers: []string{"A", "B"}, Start: time.Date(2026, 11, 11, 20, 30, 0, 0, time.UTC), End: time.Date(2026, 11, 11, 21, 0, 0, 0, time.UTC), Location: "Salt Palace | Level 1 | Grand Ballroom J", Venue: "Salt Palace", Level: "Level 1", Room: "Grand Ballroom J"},
			},
		}, lines[0]},
		{"an empty answer", model.Answer{Question: "x"}, lines[1]},
	}
	for _, c := range cases {
		t.Run("when "+c.name+" is encoded for /api/ask", func(t *testing.T) {
			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(toV1(c.answer)); err != nil {
				t.Fatal(err)
			}
			t.Run("it should match the response recorded before the types moved, byte for byte", func(t *testing.T) {
				if got := strings.TrimRight(buf.String(), "\n"); got != c.want {
					t.Fatalf("\n got %s\nwant %s", got, c.want)
				}
			})
		})
	}
}
