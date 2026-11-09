package model

import "strings"

// Chip is one need an attendee may state; never a diagnosis.
type Chip struct {
	ID    string
	Label string
}

// Chips is every chip the profile may hold, in the order the page shows them.
var Chips = []Chip{
	{"step-free", "Step-free only"},
	{"avoid-crowds", "Avoid crowds and noise"},
	{"quiet-matters", "Quiet spaces matter"},
	{"all-gender", "All-gender restrooms"},
	{"captions-asl", "Captions or ASL"},
	{"wheels", "Wide doors and ramps"},
	{"dietary", "Dietary needs, unspecified"},
}

// Source says how a chip got into the profile: the attendee chose it, or Wide Door learned it from a question.
type Source string

const (
	Chosen  Source = "chosen"
	Learned Source = "learned"
)

// Pick is one chip in a profile with where it came from.
type Pick struct {
	Chip   string
	Source Source
}

// Preferences is the attendee's profile: the chips they hold, in catalogue order, each at most once.
// Encoded as comma-separated ids, a learned chip carrying a trailing "~": "step-free,quiet-matters~".
type Preferences []Pick

const learnedMark = "~"

// ParsePreferences reads the encoding above, dropping unknown ids and duplicates; a chosen chip outranks a learned one.
func ParsePreferences(s string) Preferences {
	sources := map[string]Source{}
	for _, raw := range strings.Split(s, ",") {
		id := strings.TrimSpace(raw)
		source := Chosen
		if strings.HasSuffix(id, learnedMark) {
			id, source = strings.TrimSuffix(id, learnedMark), Learned
		}
		if sources[id] != Chosen {
			sources[id] = source
		}
	}
	var out Preferences
	for _, c := range Chips {
		if source, ok := sources[c.ID]; ok {
			out = append(out, Pick{Chip: c.ID, Source: source})
		}
	}
	return out
}

func (p Preferences) Has(id string) bool {
	for _, pick := range p {
		if pick.Chip == id {
			return true
		}
	}
	return false
}

// With adds id as a chosen chip, promoting it if it was only learned.
func (p Preferences) With(id string) Preferences {
	return ParsePreferences(p.String() + "," + id)
}

// WithLearned adds id as a learned chip; a chip already held keeps its source.
func (p Preferences) WithLearned(id string) Preferences {
	return ParsePreferences(p.String() + "," + id + learnedMark)
}

// IsLearned says whether id is in the profile because a question implied it, not because the attendee chose it.
func (p Preferences) IsLearned(id string) bool {
	for _, pick := range p {
		if pick.Chip == id {
			return pick.Source == Learned
		}
	}
	return false
}

func (p Preferences) Without(id string) Preferences {
	var out Preferences
	for _, pick := range p {
		if pick.Chip != id {
			out = append(out, pick)
		}
	}
	return out
}

// IDs lists the chips without their sources.
func (p Preferences) IDs() []string {
	out := make([]string, 0, len(p))
	for _, pick := range p {
		out = append(out, pick.Chip)
	}
	return out
}

func (p Preferences) String() string {
	parts := make([]string, 0, len(p))
	for _, pick := range p {
		if pick.Source == Learned {
			parts = append(parts, pick.Chip+learnedMark)
		} else {
			parts = append(parts, pick.Chip)
		}
	}
	return strings.Join(parts, ",")
}

// Applied says how one chip changed an answer, in words an attendee can read.
type Applied struct {
	Chip   string `json:"chip"`
	Effect string `json:"effect"`
}
