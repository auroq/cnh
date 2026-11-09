package esindex

import (
	"strings"
	"testing"
)

func TestBulkBody(t *testing.T) {
	body, err := BulkBody("sessions", []Doc{
		{ID: "1", Body: map[string]string{"title": "A"}},
		{ID: "2", Body: map[string]string{"title": "B"}},
	})
	if err != nil {
		t.Fatalf("BulkBody: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(body), "\n"), "\n")

	t.Run("when two docs are rendered", func(t *testing.T) {
		t.Run("it should emit an action line and a source line per doc", func(t *testing.T) {
			if len(lines) != 4 {
				t.Fatalf("got %d lines, want 4: %q", len(lines), lines)
			}
		})
		t.Run("it should put the index and id in the action line", func(t *testing.T) {
			if lines[0] != `{"index":{"_id":"1","_index":"sessions"}}` {
				t.Fatalf("action = %s", lines[0])
			}
		})
		t.Run("it should put the document in the source line", func(t *testing.T) {
			if lines[3] != `{"title":"B"}` {
				t.Fatalf("source = %s", lines[3])
			}
		})
	})
}
