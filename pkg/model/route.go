package model

// Constraints is what the attendee's profile says a route must not do.
type Constraints struct {
	NoStairs    bool
	AvoidZones  []string
	WantAmenity string
}

// Obstacle is one thing on the way that matters to somebody: stairs, a crowd zone, a door nobody has checked.
type Obstacle struct {
	Kind     string `json:"kind"`
	Text     string `json:"text"`
	Verified bool   `json:"verified"`
}

// Step is one plain-language instruction on the way; Obstacle is set when the step itself is the obstacle.
type Step struct {
	Text     string    `json:"text"`
	Kind     string    `json:"kind"`
	Obstacle *Obstacle `json:"obstacle,omitempty"`
	Verified bool      `json:"verified"`
}

// Route is the way from where the attendee is to the first place, as facts, never minutes (ADR 0012).
// Avoided names the zones the route stayed out of; Amenities names what is on the way.
type Route struct {
	From      string     `json:"from"`
	To        string     `json:"to"`
	Steps     []Step     `json:"steps"`
	Obstacles []Obstacle `json:"obstacles,omitempty"`
	Avoided   []string   `json:"avoided,omitempty"`
	Amenities []string   `json:"amenities,omitempty"`
}
