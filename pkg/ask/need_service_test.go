package ask

import (
	"context"
	"strings"
	"testing"
)

func TestAskUnderstandsNeeds(t *testing.T) {
	ctx := context.Background()
	needs := seededNeeds(t)
	restroom := Place{ID: "agr-sp-l2-255", Kind: "all-gender-restroom", Name: "All-gender restroom outside 255", Venue: "Salt Palace", Level: "Level 2", Near: "Outside room 255", Verified: true}
	nursing := Place{ID: "nursing-room", Kind: "nursing-room", Name: "Private nursing room", Venue: "Salt Palace", Verified: true}
	kindsAsked := func(store *fakeStore) string { return strings.Join(store.kinds, ",") }

	t.Run("when the attendee says they need to fix their makeup", func(t *testing.T) {
		store := &fakeStore{byKind: map[string][]Place{"all-gender-restroom": {restroom}}}
		ans, err := (&Service{Store: store, Needs: needs}).Ask(ctx, "I need to fix my makeup", "")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should offer the first kind the need lists", func(t *testing.T) {
			if kindsAsked(store) != "all-gender-restroom" {
				t.Fatalf("kinds asked %q", kindsAsked(store))
			}
		})
		t.Run("it should not ask the store to guess the need", func(t *testing.T) {
			if len(store.guessed) != 0 {
				t.Fatalf("guessed %q", store.guessed)
			}
		})
		t.Run("it should name the need, then the place, then the note", func(t *testing.T) {
			want := "You asked for somewhere private. All-gender restroom outside 255: Salt Palace, Level 2, Outside room 255. Published by the organiser. Single-stall where known."
			if ans.Summary != want {
				t.Fatalf("\n got %q\nwant %q", ans.Summary, want)
			}
		})
		t.Run("it should report the need it understood", func(t *testing.T) {
			if ans.Need == nil || ans.Need.ID != "private-space" {
				t.Fatalf("need %+v", ans.Need)
			}
		})
	})

	t.Run("when the attendee asks where to inject insulin", func(t *testing.T) {
		t.Run("and the venue has a nursing room", func(t *testing.T) {
			store := &fakeStore{byKind: map[string][]Place{"nursing-room": {nursing}, "all-gender-restroom": {restroom}}}
			ans, err := (&Service{Store: store, Needs: needs}).Ask(ctx, "where can I inject insulin", "")
			if err != nil {
				t.Fatal(err)
			}
			t.Run("it should offer the nursing room", func(t *testing.T) {
				if ans.Places[0].ID != nursing.ID {
					t.Fatalf("places %+v", ans.Places)
				}
			})
			t.Run("it should say first aid staff can help", func(t *testing.T) {
				if !strings.HasSuffix(ans.Summary, " A private room; first aid staff can help.") {
					t.Fatalf("summary %q", ans.Summary)
				}
			})
		})
		t.Run("and the first kinds the need lists have no place", func(t *testing.T) {
			store := &fakeStore{byKind: map[string][]Place{"all-gender-restroom": {restroom}}}
			ans, err := (&Service{Store: store, Needs: needs}).Ask(ctx, "where can I inject insulin", "")
			if err != nil {
				t.Fatal(err)
			}
			t.Run("it should try the kinds in the need's order until one has a place", func(t *testing.T) {
				if kindsAsked(store) != "nursing-room,first-aid,all-gender-restroom" {
					t.Fatalf("kinds asked %q", kindsAsked(store))
				}
			})
			t.Run("it should offer the restroom as a stand-in", func(t *testing.T) {
				if ans.Places[0].ID != restroom.ID {
					t.Fatalf("places %+v", ans.Places)
				}
			})
		})
	})

	t.Run("when the phrasing matches no need word for word", func(t *testing.T) {
		t.Run("and the store is sure which need it means", func(t *testing.T) {
			menstrual := Need{ID: "menstrual", Name: "period products", Kinds: []string{"all-gender-restroom", "info-desk"}, Note: "Ask at the information desk if the restroom has none."}
			store := &fakeStore{needs: []NeedMatch{{Need: menstrual, Score: MinNeedScore + 2}}, byKind: map[string][]Place{"all-gender-restroom": {restroom}}}
			ans, err := (&Service{Store: store, Needs: needs}).Ask(ctx, "it is that time of the month", "")
			if err != nil {
				t.Fatal(err)
			}
			t.Run("it should ask the store which need the question means", func(t *testing.T) {
				if len(store.guessed) != 1 || store.guessed[0] != "it is that time of the month" {
					t.Fatalf("guessed %q", store.guessed)
				}
			})
			t.Run("it should offer the need's first kind", func(t *testing.T) {
				if kindsAsked(store) != "all-gender-restroom" {
					t.Fatalf("kinds asked %q", kindsAsked(store))
				}
			})
			t.Run("it should name the need the store understood", func(t *testing.T) {
				if !strings.HasPrefix(ans.Summary, "You asked for period products. ") {
					t.Fatalf("summary %q", ans.Summary)
				}
			})
		})
		t.Run("and the store is only guessing", func(t *testing.T) {
			store := &fakeStore{needs: []NeedMatch{{Need: Need{ID: "rest"}, Score: MinNeedScore - 1}}, places: []Place{restroom}}
			ans, err := (&Service{Store: store, Needs: needs}).Ask(ctx, "sign language", "")
			if err != nil {
				t.Fatal(err)
			}
			t.Run("it should fall through to the plain place search", func(t *testing.T) {
				if kindsAsked(store) != "" || len(store.kinds) != 1 {
					t.Fatalf("kinds asked %q", store.kinds)
				}
			})
			t.Run("it should not claim to have understood a need", func(t *testing.T) {
				if ans.Need != nil || strings.Contains(ans.Summary, "You asked for") {
					t.Fatalf("need %+v summary %q", ans.Need, ans.Summary)
				}
			})
		})
	})

	t.Run("when the question names a kind of place", func(t *testing.T) {
		store := &fakeStore{places: []Place{restroom}}
		if _, err := (&Service{Store: store, Needs: needs}).Ask(ctx, "nearest restroom", ""); err != nil {
			t.Fatal(err)
		}
		t.Run("it should not ask the store to guess a need", func(t *testing.T) {
			if len(store.guessed) != 0 {
				t.Fatalf("guessed %q", store.guessed)
			}
		})
	})

	t.Run("when a need is understood but no kind it lists has a place", func(t *testing.T) {
		store := &fakeStore{byKind: map[string][]Place{"": {restroom}}}
		ans, err := (&Service{Store: store, Needs: needs}).Ask(ctx, "emergency exit", "")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should end with the plain place search", func(t *testing.T) {
			if kindsAsked(store) != "emergency-exit,step-free-entrance,info-desk," {
				t.Fatalf("kinds asked %q", kindsAsked(store))
			}
		})
		t.Run("it should still name the need it understood", func(t *testing.T) {
			if !strings.HasPrefix(ans.Summary, "You asked for the way out. ") {
				t.Fatalf("summary %q", ans.Summary)
			}
		})
	})
}

func TestAskSpanNamesTheNeed(t *testing.T) {
	t.Run("when the attendee says they are lost", func(t *testing.T) {
		o := observe(t)
		desk := Place{ID: "registration-200-south", Kind: "info-desk", Name: "Registration", Venue: "Salt Palace", Verified: true}
		svc := &Service{Store: &fakeStore{places: []Place{desk}}, Needs: seededNeeds(t), Tracer: o.tracer, Meter: o.meter}
		if _, err := svc.Ask(context.Background(), "I'm lost", ""); err != nil {
			t.Fatal(err)
		}
		t.Run("it should record the need on the ask span", func(t *testing.T) {
			root, ok := o.spans()["ask.Ask"]
			if !ok || !hasAttribute(root.Attributes(), attrNeed.String("help")) {
				t.Fatalf("spans %v", names(o.recorder.Ended()))
			}
		})
	})
}
