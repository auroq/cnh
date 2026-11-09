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
		{"a must chip", "step-free!", "step-free!"},
		{"a chip both preferred and must", "step-free,step-free!", "step-free!"},
		{"a chip both learned and must", "step-free~,step-free!", "step-free!"},
		{"a must and a learned chip", "quiet-matters~,step-free!", "step-free!,quiet-matters~"},
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

func TestStrength(t *testing.T) {
	prefs := ParsePreferences("step-free!,quiet-matters,all-gender~")

	cases := []struct {
		chip string
		want Strength
	}{
		{"step-free", Must},
		{"quiet-matters", Prefer},
		{"all-gender", Prefer},
		{"dietary", ""},
	}
	for _, c := range cases {
		t.Run("when the profile holds "+c.chip+" as "+quoted(string(c.want)), func(t *testing.T) {
			t.Run("it should report that strength", func(t *testing.T) {
				if got := prefs.StrengthOf(c.chip); got != c.want {
					t.Fatalf("got %q, want %q", got, c.want)
				}
			})
		})
	}

	t.Run("when a learned chip is kept", func(t *testing.T) {
		t.Run("it should become a prefer, never a must", func(t *testing.T) {
			if got := prefs.With("all-gender").StrengthOf("all-gender"); got != Prefer {
				t.Fatalf("got %q", got)
			}
		})
	})

	t.Run("when a prefer is set to must", func(t *testing.T) {
		t.Run("it should encode with a bang", func(t *testing.T) {
			if got := prefs.WithStrength("quiet-matters", Must).String(); got != "step-free!,quiet-matters!,all-gender~" {
				t.Fatalf("got %q", got)
			}
		})
	})

	t.Run("when a must is set back to prefer", func(t *testing.T) {
		t.Run("it should drop the bang", func(t *testing.T) {
			if got := prefs.WithStrength("step-free", Prefer).String(); got != "step-free,quiet-matters,all-gender~" {
				t.Fatalf("got %q", got)
			}
		})
	})

	t.Run("when counting", func(t *testing.T) {
		t.Run("it should count one must", func(t *testing.T) {
			if prefs.Musts() != 1 {
				t.Fatalf("musts %d", prefs.Musts())
			}
		})
		t.Run("it should count two preferences", func(t *testing.T) {
			if prefs.Prefers() != 2 {
				t.Fatalf("prefers %d", prefs.Prefers())
			}
		})
	})
}
