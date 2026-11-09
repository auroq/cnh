package ask

import (
	"os"
	"testing"

	"github.com/auroq/cnh/pkg/venue"
)

func seededNeeds(t *testing.T) []Need {
	t.Helper()
	f, err := os.Open("../../data/venue/needs.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	needs, err := venue.LoadNeeds(f)
	if err != nil {
		t.Fatal(err)
	}
	return needs
}

func TestNeedOf(t *testing.T) {
	needs := seededNeeds(t)
	cases := []struct {
		question string
		want     string
	}{
		{"I'm in a wheelchair", "mobility-accessible"},
		{"I can't do stairs", "mobility-accessible"},
		{"is there a step-free way in", "mobility-accessible"},
		{"mobility accessible route please", "mobility-accessible"},
		{"I need to do my makeup", "private-space"},
		{"where can I fix my makeup", "private-space"},
		{"somewhere to change clothes", "private-space"},
		{"somewhere private", "private-space"},
		{"where can I inject insulin", "medical-private"},
		{"I need an injection", "medical-private"},
		{"take my medication", "medical-private"},
		{"medical", "medical-private"},
		{"I need a tampon", "menstrual"},
		{"a pad", "menstrual"},
		{"period products", "menstrual"},
		{"first aid", "first-aid"},
		{"I'm hurt", "first-aid"},
		{"someone is injured", "first-aid"},
		{"medic", "first-aid"},
		{"emergency exit", "emergency-exit"},
		{"fire exit", "emergency-exit"},
		{"how do I get out", "emergency-exit"},
		{"exit", "emergency-exit"},
		{"I need to sit down", "rest"},
		{"somewhere to rest", "rest"},
		{"I'm overwhelmed", "rest"},
		{"staff", "help"},
		{"I'm lost", "help"},
		{"information", "help"},
	}
	for _, c := range cases {
		t.Run("when the attendee says "+said(c.question), func(t *testing.T) {
			got, ok := NeedOf(c.question, needs)
			t.Run("it should understand the need "+c.want, func(t *testing.T) {
				if !ok || got.ID != c.want {
					t.Fatalf("got %q (matched %v)", got.ID, ok)
				}
			})
		})
	}

	misses := []string{"nearest all-gender restroom", "sign language", "restroom", "", "padded seats"}
	for _, q := range misses {
		t.Run("when the attendee says "+said(q), func(t *testing.T) {
			_, ok := NeedOf(q, needs)
			t.Run("it should not match a need", func(t *testing.T) {
				if ok {
					t.Fatalf("matched a need for %q", q)
				}
			})
		})
	}
}

func TestBestNeed(t *testing.T) {
	private := Need{ID: "private-space"}
	cases := []struct {
		name    string
		matches []NeedMatch
		want    bool
	}{
		{"the store is sure", []NeedMatch{{Need: private, Score: MinNeedScore + 1}}, true},
		{"the store is barely sure", []NeedMatch{{Need: private, Score: MinNeedScore}}, true},
		{"the store is guessing", []NeedMatch{{Need: private, Score: MinNeedScore - 0.1}}, false},
		{"the store found nothing", nil, false},
	}
	for _, c := range cases {
		t.Run("when "+c.name, func(t *testing.T) {
			_, ok := bestNeed(c.matches)
			t.Run("it should take the match only above the threshold", func(t *testing.T) {
				if ok != c.want {
					t.Fatalf("got %v, want %v", ok, c.want)
				}
			})
		})
	}
}

func said(s string) string {
	if s == "" {
		return "nothing"
	}
	return `"` + s + `"`
}
