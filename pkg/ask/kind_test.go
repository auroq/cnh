package ask

import "testing"

func TestKindOf(t *testing.T) {
	cases := []struct {
		question string
		want     string
	}{
		{"nearest elevator", "elevator"},
		{"where is the lift", "elevator"},
		{"nearest all-gender restroom", "all-gender-restroom"},
		{"closest toilet", "all-gender-restroom"},
		{"bathroom", "all-gender-restroom"},
		{"nearest quiet room", "quiet-room"},
		{"somewhere sensory friendly", "quiet-room"},
		{"prayer room", "prayer-room"},
		{"where can I pray", "prayer-room"},
		{"quiet place to pray", "prayer-room"},
		{"nursing room", "nursing-room"},
		{"lactation room", "nursing-room"},
		{"step-free entrance", "step-free-entrance"},
		{"accessible entrance", "step-free-entrance"},
		{"I need help renting a wheelchair", "mobility-equipment"},
		{"accessible parking info", "parking"},
		{"community hub", "community-hub"},
		{"coat and bag check", "coat-check"},
		{"where do I leave my coat", "coat-check"},
		{"registration and help desk", "info-desk"},
		{"info desk", "info-desk"},
		{"help me find a quiet room", "quiet-room"},
		{"help", ""},
		{"a quiet corner to sit", ""},
		{"sign language", ""},
		{"uplifting talks about bagels", ""},
		{"", ""},
	}
	for _, c := range cases {
		t.Run("when the question is "+quoteOrEmpty(c.question), func(t *testing.T) {
			t.Run("it should infer "+quoteOrEmpty(c.want), func(t *testing.T) {
				if got := KindOf(c.question); got != c.want {
					t.Fatalf("got %q, want %q", got, c.want)
				}
			})
		})
	}
}

func quoteOrEmpty(s string) string {
	if s == "" {
		return "nothing"
	}
	return `"` + s + `"`
}
