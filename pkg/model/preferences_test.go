package model

import "testing"

func TestParsePreferences(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"nothing", "", ""},
		{"one chip", "step-free", "step-free"},
		{"two chips out of catalogue order", "wheels,step-free", "step-free,wheels"},
		{"a chip twice", "wheels,wheels", "wheels"},
		{"an unknown chip", "step-free,diagnosis", "step-free"},
		{"spaces around the ids", " step-free , wheels ", "step-free,wheels"},
		{"only unknown chips", "x,y", ""},
		{"a learned chip", "quiet-matters~", "quiet-matters~"},
		{"a chip both learned and chosen", "quiet-matters~,quiet-matters", "quiet-matters"},
		{"a chosen and a learned chip", "wheels,step-free~", "step-free~,wheels"},
	}
	for _, c := range cases {
		t.Run("when prefs say "+c.name, func(t *testing.T) {
			got := ParsePreferences(c.in)
			t.Run("it should keep "+quoted(c.want), func(t *testing.T) {
				if got.String() != c.want {
					t.Fatalf("got %q, want %q", got.String(), c.want)
				}
			})
		})
	}
}

func TestPreferences(t *testing.T) {
	prefs := ParsePreferences("step-free,wheels")

	t.Run("when a chip is in the profile", func(t *testing.T) {
		t.Run("it should be reported as present", func(t *testing.T) {
			if !prefs.Has("wheels") {
				t.Fatal("wheels missing")
			}
		})
		t.Run("and it is removed", func(t *testing.T) {
			t.Run("it should leave the other chips", func(t *testing.T) {
				if got := prefs.Without("wheels").String(); got != "step-free" {
					t.Fatalf("got %q", got)
				}
			})
		})
	})

	t.Run("when a chip is not in the profile", func(t *testing.T) {
		t.Run("it should be reported as absent", func(t *testing.T) {
			if prefs.Has("dietary") {
				t.Fatal("dietary present")
			}
		})
		t.Run("and it is added", func(t *testing.T) {
			t.Run("it should land in catalogue order", func(t *testing.T) {
				if got := prefs.With("avoid-crowds").String(); got != "step-free,avoid-crowds,wheels" {
					t.Fatalf("got %q", got)
				}
			})
		})
		t.Run("and an unknown id is added", func(t *testing.T) {
			t.Run("it should be ignored", func(t *testing.T) {
				if got := prefs.With("diagnosis").String(); got != "step-free,wheels" {
					t.Fatalf("got %q", got)
				}
			})
		})
	})

	t.Run("when a learned chip is chosen", func(t *testing.T) {
		learned := ParsePreferences("quiet-matters~")
		t.Run("it should become a chosen chip", func(t *testing.T) {
			if got := learned.With("quiet-matters").String(); got != "quiet-matters" {
				t.Fatalf("got %q", got)
			}
		})
		t.Run("it should list its id without the mark", func(t *testing.T) {
			if got := learned.IDs(); len(got) != 1 || got[0] != "quiet-matters" {
				t.Fatalf("got %q", got)
			}
		})
	})

	t.Run("when the profile is empty", func(t *testing.T) {
		t.Run("it should print as nothing", func(t *testing.T) {
			if got := (Preferences)(nil).String(); got != "" {
				t.Fatalf("got %q", got)
			}
		})
	})
}

func quoted(s string) string {
	if s == "" {
		return "nothing"
	}
	return `"` + s + `"`
}
