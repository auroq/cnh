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

// Strength says what a chip does to an answer: Prefer ranks matching places first, Must removes the rest.
// A learned chip is always Prefer; Must is a deliberate choice.
type Strength string

const (
	Prefer Strength = "prefer"
	Must   Strength = "must"
)

// Pick is one chip in a profile with where it came from and how hard it binds.
type Pick struct {
	Chip     string
	Source   Source
	Strength Strength
}

// Preferences is the attendee's profile: the chips they hold, in catalogue order, each at most once.
// Encoded as comma-separated ids, a learned chip carrying a trailing "~" and a must a trailing "!": "step-free!,quiet-matters~".
type Preferences []Pick

const (
	learnedMark = "~"
	mustMark    = "!"
)

// ParsePreferences reads the encoding above, dropping unknown ids and duplicates; a chosen chip outranks a learned one and a must outranks a prefer.
func ParsePreferences(s string) Preferences {
	picks := map[string]Pick{}
	for _, raw := range strings.Split(s, ",") {
		id := strings.TrimSpace(raw)
		pick := Pick{Source: Chosen, Strength: Prefer}
		switch {
		case strings.HasSuffix(id, mustMark):
			id, pick.Strength = strings.TrimSuffix(id, mustMark), Must
		case strings.HasSuffix(id, learnedMark):
			id, pick.Source = strings.TrimSuffix(id, learnedMark), Learned
		}
		pick.Chip = id
		picks[id] = stronger(picks[id], pick)
	}
	var out Preferences
	for _, c := range Chips {
		if pick, ok := picks[c.ID]; ok {
			out = append(out, pick)
		}
	}
	return out
}

func stronger(have, next Pick) Pick {
	if have.Chip == "" || have.Source == Learned && next.Source == Chosen || next.Strength == Must {
		return next
	}
	return have
}

func (p Preferences) Has(id string) bool {
	for _, pick := range p {
		if pick.Chip == id {
			return true
		}
	}
	return false
}

// With adds id as a chosen chip, promoting it if it was only learned; a must stays a must.
func (p Preferences) With(id string) Preferences {
	return ParsePreferences(p.String() + "," + id)
}

// WithStrength sets id as a chosen chip at the given strength, replacing whatever it was.
func (p Preferences) WithStrength(id string, strength Strength) Preferences {
	mark := ""
	if strength == Must {
		mark = mustMark
	}
	return ParsePreferences(p.Without(id).String() + "," + id + mark)
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

// IsMust says whether id filters answers rather than ranking them.
func (p Preferences) IsMust(id string) bool {
	return p.StrengthOf(id) == Must
}

// StrengthOf is the strength of id, or "" when the profile does not hold it.
func (p Preferences) StrengthOf(id string) Strength {
	for _, pick := range p {
		if pick.Chip == id {
			return pick.Strength
		}
	}
	return ""
}

// Musts counts the chips that filter; Prefers counts the rest.
func (p Preferences) Musts() int {
	n := 0
	for _, pick := range p {
		if pick.Strength == Must {
			n++
		}
	}
	return n
}

func (p Preferences) Prefers() int {
	return len(p) - p.Musts()
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
		switch {
		case pick.Strength == Must:
			parts = append(parts, pick.Chip+mustMark)
		case pick.Source == Learned:
			parts = append(parts, pick.Chip+learnedMark)
		default:
			parts = append(parts, pick.Chip)
		}
	}
	return strings.Join(parts, ",")
}

// Applied says how one chip changed an answer, in words an attendee can read.
type Applied struct {
	Chip     string   `json:"chip"`
	Effect   string   `json:"effect"`
	Strength Strength `json:"strength,omitempty"`
}
