package esindex

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/auroq/cnh/pkg/model"
)

type capturedCall struct {
	method string
	path   string
	query  string
	body   map[string]any
}

func fakeReportsES(t *testing.T, response string) (*Reports, *capturedCall) {
	t.Helper()
	got := &capturedCall{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.method, got.path, got.query = r.Method, r.URL.Path, r.URL.RawQuery
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &got.body)
		_, _ = io.WriteString(w, response)
	}))
	t.Cleanup(srv.Close)
	return &Reports{Client: &Client{URL: srv.URL, HTTP: srv.Client()}}, got
}

func TestReports(t *testing.T) {
	ctx := context.Background()
	at := time.Date(2026, 11, 9, 21, 52, 0, 0, time.UTC)

	t.Run("when a report is saved", func(t *testing.T) {
		store, got := fakeReportsES(t, `{"result":"created"}`)
		err := store.Save(ctx, model.Report{ID: "abc", PlaceID: "elevators-hyatt", Verdict: model.VerdictProblem, Note: "out of order", At: at, Token: "tok", Source: "attendee"})
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should put the document under its id in the reports index", func(t *testing.T) {
			if got.method != http.MethodPut || got.path != "/reports/_doc/abc" {
				t.Fatalf("%s %s", got.method, got.path)
			}
		})
		t.Run("it should wait for the refresh so the next answer sees it", func(t *testing.T) {
			if got.query != "refresh=wait_for" {
				t.Fatalf("query %s", got.query)
			}
		})
		t.Run("it should write the place id and verdict as the model names them", func(t *testing.T) {
			if got.body["place_id"] != "elevators-hyatt" || got.body["verdict"] != "problem" {
				t.Fatalf("body %v", got.body)
			}
		})
	})

	t.Run("when reports since a time are asked for two places", func(t *testing.T) {
		store, got := fakeReportsES(t, `{"hits":{"hits":[{"_source":{"id":"1","place_id":"a","verdict":"works","at":"2026-11-09T21:00:00Z","token":"x"}}]}}`)
		reps, err := store.Since(ctx, []string{"a", "b"}, at)
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should search the reports index", func(t *testing.T) {
			if got.path != "/reports/_search" {
				t.Fatalf("path %s", got.path)
			}
		})
		t.Run("it should filter on those place ids and the time", func(t *testing.T) {
			b, _ := json.Marshal(got.body["query"])
			want := `{"bool":{"filter":[{"terms":{"place_id":["a","b"]}},{"range":{"at":{"gte":"2026-11-09T21:52:00Z"}}}]}}`
			if string(b) != want {
				t.Fatalf("query %s", b)
			}
		})
		t.Run("it should return the reports as the model", func(t *testing.T) {
			if len(reps) != 1 || reps[0].PlaceID != "a" || reps[0].Verdict != model.VerdictWorks {
				t.Fatalf("reports %+v", reps)
			}
		})
	})

	t.Run("when a report is hidden", func(t *testing.T) {
		store, got := fakeReportsES(t, `{"result":"updated"}`)
		if err := store.Hide(ctx, "abc"); err != nil {
			t.Fatal(err)
		}
		t.Run("it should update only the hidden flag", func(t *testing.T) {
			b, _ := json.Marshal(got.body)
			if got.path != "/reports/_update/abc" || string(b) != `{"doc":{"hidden":true}}` {
				t.Fatalf("%s %s", got.path, b)
			}
		})
	})

	kinds := []struct {
		name     string
		response string
		want     string
	}{
		{"a known place", `{"hits":{"hits":[{"_source":{"kind":"elevator"}}]}}`, "elevator"},
		{"an unknown place", `{"hits":{"hits":[]}}`, ""},
	}
	for _, c := range kinds {
		t.Run("when the kind of "+c.name+" is asked", func(t *testing.T) {
			store, got := fakeReportsES(t, c.response)
			kind, err := store.PlaceKind(ctx, "x")
			if err != nil {
				t.Fatal(err)
			}
			t.Run("it should ask the places index for that id", func(t *testing.T) {
				b, _ := json.Marshal(got.body["query"])
				if got.path != "/places/_search" || string(b) != `{"term":{"id":"x"}}` {
					t.Fatalf("%s %s", got.path, b)
				}
			})
			t.Run("it should answer "+c.want, func(t *testing.T) {
				if kind != c.want {
					t.Fatalf("kind %q", kind)
				}
			})
		})
	}
}
