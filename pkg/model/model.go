// Package model is what the business means: the types every port speaks, with no dependency on how they are stored or served.
package model

import "time"

// Place is somewhere an attendee may need: an access feature, or a room named by the schedule.
type Place struct {
	ID       string   `json:"id"`
	Kind     string   `json:"kind"`
	Name     string   `json:"name"`
	Venue    string   `json:"venue"`
	Level    string   `json:"level"`
	Near     string   `json:"near"`
	Notes    string   `json:"notes,omitempty"`
	For      []string `json:"for,omitempty"`
	Route    []string `json:"route,omitempty"`
	Verified bool     `json:"verified"`
	Source   string   `json:"source"`
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

// Here is where the attendee is: a room they named, resolved to a venue and level.
type Here struct {
	Asked string `json:"asked"`
	Room  string `json:"room,omitempty"`
	Venue string `json:"venue,omitempty"`
	Level string `json:"level,omitempty"`
}

// Answer is what the attendee gets back: a sentence first, then the places and sessions behind it.
type Answer struct {
	Question string    `json:"question"`
	Here     Here      `json:"here"`
	Summary  string    `json:"summary"`
	Places   []Place   `json:"places"`
	Sessions []Session `json:"sessions"`
}
