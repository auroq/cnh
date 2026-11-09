// Package model is what the business means: the types every port speaks, with no dependency on how they are stored or served.
package model

import "time"

// Place is somewhere an attendee may need: an access feature, or a room named by the schedule.
type Place struct {
	ID       string          `json:"id"`
	Kind     string          `json:"kind"`
	Name     string          `json:"name"`
	Venue    string          `json:"venue"`
	Level    string          `json:"level"`
	Near     string          `json:"near"`
	Notes    string          `json:"notes,omitempty"`
	For      []string        `json:"for,omitempty"`
	Route    []string        `json:"route,omitempty"`
	Verified bool            `json:"verified"`
	Source   string          `json:"source"`
	Aliases  []string        `json:"aliases,omitempty"`
	Side     string          `json:"side,omitempty"`
	LeadsTo  string          `json:"leads_to,omitempty"`
	Attendee *AttendeeStatus `json:"attendee,omitempty"`
}

// Room is a location the schedule names, split into its parts.
type Room struct {
	Venue string
	Level string
	Room  string
}

// Session is a talk on the schedule.
type Session struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Speakers    []string  `json:"speakers"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	Location    string    `json:"location"`
	Venue       string    `json:"venue"`
	Level       string    `json:"level,omitempty"`
	Room        string    `json:"room,omitempty"`
	Description string    `json:"description,omitempty"`
}

// Here is where the attendee is: a place they described in their own words, resolved to a venue and level.
// Side is the street side when they stand at a door; Via names the rule that resolved what they said.
type Here struct {
	Asked string `json:"asked"`
	Room  string `json:"room,omitempty"`
	Venue string `json:"venue,omitempty"`
	Level string `json:"level,omitempty"`
	Side  string `json:"side,omitempty"`
	Via   string `json:"via,omitempty"`
}

// Need is what an attendee said they need, in their words, mapped to the kinds of place that can meet it.
// Kinds are in the order to offer them; Note is said when the place offered is a stand-in for what was asked.
type Need struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Phrases []string `json:"phrases"`
	Kinds   []string `json:"kinds"`
	Note    string   `json:"note,omitempty"`
}

// Answer is what the attendee gets back: a sentence first, then the places and sessions behind it.
// Need is set when the question was understood as a need rather than a kind of place.
type Answer struct {
	Question string    `json:"question"`
	Here     Here      `json:"here"`
	Summary  string    `json:"summary"`
	Need     *Need     `json:"need,omitempty"`
	Places   []Place   `json:"places"`
	Sessions []Session `json:"sessions"`
}

// Verdict is what an attendee who stood at a place said about it.
type Verdict string

const (
	VerdictWorks          Verdict = "works"
	VerdictProblem        Verdict = "problem"
	VerdictNotAsDescribed Verdict = "not-as-described"
)

// Report is one attendee's word about one place: anonymous, timed, and hideable by a volunteer.
// Token is the random per-browser id used only to count distinct people and to rate limit; it is never shown.
type Report struct {
	ID      string    `json:"id"`
	PlaceID string    `json:"place_id"`
	Verdict Verdict   `json:"verdict"`
	Note    string    `json:"note,omitempty"`
	At      time.Time `json:"at"`
	Token   string    `json:"token"`
	Source  string    `json:"source"`
	Hidden  bool      `json:"hidden"`
}

// AttendeeStatus is what live reports add up to for one place.
// Status is none, verified, problem or disputed; Count is how many distinct attendees said so.
type AttendeeStatus struct {
	Status     string    `json:"status"`
	Count      int       `json:"count,omitempty"`
	LatestNote string    `json:"latest_note,omitempty"`
	LatestAt   time.Time `json:"latest_at,omitempty"`
}

const (
	StatusNone     = "none"
	StatusVerified = "verified"
	StatusProblem  = "problem"
	StatusDisputed = "disputed"
)
