package ask

import (
	"context"
	"testing"

	"github.com/auroq/cnh/pkg/model"
)

type fakeRouter struct{ got model.Constraints }

func (f *fakeRouter) Has(id string) bool { return id == "sp-l1-lobby" || id == "quiet-room-260a" }

func (f *fakeRouter) Shortest(_, _ string, c model.Constraints) (model.Route, bool) {
	f.got = c
	return model.Route{Steps: []model.Step{{Text: "Take the elevator.", Kind: "elevator", Verified: true}}}, true
}

func routedService(router *fakeRouter) *Service {
	return &Service{Store: &fakeStore{places: []Place{{ID: "quiet-room-260a", Kind: "quiet-room", Name: "Quiet room"}}}, Router: router}
}

func TestRouteFor(t *testing.T) {
	t.Run("when here and the first place are both graph nodes", func(t *testing.T) {
		cases := map[string]struct {
			prefs string
			check func(model.Constraints) bool
		}{
			"no profile":          {"", func(c model.Constraints) bool { return !c.NoStairs && len(c.AvoidZones) == 0 }},
			"step-free chosen":    {"step-free", func(c model.Constraints) bool { return c.NoStairs }},
			"wheels chosen":       {"wheels", func(c model.Constraints) bool { return c.NoStairs }},
			"avoid-crowds chosen": {"avoid-crowds", func(c model.Constraints) bool { return len(c.AvoidZones) == 1 && c.AvoidZones[0] == "showcase" }},
		}
		for name, c := range cases {
			t.Run("and "+name, func(t *testing.T) {
				router := &fakeRouter{}
				ans, err := routedService(router).AskFor(context.Background(), "quiet room", "sp-l1-lobby", model.ParsePreferences(c.prefs))
				if err != nil {
					t.Fatal(err)
				}
				t.Run("it should attach the route", func(t *testing.T) {
					if ans.Route == nil || len(ans.Route.Steps) != 1 {
						t.Fatalf("got %v", ans.Route)
					}
				})
				t.Run("it should route under the profile's constraints", func(t *testing.T) {
					if !c.check(router.got) {
						t.Fatalf("got %+v", router.got)
					}
				})
			})
		}
	})

	t.Run("when here is not a graph node", func(t *testing.T) {
		ans, _ := routedService(&fakeRouter{}).Ask(context.Background(), "quiet room", "")
		t.Run("it should attach no route", func(t *testing.T) {
			if ans.Route != nil {
				t.Fatalf("got %v", ans.Route)
			}
		})
	})
}
