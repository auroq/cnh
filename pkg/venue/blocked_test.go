package venue

import (
	"strings"
	"testing"
)

func TestLoadBlockedWords(t *testing.T) {
	t.Run("when the list has comments, blank lines, mixed case and phrases", func(t *testing.T) {
		got, err := LoadBlockedWords(strings.NewReader("# words a note may not contain\n\n  Casino \nkill yourself\n\n"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should keep only the words and phrases, lowercased and trimmed", func(t *testing.T) {
			if len(got) != 2 || got[0] != "casino" || got[1] != "kill yourself" {
				t.Fatalf("got %q", got)
			}
		})
	})
}
