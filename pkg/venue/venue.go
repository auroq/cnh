// Package venue holds the hand-curated access features and the rooms derived from the schedule.
package venue

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/auroq/cnh/pkg/model"
	"gopkg.in/yaml.v3"
)

// Place and Room are the model types; this package owns the curated file's shape and the rooms derived from the schedule.
type (
	Place = model.Place
	Room  = model.Room
)

type placeYAML struct {
	ID       string   `yaml:"id"`
	Kind     string   `yaml:"kind"`
	Name     string   `yaml:"name"`
	Venue    string   `yaml:"venue"`
	Level    string   `yaml:"level"`
	Near     string   `yaml:"near"`
	Notes    string   `yaml:"notes"`
	For      []string `yaml:"for"`
	Route    []string `yaml:"route"`
	Verified bool     `yaml:"verified"`
	Source   string   `yaml:"source"`
}

func (p placeYAML) model() model.Place {
	return model.Place{ID: p.ID, Kind: p.Kind, Name: p.Name, Venue: p.Venue, Level: p.Level, Near: p.Near, Notes: p.Notes, For: p.For, Route: p.Route, Verified: p.Verified, Source: p.Source}
}

type file struct {
	Places []placeYAML `yaml:"places"`
}

// Load reads places.yaml and validates that every place has an id, kind, name, venue and source.
func Load(r io.Reader) ([]Place, error) {
	var f file
	if err := yaml.NewDecoder(r).Decode(&f); err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for i, p := range f.Places {
		switch {
		case p.ID == "":
			return nil, fmt.Errorf("place %d: missing id", i)
		case seen[p.ID]:
			return nil, fmt.Errorf("place %s: duplicate id", p.ID)
		case p.Kind == "" || p.Name == "" || p.Venue == "" || p.Source == "":
			return nil, fmt.Errorf("place %s: kind, name, venue and source are required", p.ID)
		}
		seen[p.ID] = true
	}
	places := make([]Place, 0, len(f.Places))
	for _, p := range f.Places {
		places = append(places, p.model())
	}
	return places, nil
}

// RoomsFromLocations turns the distinct "Venue | Level | Room" strings of the schedule into room places.
// Rooms are verified because the organiser's own schedule names them.
func RoomsFromLocations(locations []string) []Place {
	seen := map[string]Place{}
	for _, loc := range locations {
		r, ok := SplitLocation(loc)
		if !ok || r.Room == "" {
			continue
		}
		id := "room-" + slug(r.Venue+"-"+r.Level+"-"+r.Room)
		seen[id] = Place{
			ID:       id,
			Kind:     "room",
			Name:     r.Room,
			Venue:    r.Venue,
			Level:    r.Level,
			Near:     loc,
			Verified: true,
			Source:   "https://sessionize.com/api/v2/uo948i5p/view/markdown",
		}
	}
	out := make([]Place, 0, len(seen))
	for _, p := range seen {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// SplitLocation parses "Salt Palace | Level 1 | Solutions Showcase | Halls 5 | Cosmic Canyon Zone".
// The venue is the first part, the level is the part that starts with "Level", and the room is everything else joined.
func SplitLocation(loc string) (Room, bool) {
	parts := strings.Split(loc, " | ")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		return Room{}, false
	}
	r := Room{Venue: strings.TrimSpace(parts[0])}
	var rest []string
	for _, p := range parts[1:] {
		p = strings.TrimSpace(p)
		if r.Level == "" && strings.HasPrefix(p, "Level ") {
			r.Level = p
			continue
		}
		rest = append(rest, p)
	}
	r.Room = strings.Join(rest, ", ")
	return r, true
}

func slug(s string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			dash = false
		default:
			if !dash && b.Len() > 0 {
				b.WriteByte('-')
				dash = true
			}
		}
	}
	return strings.TrimRight(b.String(), "-")
}

// File is a places.yaml on disk, loaded on demand.
type File string

// Places loads and validates the file.
func (f File) Places() ([]Place, error) {
	r, err := os.Open(string(f))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return Load(r)
}
