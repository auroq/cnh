package acceptance

import (
	"strings"
	"testing"
)

func TestLocateInMyWords(t *testing.T) {
	t.Parallel()

	for _, c := range inMyWords {
		t.Run("when I say I am at "+c.From, func(t *testing.T) {
			t.Parallel()
			a := ask(t, "nearest all-gender restroom", c.From)
			t.Run("it should know I am at "+c.Here.Room+" via "+c.Here.Via, func(t *testing.T) {
				want := c.Here
				want.Asked = c.From
				if a.Here != want {
					t.Fatalf("here = %+v", a.Here)
				}
			})
		})
	}

	t.Run("when I am outside on the 100 S side and ask how to get in without stairs", func(t *testing.T) {
		t.Parallel()
		a := ask(t, "how do I get in without stairs", "the 100 S side")
		t.Run("it should offer a step-free entrance", func(t *testing.T) {
			if len(a.Places) == 0 || a.Places[0].Kind != "step-free-entrance" {
				t.Fatalf("places = %+v", a.Places)
			}
		})
		t.Run("it should name the side the door is on", func(t *testing.T) {
			if len(a.Places) == 0 || a.Places[0].Side == "" || !strings.Contains(a.Summary, a.Places[0].Side+" side") {
				t.Fatalf("summary = %q, places = %+v", a.Summary, a.Places)
			}
		})
	})
}
