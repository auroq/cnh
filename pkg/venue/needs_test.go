package venue

import (
	"os"
	"strings"
	"testing"
)

func TestLoadNeeds(t *testing.T) {
	t.Run("when loading the committed needs.yaml", func(t *testing.T) {
		f, err := os.Open("../../data/venue/needs.yaml")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		needs, err := LoadNeeds(f)
		if err != nil {
			t.Fatalf("LoadNeeds: %v", err)
		}
		t.Run("it should load the eight seeded needs", func(t *testing.T) {
			if len(needs) != 8 {
				t.Fatalf("got %d needs", len(needs))
			}
		})
		t.Run("it should give every need at least one kind to offer", func(t *testing.T) {
			for _, n := range needs {
				if len(n.Kinds) == 0 {
					t.Fatalf("%s offers no kind", n.ID)
				}
			}
		})
	})

	broken := []struct {
		name string
		yaml string
	}{
		{"two needs share an id", "needs:\n- {id: a, name: n, phrases: [p], kinds: [k]}\n- {id: a, name: n, phrases: [p], kinds: [k]}\n"},
		{"a need has no phrases", "needs:\n- {id: a, name: n, phrases: [], kinds: [k]}\n"},
		{"a need has no kinds", "needs:\n- {id: a, name: n, phrases: [p]}\n"},
		{"a need has no name", "needs:\n- {id: a, phrases: [p], kinds: [k]}\n"},
	}
	for _, c := range broken {
		t.Run("when "+c.name, func(t *testing.T) {
			_, err := LoadNeeds(strings.NewReader(c.yaml))
			t.Run("it should fail", func(t *testing.T) {
				if err == nil {
					t.Fatal("expected an error")
				}
			})
		})
	}
}
