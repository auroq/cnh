package model

import "testing"

func TestLearnedPreferences(t *testing.T) {
	cases := []struct {
		name  string
		have  string
		learn string
		want  string
	}{
		{"the profile is empty", "", "step-free", "step-free~"},
		{"the chip is already chosen", "step-free", "step-free", "step-free"},
		{"the chip is already learned", "step-free~", "step-free", "step-free~"},
		{"another chip is chosen", "wheels", "step-free", "step-free~,wheels"},
		{"the chip is unknown", "wheels", "diagnosis", "wheels"},
	}
	for _, c := range cases {
		t.Run("when "+c.name+" and "+c.learn+" is learned", func(t *testing.T) {
			got := ParsePreferences(c.have).WithLearned(c.learn)
			t.Run("it should hold "+quoted(c.want), func(t *testing.T) {
				if got.String() != c.want {
					t.Fatalf("got %q, want %q", got.String(), c.want)
				}
			})
		})
	}

	t.Run("when a profile holds a chosen and a learned chip", func(t *testing.T) {
		prefs := ParsePreferences("step-free~,wheels")
		sources := map[string]bool{"step-free": true, "wheels": false, "dietary": false}
		for id, learned := range sources {
			t.Run("it should report "+id+" learned: "+boolWord(learned), func(t *testing.T) {
				if prefs.IsLearned(id) != learned {
					t.Fatalf("IsLearned(%q) = %v", id, !learned)
				}
			})
		}
	})
}

func boolWord(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
