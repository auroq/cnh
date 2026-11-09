package esindex

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	noHits     = `{"hits":{"hits":[]}}`
	crashHit   = `{"hits":{"hits":[{"_source":{"id":"crash","title":"Sign Language Crash Course","room":"Grand Ballroom J","venue":"Salt Palace","level":"Level 1"}}]}}`
	twoTalks   = `{"hits":{"hits":[{"_source":{"id":"crash","title":"Sign Language Crash Course"}},{"_source":{"id":"keynote","title":"Keynote: Welcome"}}]}}`
	exactQuery = `{"term":{"title.keyword":"Sign Language Crash Course"}}`
)

func sequencedES(t *testing.T, responses ...string) (*Store, *[]map[string]any) {
	t.Helper()
	bodies := &[]map[string]any{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &body)
		*bodies = append(*bodies, body)
		i := len(*bodies) - 1
		if i >= len(responses) {
			i = len(responses) - 1
		}
		_, _ = io.WriteString(w, responses[i])
	}))
	t.Cleanup(srv.Close)
	return &Store{Client: &Client{URL: srv.URL, HTTP: srv.Client()}}, bodies
}

func queryOf(body map[string]any) string {
	b, _ := json.Marshal(body["query"])
	return string(b)
}

func TestMatchSession(t *testing.T) {
	ctx := context.Background()

	t.Run("when the pasted title is exactly a session title", func(t *testing.T) {
		store, bodies := sequencedES(t, crashHit)
		got, found, err := store.MatchSession(ctx, "Sign Language Crash Course")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should ask for the exact title first", func(t *testing.T) {
			if q := queryOf((*bodies)[0]); q != exactQuery {
				t.Fatalf("query %s", q)
			}
		})
		t.Run("it should stop after one query", func(t *testing.T) {
			if len(*bodies) != 1 {
				t.Fatalf("%d queries", len(*bodies))
			}
		})
		t.Run("it should return that session", func(t *testing.T) {
			if !found || got.ID != "crash" {
				t.Fatalf("found %v session %+v", found, got)
			}
		})
	})

	t.Run("when the pasted title is close but not exact", func(t *testing.T) {
		store, bodies := sequencedES(t, noHits, crashHit)
		got, found, err := store.MatchSession(ctx, "sign langauge crash course")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should fall back to a fuzzy match that needs most of the words", func(t *testing.T) {
			if q := queryOf((*bodies)[1]); q != `{"match":{"title":{"fuzziness":"AUTO","minimum_should_match":"75%","query":"sign langauge crash course"}}}` {
				t.Fatalf("query %s", q)
			}
		})
		t.Run("it should return the best match", func(t *testing.T) {
			if !found || got.ID != "crash" {
				t.Fatalf("found %v session %+v", found, got)
			}
		})
	})

	t.Run("when nothing on the schedule is close", func(t *testing.T) {
		store, _ := sequencedES(t, noHits)
		_, found, err := store.MatchSession(ctx, "A talk that does not exist")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should say so", func(t *testing.T) {
			if found {
				t.Fatal("found a session")
			}
		})
	})
}

func TestSessionsByID(t *testing.T) {
	t.Run("when two ids are looked up", func(t *testing.T) {
		store, bodies := sequencedES(t, twoTalks)
		got, err := store.SessionsByID(context.Background(), []string{"crash", "keynote"})
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should search the sessions index", func(t *testing.T) {
			if (*bodies)[0]["size"] != float64(2) {
				t.Fatalf("body %v", (*bodies)[0])
			}
		})
		t.Run("it should ask for exactly those ids", func(t *testing.T) {
			if q := queryOf((*bodies)[0]); q != `{"terms":{"id":["crash","keynote"]}}` {
				t.Fatalf("query %s", q)
			}
		})
		t.Run("it should return both sessions", func(t *testing.T) {
			if len(got) != 2 {
				t.Fatalf("got %+v", got)
			}
		})
	})

	t.Run("when no ids are given", func(t *testing.T) {
		store, bodies := sequencedES(t, twoTalks)
		got, err := store.SessionsByID(context.Background(), nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should not ask the store", func(t *testing.T) {
			if len(*bodies) != 0 || len(got) != 0 {
				t.Fatalf("%d queries, %d sessions", len(*bodies), len(got))
			}
		})
	})
}
