package acceptance

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/auroq/cnh/testing/helpers"
)

const roomNotFound = "We could not find a room called"

func page(t *testing.T, question, from string) string {
	t.Helper()
	status, body := helpers.Get(t, "/ask", helpers.Query("q", question, "from", from), "text/html")
	if status != http.StatusOK {
		t.Fatalf("GET /ask: %d: %s", status, body)
	}
	return body
}

func keysOf(t *testing.T, raw json.RawMessage) []string {
	t.Helper()
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("not an object: %v: %s", err, raw)
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func firstOf(t *testing.T, raw json.RawMessage) json.RawMessage {
	t.Helper()
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil || len(items) == 0 {
		t.Fatalf("not a non-empty array: %v: %s", err, raw)
	}
	return items[0]
}

func within(keys, allowed []string) bool {
	ok := map[string]bool{}
	for _, k := range allowed {
		ok[k] = true
	}
	for _, k := range keys {
		if !ok[k] {
			return false
		}
	}
	return true
}

func TestPage(t *testing.T) {
	t.Parallel()

	t.Run("when I ask from a room nobody has heard of", func(t *testing.T) {
		t.Parallel()
		body := page(t, "nearest all-gender restroom", "Narnia")
		t.Run("it should say the room could not be found", func(t *testing.T) {
			if !strings.Contains(body, roomNotFound+" \"Narnia\"") {
				t.Fatalf("page does not say %q", roomNotFound)
			}
		})
		t.Run("it should still give an answer", func(t *testing.T) {
			if !strings.Contains(body, "All-gender restroom outside 255") {
				t.Fatal("page has no answer")
			}
		})
	})

	t.Run("when I ask without saying where I am", func(t *testing.T) {
		t.Parallel()
		body := page(t, "coat and bag check", "")
		t.Run("it should not say a room could not be found", func(t *testing.T) {
			if strings.Contains(body, roomNotFound) {
				t.Fatalf("page says %q", roomNotFound)
			}
		})
	})
}

func TestAPIShape(t *testing.T) {
	t.Parallel()

	t.Run("when I ask the JSON API for the nearest all-gender restroom from 255 BC", func(t *testing.T) {
		t.Parallel()
		var a map[string]json.RawMessage
		helpers.GetJSON(t, "/api/ask", helpers.Query("q", "nearest all-gender restroom", "from", "255 BC"), &a)
		t.Run("it should have question, here, summary, places and sessions", func(t *testing.T) {
			want := []string{"here", "places", "question", "sessions", "summary"}
			if got := keysOf(t, mustRaw(t, a)); strings.Join(got, ",") != strings.Join(want, ",") {
				t.Fatalf("keys = %v", got)
			}
		})
		t.Run("it should describe here with asked, room, venue and level", func(t *testing.T) {
			if got := keysOf(t, a["here"]); !within(got, []string{"asked", "room", "venue", "level"}) {
				t.Fatalf("here keys = %v", got)
			}
		})
		t.Run("it should describe a place with the documented fields", func(t *testing.T) {
			got := keysOf(t, firstOf(t, a["places"]))
			if !within(got, []string{"id", "kind", "name", "venue", "level", "near", "notes", "for", "verified", "source"}) {
				t.Fatalf("place keys = %v", got)
			}
		})
		t.Run("it should describe a session with the documented fields", func(t *testing.T) {
			got := keysOf(t, firstOf(t, a["sessions"]))
			if !within(got, []string{"id", "title", "url", "speakers", "start", "end", "location", "venue", "level", "room"}) {
				t.Fatalf("session keys = %v", got)
			}
		})
	})
}

func mustRaw(t *testing.T, v any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return raw
}

func TestProbes(t *testing.T) {
	t.Parallel()

	for _, path := range []string{"/healthz", "/readyz"} {
		t.Run("when I GET "+path, func(t *testing.T) {
			t.Parallel()
			status, body := helpers.Get(t, path, nil, "text/plain")
			t.Run("it should answer 200", func(t *testing.T) {
				if status != http.StatusOK {
					t.Fatalf("status = %d: %s", status, body)
				}
			})
		})
	}
}
