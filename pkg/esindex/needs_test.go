package esindex

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/auroq/cnh/pkg/venue"
)

const twoNeeds = `{"hits":{"hits":[
  {"_score":12.5,"_source":{"id":"medical-private","name":"a private place for medication","phrases":["insulin","inject"],"kinds":["nursing-room","first-aid"],"note":"A private room; first aid staff can help."}},
  {"_score":3.1,"_source":{"id":"first-aid","name":"first aid","phrases":["hurt"],"kinds":["first-aid","info-desk"]}}
]}}`

func TestFindNeeds(t *testing.T) {
	t.Run("when asked where to take medication", func(t *testing.T) {
		store, got := fakeES(t, twoNeeds)
		store.NeedsIndex = "needs-test"
		matches, err := store.FindNeeds(context.Background(), "where can I take my meds in private")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should search the needs index", func(t *testing.T) {
			if got.path != "/needs-test/_search" {
				t.Fatalf("path %s", got.path)
			}
		})
		t.Run("it should ask the semantic_text field, not BM25", func(t *testing.T) {
			b, _ := json.Marshal(got.body["query"])
			if string(b) != `{"semantic":{"field":"content","query":"where can I take my meds in private"}}` {
				t.Fatalf("query %s", b)
			}
		})
		t.Run("it should return the needs best first with their scores", func(t *testing.T) {
			if len(matches) != 2 || matches[0].Need.ID != "medical-private" || matches[0].Score != 12.5 || matches[1].Score != 3.1 {
				t.Fatalf("got %+v", matches)
			}
		})
		t.Run("it should carry the need's kinds and note through", func(t *testing.T) {
			n := matches[0].Need
			if len(n.Kinds) != 2 || n.Kinds[0] != "nursing-room" || n.Note != "A private room; first aid staff can help." {
				t.Fatalf("got %+v", n)
			}
		})
	})
}

func TestIndexNeeds(t *testing.T) {
	t.Run("when needs are indexed", func(t *testing.T) {
		var bulk []byte
		var paths []string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			paths = append(paths, r.Method+" "+r.URL.Path)
			switch r.URL.Path {
			case "/_bulk":
				bulk, _ = io.ReadAll(r.Body)
				_, _ = io.WriteString(w, `{"errors":false,"items":[]}`)
			case "/needs/_count":
				_, _ = io.WriteString(w, `{"count":1}`)
			default:
				_, _ = io.WriteString(w, `{}`)
			}
		}))
		t.Cleanup(srv.Close)
		now := time.Date(2026, 11, 9, 20, 0, 0, 0, time.UTC)
		x := &Indexer{Client: &Client{URL: srv.URL, HTTP: srv.Client()}, Now: func() time.Time { return now }}
		need := venue.Need{ID: "private-space", Name: "somewhere private", Phrases: []string{"fix my makeup", "change clothes"}, Kinds: []string{"all-gender-restroom", "quiet-room"}, Note: "Single-stall where known."}
		res, err := x.IndexNeeds(context.Background(), []venue.Need{need})
		if err != nil {
			t.Fatal(err)
		}
		var doc map[string]any
		lines := splitLines(bulk)
		if len(lines) < 2 {
			t.Fatalf("bulk body %q", bulk)
		}
		if err := json.Unmarshal(lines[1], &doc); err != nil {
			t.Fatal(err)
		}
		t.Run("it should write to the needs index by default", func(t *testing.T) {
			if paths[len(paths)-2] != "POST /needs/_refresh" {
				t.Fatalf("paths %v", paths)
			}
		})
		t.Run("it should build the semantic content from the name and the phrases", func(t *testing.T) {
			if doc["content"] != "somewhere private. fix my makeup. change clothes" {
				t.Fatalf("content %q", doc["content"])
			}
		})
		t.Run("it should keep the kinds in the order to offer them", func(t *testing.T) {
			kinds, _ := doc["kinds"].([]any)
			if len(kinds) != 2 || kinds[0] != "all-gender-restroom" {
				t.Fatalf("kinds %v", doc["kinds"])
			}
		})
		t.Run("it should stamp when it was ingested", func(t *testing.T) {
			if doc["ingested_at"] != "2026-11-09T20:00:00Z" {
				t.Fatalf("ingested_at %v", doc["ingested_at"])
			}
		})
		t.Run("it should report what the index holds", func(t *testing.T) {
			if res.Sent != 1 || res.Count != 1 || res.Failed != 0 {
				t.Fatalf("got %+v", res)
			}
		})
	})
}

func splitLines(b []byte) [][]byte {
	var out [][]byte
	start := 0
	for i, c := range b {
		if c == '\n' {
			out = append(out, b[start:i])
			start = i + 1
		}
	}
	if start < len(b) {
		out = append(out, b[start:])
	}
	return out
}
