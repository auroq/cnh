package web

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/auroq/cnh/pkg/ask"
)

type fakeAsker struct {
	answer ask.Answer
	err    error
}

func (f fakeAsker) Ask(context.Context, string, string) (ask.Answer, error) { return f.answer, f.err }
func (f fakeAsker) Locate(context.Context, string) (ask.Here, error)        { return f.answer.Here, f.err }
func (f fakeAsker) Ready(context.Context) error                             { return f.err }

func get(t *testing.T, a Asker, target string) (*http.Response, string) {
	t.Helper()
	srv := httptest.NewServer(New(a, slog.New(slog.NewTextHandler(io.Discard, nil))).Handler())
	t.Cleanup(srv.Close)
	res, err := http.Get(srv.URL + target)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { res.Body.Close() })
	body, _ := io.ReadAll(res.Body)
	return res, string(body)
}

func TestServer(t *testing.T) {
	answered := fakeAsker{answer: ask.Answer{
		Summary: "Quiet room 260A: Salt Palace, Level 2. Same building, you are on Level 1 and it is on Level 2. Published by the organiser.",
		Here:    ask.Here{Asked: "Grand Ballroom J", Room: "Grand Ballroom J", Venue: "Salt Palace", Level: "Level 1"},
		Places:  []ask.Place{{Name: "Quiet room 260A", Venue: "Salt Palace", Level: "Level 2", Verified: true}},
	}}
	down := fakeAsker{err: errors.New("no elasticsearch")}

	t.Run("when the page is asked for a quiet room from Grand Ballroom J", func(t *testing.T) {
		res, body := get(t, answered, "/ask?q=quiet+room&from=Grand+Ballroom+J")
		t.Run("it should answer with HTML", func(t *testing.T) {
			if res.StatusCode != 200 || !strings.HasPrefix(res.Header.Get("Content-Type"), "text/html") {
				t.Fatalf("%d %s", res.StatusCode, res.Header.Get("Content-Type"))
			}
		})
		t.Run("it should put the summary in the title so a screen reader hears it first", func(t *testing.T) {
			if !strings.Contains(body, "<title>"+answered.answer.Summary+" | Wide Door</title>") {
				t.Fatalf("title missing in %s", body)
			}
		})
		t.Run("it should put the summary in the live region", func(t *testing.T) {
			if !strings.Contains(body, `<p class="summary">`+answered.answer.Summary+"</p>") {
				t.Fatalf("summary missing in %s", body)
			}
		})
	})

	t.Run("when the API is asked the same question", func(t *testing.T) {
		res, body := get(t, answered, "/api/ask?q=quiet+room&from=Grand+Ballroom+J")
		var got ask.Answer
		if err := json.Unmarshal([]byte(body), &got); err != nil {
			t.Fatal(err)
		}
		t.Run("it should answer with JSON", func(t *testing.T) {
			if res.StatusCode != 200 || res.Header.Get("Content-Type") != "application/json" {
				t.Fatalf("%d %s", res.StatusCode, res.Header.Get("Content-Type"))
			}
		})
		t.Run("it should return the whole answer", func(t *testing.T) {
			if got.Summary != answered.answer.Summary || len(got.Places) != 1 {
				t.Fatalf("got %+v", got)
			}
		})
	})

	t.Run("when the search is down", func(t *testing.T) {
		cases := []struct {
			target string
			status int
			text   string
		}{
			{"/ask?q=quiet+room", 200, searchDown},
			{"/api/ask?q=quiet+room", http.StatusBadGateway, "search unavailable"},
			{"/readyz", http.StatusServiceUnavailable, "search unavailable"},
		}
		for _, c := range cases {
			t.Run("and "+c.target+" is requested", func(t *testing.T) {
				res, body := get(t, down, c.target)
				t.Run("it should say so without crashing", func(t *testing.T) {
					if res.StatusCode != c.status || !strings.Contains(body, c.text) {
						t.Fatalf("%d %q", res.StatusCode, body)
					}
				})
			})
		}
	})

	t.Run("when the liveness probe is hit", func(t *testing.T) {
		res, _ := get(t, down, "/healthz")
		t.Run("it should be ok even with the search down", func(t *testing.T) {
			if res.StatusCode != 200 {
				t.Fatalf("%d", res.StatusCode)
			}
		})
	})
}
