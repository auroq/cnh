package ask

import (
	"strings"

	"github.com/auroq/cnh/pkg/model"
)

// Router is what ask needs from a walking graph: whether a place is a node, and the shortest way between two under constraints.
type Router interface {
	Has(id string) bool
	Shortest(from, to string, c model.Constraints) (model.Route, bool)
}

type routeEffect struct {
	chip  string
	apply func(*model.Constraints)
	says  string
}

var routeEffects = []routeEffect{
	{"step-free", func(c *model.Constraints) { c.NoStairs = true }, "the route avoids stairs"},
	{"wheels", func(c *model.Constraints) { c.NoStairs = true }, "the route avoids stairs"},
	{"avoid-crowds", func(c *model.Constraints) { c.AvoidZones = append(c.AvoidZones, "showcase") }, "the route avoids the Solutions Showcase when it can"},
}

// routeFor is the hook after ranking: when both ends are graph nodes, the answer gains the way between them.
func (s *Service) routeFor(ans *Answer, prefs Preferences) {
	if s.Router == nil || len(ans.Places) == 0 {
		return
	}
	from := s.nodeFor(ans.Here)
	to := ans.Places[0].ID
	if from == "" || !s.Router.Has(to) || from == to {
		return
	}
	var c model.Constraints
	var applied []routeEffect
	for _, e := range routeEffects {
		if prefs.Has(e.chip) {
			e.apply(&c)
			applied = append(applied, e)
		}
	}
	r, ok := s.Router.Shortest(from, to, c)
	if !ok {
		return
	}
	ans.Route = &r
	for _, e := range applied {
		addApplied(ans, e.chip, e.says)
	}
}

func (s *Service) nodeFor(here Here) string {
	for _, id := range []string{strings.TrimSpace(here.Asked), here.Room} {
		if id != "" && s.Router.Has(id) {
			return id
		}
	}
	return ""
}
