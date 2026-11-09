package route

import (
	"strings"
	"testing"

	"github.com/auroq/cnh/pkg/model"
)

const fixture = `
nodes:
  - {id: l1, name: Level 1 lobby}
  - {id: l2, name: Level 2 lobby}
  - {id: lift, name: the elevators by registration}
  - {id: showcase, name: the Solutions Showcase}
  - {id: hub, name: the Community Hub}
  - {id: island, name: a room with no way in}
edges:
  - {from: l1, to: l2, kind: stairs, step_free: false, verified: false, source: plan}
  - {from: l1, to: lift, kind: corridor, step_free: true, verified: true, source: organiser}
  - {from: lift, to: l2, kind: elevator, step_free: true, verified: true, source: organiser}
  - {from: l1, to: showcase, kind: corridor, step_free: true, through_zone: showcase, amenities: [restroom], verified: true, source: organiser}
  - {from: showcase, to: hub, kind: corridor, step_free: true, through_zone: showcase, verified: true, source: organiser}
  - {from: l2, to: hub, kind: corridor, step_free: true, verified: false, source: plan}
`

func load(t *testing.T, text string) *Graph {
	t.Helper()
	g, err := Load(strings.NewReader(text))
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func kindsOf(r model.Route) string {
	var k []string
	for _, s := range r.Steps {
		k = append(k, s.Kind)
	}
	return strings.Join(k, ",")
}

func TestShortest(t *testing.T) {
	g := load(t, fixture)

	t.Run("when stairs are allowed", func(t *testing.T) {
		r, ok := g.Shortest("l1", "l2", model.Constraints{})
		t.Run("it should find a route", func(t *testing.T) {
			if !ok {
				t.Fatal("no route")
			}
		})
		t.Run("it should take the one-edge way", func(t *testing.T) {
			if len(r.Steps) != 1 {
				t.Fatalf("got %d steps", len(r.Steps))
			}
		})
	})

	t.Run("when no stairs is set", func(t *testing.T) {
		stairsOnly := load(t, `
nodes:
  - {id: a, name: A}
  - {id: b, name: B}
edges:
  - {from: a, to: b, kind: stairs, step_free: false, verified: true, source: s}
`)
		t.Run("and an elevator path exists", func(t *testing.T) {
			r, ok := g.Shortest("l1", "hub", model.Constraints{NoStairs: true, AvoidZones: []string{"showcase"}})
			t.Run("it should find a route", func(t *testing.T) {
				if !ok {
					t.Fatal("no route")
				}
			})
			t.Run("it should use the elevator and not the stairs", func(t *testing.T) {
				if got := kindsOf(r); got != "corridor,elevator,corridor" {
					t.Fatalf("got %s", got)
				}
			})
			t.Run("it should say which elevator in plain words", func(t *testing.T) {
				if r.Steps[0].Text != "Follow the corridor to the elevators by registration." {
					t.Fatalf("got %q", r.Steps[0].Text)
				}
			})
			t.Run("it should list no obstacles", func(t *testing.T) {
				if len(r.Obstacles) != 0 {
					t.Fatalf("got %v", r.Obstacles)
				}
			})
		})
		t.Run("and only stairs connect the two", func(t *testing.T) {
			_, ok := stairsOnly.Shortest("a", "b", model.Constraints{NoStairs: true})
			t.Run("it should say there is no route", func(t *testing.T) {
				if ok {
					t.Fatal("found a route over stairs")
				}
			})
		})
	})

	t.Run("when the showcase is to be avoided", func(t *testing.T) {
		t.Run("and a longer way exists", func(t *testing.T) {
			r, _ := g.Shortest("l1", "hub", model.Constraints{AvoidZones: []string{"showcase"}})
			t.Run("it should take the longer way", func(t *testing.T) {
				if len(r.Steps) != 2 || r.Steps[0].Kind != "stairs" {
					t.Fatalf("got %s", kindsOf(r))
				}
			})
			t.Run("it should report the zone avoided", func(t *testing.T) {
				if len(r.Avoided) != 1 || r.Avoided[0] != "showcase" {
					t.Fatalf("got %v", r.Avoided)
				}
			})
			t.Run("it should mark the stairs as an obstacle", func(t *testing.T) {
				if len(r.Obstacles) != 1 || r.Obstacles[0].Kind != "stairs" {
					t.Fatalf("got %v", r.Obstacles)
				}
			})
		})
		t.Run("and no way avoids it", func(t *testing.T) {
			r, ok := g.Shortest("l1", "showcase", model.Constraints{AvoidZones: []string{"showcase"}})
			t.Run("it should still find the route", func(t *testing.T) {
				if !ok {
					t.Fatal("no route")
				}
			})
			t.Run("it should report nothing avoided", func(t *testing.T) {
				if len(r.Avoided) != 0 {
					t.Fatalf("got %v", r.Avoided)
				}
			})
			t.Run("it should warn that no alternative exists", func(t *testing.T) {
				if len(r.Obstacles) != 1 || !strings.Contains(r.Obstacles[0].Text, "No alternative") {
					t.Fatalf("got %v", r.Obstacles)
				}
			})
			t.Run("it should say the step passes through the showcase", func(t *testing.T) {
				if !strings.Contains(r.Steps[0].Text, "passes through the Solutions Showcase") {
					t.Fatalf("got %q", r.Steps[0].Text)
				}
			})
		})
	})

	t.Run("when a restroom is wanted on the way", func(t *testing.T) {
		r, _ := g.Shortest("l1", "showcase", model.Constraints{WantAmenity: "restroom"})
		t.Run("it should list the amenity", func(t *testing.T) {
			if len(r.Amenities) != 1 || r.Amenities[0] != "restroom" {
				t.Fatalf("got %v", r.Amenities)
			}
		})
		t.Run("it should say so in the step", func(t *testing.T) {
			if !strings.HasSuffix(r.Steps[0].Text, "A restroom is on the way.") {
				t.Fatalf("got %q", r.Steps[0].Text)
			}
		})
	})

	t.Run("when the destination is unreachable", func(t *testing.T) {
		_, ok := g.Shortest("l1", "island", model.Constraints{})
		t.Run("it should say there is no route", func(t *testing.T) {
			if ok {
				t.Fatal("found a route to an island")
			}
		})
	})

	t.Run("when a node is not in the graph", func(t *testing.T) {
		_, ok := g.Shortest("l1", "room-999", model.Constraints{})
		t.Run("it should say there is no route", func(t *testing.T) {
			if ok {
				t.Fatal("found a route to an unknown node")
			}
		})
	})

	t.Run("when a step is unverified", func(t *testing.T) {
		r, _ := g.Shortest("l1", "l2", model.Constraints{})
		t.Run("it should carry the marking", func(t *testing.T) {
			if r.Steps[0].Verified {
				t.Fatal("stairs read off a plan marked verified")
			}
		})
	})
}

func TestLoad(t *testing.T) {
	cases := map[string]string{
		"an edge names an unknown node": "nodes: [{id: a, name: A}]\nedges: [{from: a, to: b, kind: corridor}]",
		"an edge has an unknown kind":   "nodes: [{id: a, name: A}, {id: b, name: B}]\nedges: [{from: a, to: b, kind: teleport}]",
		"a node has no id":              "nodes: [{name: A}]",
	}
	for name, text := range cases {
		t.Run("when "+name, func(t *testing.T) {
			_, err := Load(strings.NewReader(text))
			t.Run("it should refuse the file", func(t *testing.T) {
				if err == nil {
					t.Fatal("loaded")
				}
			})
		})
	}
}
