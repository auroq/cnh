package acceptance

import (
	"strings"
	"testing"

	"github.com/auroq/cnh/testing/helpers"
)

func TestScan(t *testing.T) {
	t.Parallel()

	t.Run("when I open the page from the code on the Grand Ballroom J door", func(t *testing.T) {
		t.Parallel()
		status, body := helpers.Get(t, "/", helpers.Query("from", "room-salt-palace-level-1-grand-ballroom-j"), "text/html")
		t.Run("it should serve the page", func(t *testing.T) {
			if status != 200 {
				t.Fatalf("status %d: %s", status, body)
			}
		})
		t.Run("it should say I scanned the code at Grand Ballroom J", func(t *testing.T) {
			if !strings.Contains(body, "You scanned the code at <strong>Grand Ballroom J</strong>, Salt Palace Level 1.") {
				t.Fatalf("scanned-at sentence missing in %s", body)
			}
		})
		t.Run("it should put the cursor in the question field", func(t *testing.T) {
			if !strings.Contains(body, " autofocus") {
				t.Fatalf("autofocus missing in %s", body)
			}
		})
	})

	t.Run("when I ask the API from the code's id", func(t *testing.T) {
		t.Parallel()
		a := ask(t, "nearest quiet room", "room-salt-palace-level-1-grand-ballroom-j")
		t.Run("it should resolve the id to the Salt Palace", func(t *testing.T) {
			if a.Here.Venue != "Salt Palace" {
				t.Fatalf("here.venue = %q", a.Here.Venue)
			}
		})
	})
}
