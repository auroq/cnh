package ask

import (
	"strings"
	"testing"
)

func TestLearned(t *testing.T) {
	mobility := &Need{ID: "mobility-accessible", Name: "a step-free way around"}

	cases := []struct {
		question string
		need     *Need
		chips    string
		because  string
	}{
		{"nearest elevator", nil, "step-free", "an elevator"},
		{"where is the lift", nil, "step-free", "an elevator"},
		{"step-free entrance", nil, "step-free", "a step-free way in"},
		{"I can't do stairs", mobility, "step-free", "a step-free way in"},
		{"how do I get in without stairs", nil, "step-free", "stairs"},
		{"I use a wheelchair, where is the ramp", nil, "step-free,wheels", "a wheelchair or scooter"},
		{"nearest wheelchair ramp", nil, "step-free,wheels", "a wheelchair or scooter"},
		{"scooter rental", nil, "step-free,wheels", "a wheelchair or scooter"},
		{"nearest quiet room", nil, "quiet-matters", "a quiet room"},
		{"I am overwhelmed", nil, "quiet-matters", "a rest"},
		{"sensory break", nil, "quiet-matters", "a quiet room"},
		{"I need a break", nil, "quiet-matters", "a rest"},
		{"I am overstimulated", nil, "avoid-crowds", "crowds or noise"},
		{"somewhere away from the crowds", nil, "avoid-crowds", "crowds or noise"},
		{"it is too loud here", nil, "avoid-crowds", "crowds or noise"},
		{"nearest all-gender restroom", nil, "all-gender", "an all-gender restroom"},
		{"gender neutral bathroom", nil, "all-gender", "an all-gender restroom"},
		{"talks with sign language", nil, "captions-asl", "captions or sign language"},
		{"are there captions", nil, "captions-asl", "captions or sign language"},
		{"I am hard of hearing, is there an interpreter", nil, "captions-asl", "captions or sign language"},
		{"vegan lunch", nil, "dietary", "dietary needs"},
		{"gluten free food", nil, "dietary", "dietary needs"},
		{"I cannot do stairs", nil, "step-free", "stairs"},
		{"avoid the stairs please", nil, "step-free", "stairs"},
		{"nearest restroom", nil, "", ""},
		{"coat check", nil, "", ""},
		{"where are the stairs", nil, "", ""},
		{"nearest stairs to Hall B", nil, "", ""},
		{"how many steps to the keynote", nil, "", ""},
		{"is there a break between talks", nil, "", ""},
		{"the loud speaker in the keynote", nil, "", ""},
		{"", nil, "", ""},
	}
	for _, c := range cases {
		t.Run("when the question is "+quoted(c.question), func(t *testing.T) {
			got := Learned(c.question, c.need)
			t.Run("it should learn "+quoted(c.chips), func(t *testing.T) {
				if chips := chipsOf(got); chips != c.chips {
					t.Fatalf("learned %q, want %q", chips, c.chips)
				}
			})
			if c.because != "" {
				t.Run("it should say it is because the attendee asked for "+c.because, func(t *testing.T) {
					if !strings.Contains(got[0].Because, c.because) {
						t.Fatalf("because %q", got[0].Because)
					}
				})
			}
		})
	}

	t.Run("when a question implies the same chip twice", func(t *testing.T) {
		got := Learned("elevator or step-free entrance", nil)
		t.Run("it should learn the chip once", func(t *testing.T) {
			if chips := chipsOf(got); chips != "step-free" {
				t.Fatalf("learned %q", chips)
			}
		})
	})
}

func chipsOf(learned []Lesson) string {
	chips := make([]string, 0, len(learned))
	for _, l := range learned {
		chips = append(chips, l.Chip)
	}
	return strings.Join(chips, ",")
}

func quoted(s string) string {
	if s == "" {
		return "nothing"
	}
	return `"` + s + `"`
}
