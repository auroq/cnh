package esindex

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

const registrationHit = `{"hits":{"hits":[{"_source":{"id":"registration-200-south","kind":"info-desk","name":"Registration and badge pick-up","venue":"Salt Palace","side":"200 South"}}]}}`

const sessionHit = `{"hits":{"hits":[{"_source":{"id":"1400001","title":"Sign Language Crash Course","venue":"Salt Palace","level":"Level 1","room":"Grand Ballroom J"}}]}}`

func TestMeaningful(t *testing.T) {
	cases := []struct{ in, want string }{
		{"the 100 S side", "100 s side"},
		{"I'm at where everyone registered", "everyone registered"},
		{"I just got out of Sign Language Crash Course", "sign language crash course"},
		{"the moon", "moon"},
		{"the", ""},
	}
	for _, c := range cases {
		t.Run("when the attendee says "+c.in, func(t *testing.T) {
			t.Run("it should keep only the words that mean something", func(t *testing.T) {
				if got := meaningful(c.in); got != c.want {
					t.Fatalf("got %q, want %q", got, c.want)
				}
			})
		})
	}
}

func TestLocateAlias(t *testing.T) {
	t.Run("when a phrase people use is looked up", func(t *testing.T) {
		store, got := fakeES(t, registrationHit)
		here, err := store.LocateAlias(context.Background(), "where everyone registered")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should require every meaningful word in the aliases", func(t *testing.T) {
			b, _ := json.Marshal(got.body["query"])
			if string(b) != `{"match":{"aliases":{"operator":"and","query":"everyone registered"}}}` {
				t.Fatalf("query %s", b)
			}
		})
		t.Run("it should return the place with its side", func(t *testing.T) {
			if here.Room != "Registration and badge pick-up" || here.Venue != "Salt Palace" || here.Side != "200 South" {
				t.Fatalf("here %+v", here)
			}
		})
	})

	t.Run("when only filler words were said", func(t *testing.T) {
		store, got := fakeES(t, registrationHit)
		here, err := store.LocateAlias(context.Background(), "the")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should not search at all", func(t *testing.T) {
			if got.path != "" || here.Venue != "" {
				t.Fatalf("path %q, here %+v", got.path, here)
			}
		})
	})
}

func TestLocateSession(t *testing.T) {
	t.Run("when the attendee says which session they just left", func(t *testing.T) {
		store, got := fakeES(t, sessionHit)
		now := time.Date(2026, 11, 11, 21, 20, 0, 0, time.UTC)
		here, err := store.LocateSession(context.Background(), "I just got out of Sign Language Crash Course", now)
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should search the sessions index", func(t *testing.T) {
			if got.path != "/sessions/_search" {
				t.Fatalf("path %s", got.path)
			}
		})
		t.Run("it should require every meaningful word in the title", func(t *testing.T) {
			b, _ := json.Marshal(got.body["query"].(map[string]any)["bool"].(map[string]any)["must"])
			if string(b) != `[{"match":{"title":{"operator":"and","query":"sign language crash course"}}}]` {
				t.Fatalf("must %s", b)
			}
		})
		t.Run("it should prefer the session that ended closest to now", func(t *testing.T) {
			b, _ := json.Marshal(got.body["query"].(map[string]any)["bool"].(map[string]any)["should"])
			if string(b) != `[{"distance_feature":{"field":"end","origin":"2026-11-11T21:20:00Z","pivot":"30m"}}]` {
				t.Fatalf("should %s", b)
			}
		})
		t.Run("it should return the session's room, venue and level", func(t *testing.T) {
			if here.Room != "Grand Ballroom J" || here.Venue != "Salt Palace" || here.Level != "Level 1" {
				t.Fatalf("here %+v", here)
			}
		})
	})
}

func TestLocateRoomFuzzy(t *testing.T) {
	t.Run("when a room name is misspelt", func(t *testing.T) {
		store, got := fakeES(t, `{"hits":{"hits":[{"_source":{"kind":"room","name":"255 BC","venue":"Salt Palace","level":"Level 2"}}]}}`)
		here, err := store.LocateRoomFuzzy(context.Background(), "255 bcc")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should only look at rooms", func(t *testing.T) {
			b, _ := json.Marshal(got.body["query"])
			if countOf(string(b), `"filter":[{"term":{"kind":"room"}}]`) != 1 {
				t.Fatalf("query %s", b)
			}
		})
		t.Run("it should allow fuzzy matches on the name", func(t *testing.T) {
			b, _ := json.Marshal(got.body["query"])
			if countOf(string(b), `"fuzziness":"AUTO"`) != 1 {
				t.Fatalf("query %s", b)
			}
		})
		t.Run("it should return the room", func(t *testing.T) {
			if here.Room != "255 BC" {
				t.Fatalf("here %+v", here)
			}
		})
	})
}

func TestLocateNearby(t *testing.T) {
	t.Run("when words describe a place loosely", func(t *testing.T) {
		store, got := fakeES(t, registrationHit)
		if _, err := store.LocateNearby(context.Background(), "the place with the long badge queue"); err != nil {
			t.Fatal(err)
		}
		t.Run("it should ask the semantic field", func(t *testing.T) {
			b, _ := json.Marshal(got.body["query"])
			if string(b) != `{"semantic":{"field":"content","query":"the place with the long badge queue"}}` {
				t.Fatalf("query %s", b)
			}
		})
		t.Run("it should only accept a confident match", func(t *testing.T) {
			if got.body["min_score"] != nearbyMinScore {
				t.Fatalf("min_score %v", got.body["min_score"])
			}
		})
	})
}
