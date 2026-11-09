package esindex

import (
	"testing"
	"time"

	"github.com/auroq/cnh/pkg/schedule"
	"github.com/auroq/cnh/pkg/venue"
)

func TestSessionDoc(t *testing.T) {
	now := time.Date(2026, 11, 9, 18, 0, 0, 0, time.UTC)
	cases := []struct {
		name    string
		session schedule.Session
		content string
	}{
		{"a described session", schedule.Session{ID: "1", Title: "Sign Language Crash Course", Description: "Learn ten signs."}, "Sign Language Crash Course\n\nLearn ten signs."},
		{"a session without a description", schedule.Session{ID: "2", Title: "Coat + Bag Check"}, "Coat + Bag Check"},
	}
	for _, c := range cases {
		t.Run("when indexing "+c.name, func(t *testing.T) {
			doc := sessionDoc(c.session, now)
			t.Run("it should build the semantic content from title and description", func(t *testing.T) {
				if doc["content"] != c.content {
					t.Fatalf("content %q", doc["content"])
				}
			})
			t.Run("it should stamp when it was ingested", func(t *testing.T) {
				if doc["ingested_at"] != now {
					t.Fatalf("ingested_at %v", doc["ingested_at"])
				}
			})
		})
	}
}

func TestPlaceDoc(t *testing.T) {
	t.Run("when indexing a curated place", func(t *testing.T) {
		p := venue.Place{ID: "quiet-260a", Kind: "quiet-room", Name: "Quiet room 260A", Venue: "Salt Palace", Level: "Level 2", Verified: true, Source: "organiser"}
		doc := placeDoc(p, time.Time{})
		t.Run("it should carry the verified flag and source through", func(t *testing.T) {
			if doc["verified"] != true || doc["source"] != "organiser" {
				t.Fatalf("got %v", doc)
			}
		})
		t.Run("it should use the place's own content text for the semantic field", func(t *testing.T) {
			if doc["content"] != placeContent(p) {
				t.Fatalf("content %q", doc["content"])
			}
		})
	})
}
