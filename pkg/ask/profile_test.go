package ask

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/auroq/cnh/pkg/model"
)

func TestAskFor(t *testing.T) {
	ctx := context.Background()
	entrance := Place{ID: "step-free-200-south", Kind: "step-free-entrance", Name: "Step-free entrance, 200 South", Venue: "Salt Palace", Level: "Level 1"}
	showcaseRestroom := Place{ID: "agr-sp-l1-showcase", Kind: "all-gender-restroom", Name: "All-gender restroom, Solutions Showcase", Venue: "Salt Palace", Level: "Level 1", Near: "Solutions Showcase, Amber Hive Zone", Verified: true}
	quiet := Place{ID: "quiet-room-260a", Kind: "quiet-room", Name: "All-attendee quiet room", Venue: "Salt Palace", Level: "Level 2", Verified: true}
	talk := Session{ID: "1", Title: "Sign Language Crash Course", Start: time.Date(2026, 11, 11, 20, 30, 0, 0, time.UTC), End: time.Date(2026, 11, 11, 21, 30, 0, 0, time.UTC), Location: "Salt Palace | Level 1 | Grand Ballroom J"}

	cases := []struct {
		name       string
		prefs      string
		question   string
		store      *fakeStore
		firstKind  string
		kindsAsked string
		summaryHas string
		applied    string
	}{
		{"step-free and the nearest entrance", "step-free", "nearest entrance",
			&fakeStore{byKind: map[string][]Place{"step-free-entrance": {entrance}}},
			"step-free-entrance", "step-free-entrance", "no stairs", "step-free"},
		{"a wheelchair user and the way in", "wheels", "which door do I use to get in",
			&fakeStore{byKind: map[string][]Place{"step-free-entrance": {entrance}}},
			"step-free-entrance", "step-free-entrance", "no stairs", "wheels"},
		{"avoid crowds and a restroom in the Solutions Showcase", "avoid-crowds", "nearest all-gender restroom",
			&fakeStore{places: []Place{showcaseRestroom}},
			"all-gender-restroom", "all-gender-restroom", "crowds and noise", "avoid-crowds"},
		{"quiet matters and somewhere to rest", "quiet-matters", "somewhere to sit and rest",
			&fakeStore{byKind: map[string][]Place{"quiet-room": {quiet}}},
			"quiet-room", "quiet-room", "", "quiet-matters"},
		{"all-gender and a restroom", "all-gender", "nearest restroom",
			&fakeStore{places: []Place{showcaseRestroom}},
			"all-gender-restroom", "all-gender-restroom", "", "all-gender"},
		{"captions and a talk question", "captions-asl", "talk about sign language",
			&fakeStore{sessions: []Session{talk}},
			"", "", "captions", "captions-asl"},
		{"dietary needs and lunch", "dietary", "where can I eat lunch",
			&fakeStore{places: []Place{quiet}},
			"", "", "information desk", "dietary"},
	}
	for _, c := range cases {
		t.Run("when the profile says "+c.name, func(t *testing.T) {
			ans, err := (&Service{Store: c.store, Now: func() time.Time { return noon }}).AskFor(ctx, c.question, "", model.ParsePreferences(c.prefs))
			if err != nil {
				t.Fatal(err)
			}
			if c.firstKind != "" {
				t.Run("it should answer a "+c.firstKind+" first", func(t *testing.T) {
					if len(ans.Places) == 0 || ans.Places[0].Kind != c.firstKind {
						t.Fatalf("places %+v", ans.Places)
					}
				})
			}
			if c.kindsAsked != "" {
				t.Run("it should ask the store for "+c.kindsAsked+" first", func(t *testing.T) {
					if len(c.store.kinds) == 0 || c.store.kinds[0] != c.kindsAsked {
						t.Fatalf("kinds %q", c.store.kinds)
					}
				})
			}
			if c.summaryHas != "" {
				t.Run("it should say "+c.summaryHas, func(t *testing.T) {
					if !strings.Contains(strings.ToLower(ans.Summary), c.summaryHas) {
						t.Fatalf("summary %q", ans.Summary)
					}
				})
			}
			t.Run("it should name the chip that changed the answer", func(t *testing.T) {
				if len(ans.Applied) != 1 || ans.Applied[0].Chip != c.applied || ans.Applied[0].Effect == "" {
					t.Fatalf("applied %+v", ans.Applied)
				}
			})
		})
	}

	t.Run("when the profile has a chip the question does not touch", func(t *testing.T) {
		store := &fakeStore{places: []Place{quiet}}
		ans, err := (&Service{Store: store}).AskFor(ctx, "nearest quiet room", "", model.ParsePreferences("dietary,captions-asl"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should change nothing", func(t *testing.T) {
			if len(ans.Applied) != 0 {
				t.Fatalf("applied %+v", ans.Applied)
			}
		})
	})

	t.Run("when the profile is empty", func(t *testing.T) {
		plain, err := (&Service{Store: &fakeStore{places: []Place{showcaseRestroom}}}).Ask(ctx, "nearest all-gender restroom", "")
		if err != nil {
			t.Fatal(err)
		}
		shaped, err := (&Service{Store: &fakeStore{places: []Place{showcaseRestroom}}}).AskFor(ctx, "nearest all-gender restroom", "", nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should answer exactly as Ask does", func(t *testing.T) {
			if plain.Summary != shaped.Summary || len(shaped.Applied) != 0 {
				t.Fatalf("plain %q shaped %q applied %+v", plain.Summary, shaped.Summary, shaped.Applied)
			}
		})
	})
}
