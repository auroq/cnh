// Package route is the walking graph between places: a hand-written file of nodes and edges,
// and the shortest way between two of them under the constraints an attendee's profile sets.
// It counts edges, never minutes (ADR 0012).
package route

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/auroq/cnh/pkg/model"
)

type nodeYAML struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

type edgeYAML struct {
	From        string   `yaml:"from"`
	To          string   `yaml:"to"`
	Kind        string   `yaml:"kind"`
	StepFree    bool     `yaml:"step_free"`
	ThroughZone string   `yaml:"through_zone"`
	Amenities   []string `yaml:"amenities"`
	Verified    bool     `yaml:"verified"`
	Source      string   `yaml:"source"`
}

type fileYAML struct {
	Nodes []nodeYAML `yaml:"nodes"`
	Edges []edgeYAML `yaml:"edges"`
}

// Edge is one way between two nodes, walkable in both directions.
type Edge struct {
	From, To    string
	Kind        string
	StepFree    bool
	ThroughZone string
	Amenities   []string
	Verified    bool
	Source      string
}

// Graph is the venue as nodes and edges.
type Graph struct {
	names map[string]string
	edges []Edge
}

var kinds = map[string]bool{"corridor": true, "door": true, "stairs": true, "escalator": true, "elevator": true, "skybridge": true}

// Load reads the graph and refuses an edge that names an unknown node or kind.
func Load(r io.Reader) (*Graph, error) {
	var f fileYAML
	if err := yaml.NewDecoder(r).Decode(&f); err != nil {
		return nil, fmt.Errorf("routes: %w", err)
	}
	g := &Graph{names: map[string]string{}}
	for _, n := range f.Nodes {
		if n.ID == "" {
			return nil, fmt.Errorf("routes: a node has no id")
		}
		g.names[n.ID] = n.Name
	}
	for i, e := range f.Edges {
		if _, ok := g.names[e.From]; !ok {
			return nil, fmt.Errorf("routes: edge %d: unknown node %q", i, e.From)
		}
		if _, ok := g.names[e.To]; !ok {
			return nil, fmt.Errorf("routes: edge %d: unknown node %q", i, e.To)
		}
		if !kinds[e.Kind] {
			return nil, fmt.Errorf("routes: edge %d: unknown kind %q", i, e.Kind)
		}
		g.edges = append(g.edges, Edge(e))
	}
	return g, nil
}

// File is the curated graph on disk.
type File string

// Graph loads the file.
func (f File) Graph() (*Graph, error) {
	r, err := os.Open(string(f))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return Load(r)
}

// Has reports whether id is a node of the graph.
func (g *Graph) Has(id string) bool {
	_, ok := g.names[id]
	return ok
}

// Shortest is the way from one node to another with the fewest edges that respects the constraints.
// When a zone cannot be avoided the route goes through it and says so in Obstacles; when stairs cannot be avoided there is no route.
func (g *Graph) Shortest(from, to string, c model.Constraints) (model.Route, bool) {
	if !g.Has(from) || !g.Has(to) {
		return model.Route{}, false
	}
	avoided := c.AvoidZones
	path, ok := g.bfs(from, to, c.NoStairs, avoided)
	if !ok && len(avoided) > 0 {
		avoided = nil
		path, ok = g.bfs(from, to, c.NoStairs, nil)
	}
	if !ok {
		return model.Route{}, false
	}
	r := model.Route{From: g.names[from], To: g.names[to], Avoided: avoided}
	for _, e := range path {
		step := g.step(e, c)
		if step.Obstacle != nil {
			r.Obstacles = append(r.Obstacles, *step.Obstacle)
		}
		for _, a := range e.Amenities {
			if c.WantAmenity == "" || c.WantAmenity == a {
				r.Amenities = appendMissing(r.Amenities, a)
			}
		}
		r.Steps = append(r.Steps, step)
	}
	return r, true
}

type hop struct {
	Edge
	to string
}

func (g *Graph) bfs(from, to string, noStairs bool, avoid []string) ([]hop, bool) {
	prev := map[string]hop{}
	seen := map[string]bool{from: true}
	queue := []string{from}
	for len(queue) > 0 {
		at := queue[0]
		queue = queue[1:]
		if at == to {
			break
		}
		for _, h := range g.hops(at) {
			if seen[h.to] || (noStairs && !h.StepFree) || inZone(h.ThroughZone, avoid) {
				continue
			}
			seen[h.to] = true
			prev[h.to] = h
			queue = append(queue, h.to)
		}
	}
	if !seen[to] {
		return nil, false
	}
	var path []hop
	for at := to; at != from; {
		h := prev[at]
		path = append([]hop{h}, path...)
		at = h.From
	}
	return path, true
}

func (g *Graph) hops(at string) []hop {
	var out []hop
	for _, e := range g.edges {
		switch at {
		case e.From:
			out = append(out, hop{e, e.To})
		case e.To:
			rev := e
			rev.From, rev.To = e.To, e.From
			out = append(out, hop{rev, e.From})
		}
	}
	return out
}

func inZone(zone string, avoid []string) bool {
	for _, z := range avoid {
		if z == zone {
			return true
		}
	}
	return false
}

var zoneNames = map[string]string{"showcase": "the Solutions Showcase", "registration": "the registration area", "food": "the food and drink area"}

func (g *Graph) step(h hop, c model.Constraints) model.Step {
	dest := g.names[h.to]
	s := model.Step{Kind: h.Kind, Verified: h.Verified}
	switch h.Kind {
	case "elevator":
		s.Text = "Take the elevator to " + dest + "."
	case "stairs":
		s.Text = "Take the stairs to " + dest + "."
		s.Obstacle = &model.Obstacle{Kind: "stairs", Text: "Stairs on the way to " + dest + ".", Verified: h.Verified}
	case "escalator":
		s.Text = "Take the escalator to " + dest + "."
		s.Obstacle = &model.Obstacle{Kind: "escalator", Text: "An escalator on the way to " + dest + ", no step-free alternative on this leg.", Verified: h.Verified}
	case "door":
		s.Text = "Go through the doors to " + dest + "."
		if !h.StepFree {
			s.Obstacle = &model.Obstacle{Kind: "door", Text: "Whether these doors are step-free has not been checked.", Verified: h.Verified}
		}
	case "skybridge":
		s.Text = "Cross the connection into " + dest + "."
	default:
		s.Text = "Follow the corridor to " + dest + "."
	}
	if zone := zoneNames[h.ThroughZone]; zone != "" {
		s.Text = strings.TrimSuffix(s.Text, ".") + "; this way passes through " + zone + "."
		if inZone(h.ThroughZone, c.AvoidZones) {
			s.Obstacle = &model.Obstacle{Kind: "zone", Text: "No alternative avoids " + zone + " on this route.", Verified: h.Verified}
		}
	}
	for _, a := range h.Amenities {
		if c.WantAmenity != "" && c.WantAmenity == a {
			s.Text += " A " + a + " is on the way."
		}
	}
	return s
}

func appendMissing(have []string, s string) []string {
	for _, h := range have {
		if h == s {
			return have
		}
	}
	return append(have, s)
}
