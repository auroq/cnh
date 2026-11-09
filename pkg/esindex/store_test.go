package esindex

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type captured struct {
	path string
	body map[string]any
}

func fakeES(t *testing.T, response string) (*Store, *captured) {
	t.Helper()
	got := &captured{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.path = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &got.body)
		_, _ = io.WriteString(w, response)
	}))
	t.Cleanup(srv.Close)
	return &Store{Client: &Client{URL: srv.URL, HTTP: srv.Client()}}, got
}

const twoPlaces = `{"hits":{"hits":[
  {"_source":{"id":"hyatt","kind":"all-gender-restroom","name":"Hyatt lobby restroom","venue":"Hyatt Regency","level":"Level 1","verified":true}},
  {"_source":{"id":"sp-255","kind":"all-gender-restroom","name":"Outside 255","venue":"Salt Palace","level":"Level 2","verified":true}}
]}}`

func TestFindPlaces(t *testing.T) {
	t.Run("when asked for all-gender restrooms", func(t *testing.T) {
		store, got := fakeES(t, twoPlaces)
		places, err := store.FindPlaces(context.Background(), "all-gender restroom", "")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should search the places index", func(t *testing.T) {
			if got.path != "/places/_search" {
				t.Fatalf("path %s", got.path)
			}
		})
		t.Run("it should fuse a lexical and a semantic retriever with RRF", func(t *testing.T) {
			retrievers := got.body["retriever"].(map[string]any)["rrf"].(map[string]any)["retrievers"].([]any)
			if len(retrievers) != 2 {
				t.Fatalf("got %d retrievers", len(retrievers))
			}
		})
		t.Run("it should leave rooms out of both retrievers", func(t *testing.T) {
			b, _ := json.Marshal(got.body)
			if n := countOf(string(b), `"must_not":{"term":{"kind":"room"}}`); n != 2 {
				t.Fatalf("rooms excluded %d times, want 2 in %s", n, b)
			}
		})
		t.Run("it should not filter on kind", func(t *testing.T) {
			b, _ := json.Marshal(got.body)
			if n := countOf(string(b), `"filter"`); n != 0 {
				t.Fatalf("filtered %d times, want 0 in %s", n, b)
			}
		})
		t.Run("it should return the hits in Elasticsearch's order", func(t *testing.T) {
			if len(places) != 2 || places[0].ID != "hyatt" || places[1].ID != "sp-255" {
				t.Fatalf("got %+v", places)
			}
		})
	})

	t.Run("when asked for elevators by kind", func(t *testing.T) {
		store, got := fakeES(t, twoPlaces)
		if _, err := store.FindPlaces(context.Background(), "nearest elevator", "elevator"); err != nil {
			t.Fatal(err)
		}
		t.Run("it should keep only elevators in both retrievers", func(t *testing.T) {
			b, _ := json.Marshal(got.body)
			if n := countOf(string(b), `"filter":[{"term":{"kind":"elevator"}}]`); n != 2 {
				t.Fatalf("kind filtered %d times, want 2 in %s", n, b)
			}
		})
		t.Run("it should still leave rooms out of both retrievers", func(t *testing.T) {
			b, _ := json.Marshal(got.body)
			if n := countOf(string(b), `"must_not":{"term":{"kind":"room"}}`); n != 2 {
				t.Fatalf("rooms excluded %d times, want 2 in %s", n, b)
			}
		})
	})
}

func TestLocateRoom(t *testing.T) {
	t.Run("when the room is known", func(t *testing.T) {
		store, got := fakeES(t, `{"hits":{"hits":[{"_source":{"kind":"room","name":"255 BC","venue":"Salt Palace","level":"Level 2"}}]}}`)
		here, err := store.LocateRoom(context.Background(), "255 bc")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should only look at rooms", func(t *testing.T) {
			b, _ := json.Marshal(got.body["query"])
			if countOf(string(b), `"filter":[{"term":{"kind":"room"}}]`) != 1 {
				t.Fatalf("query %s", b)
			}
		})
		t.Run("it should return the room with its venue and level", func(t *testing.T) {
			if here.Room != "255 BC" || here.Venue != "Salt Palace" || here.Level != "Level 2" {
				t.Fatalf("got %+v", here)
			}
		})
	})

	t.Run("when nothing matches", func(t *testing.T) {
		store, _ := fakeES(t, `{"hits":{"hits":[]}}`)
		here, err := store.LocateRoom(context.Background(), "the moon")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should return an empty location", func(t *testing.T) {
			if here.Room != "" || here.Venue != "" {
				t.Fatalf("got %+v", here)
			}
		})
	})
}

func countOf(s, sub string) int {
	n := 0
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			n++
		}
	}
	return n
}
