package esindex

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/auroq/cnh/pkg/schedule"
)

func TestIndexSessionsWrites(t *testing.T) {
	t.Run("when sessions are indexed", func(t *testing.T) {
		var paths []string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			paths = append(paths, r.Method+" "+r.URL.RequestURI())
			switch {
			case r.URL.Path == "/_bulk":
				_, _ = io.WriteString(w, `{"errors":false,"items":[]}`)
			case r.URL.Path == "/sessions/_count":
				_, _ = io.WriteString(w, `{"count":2}`)
			default:
				_, _ = io.WriteString(w, `{}`)
			}
		}))
		t.Cleanup(srv.Close)
		x := &Indexer{Client: &Client{URL: srv.URL, HTTP: srv.Client()}, Batch: 1}
		res, err := x.IndexSessions(context.Background(), []schedule.Session{{ID: "1"}, {ID: "2"}})
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should send each bulk batch without waiting for a refresh", func(t *testing.T) {
			if paths[0] != "POST /_bulk?refresh=false" || paths[1] != "POST /_bulk?refresh=false" {
				t.Fatalf("got %v", paths)
			}
		})
		t.Run("it should refresh the index once, after every batch", func(t *testing.T) {
			if paths[2] != "POST /sessions/_refresh" {
				t.Fatalf("got %v", paths)
			}
		})
		t.Run("it should report what was sent and what the index holds", func(t *testing.T) {
			if res.Sent != 2 || res.Count != 2 || res.Failed != 0 {
				t.Fatalf("got %+v", res)
			}
		})
	})
}
