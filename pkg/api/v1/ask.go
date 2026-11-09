// Package v1 is the public JSON contract of GET /api/ask.
// Field names are frozen: a change to the wire shape is a new package, v2, not an edit here.
package v1

import "time"

// AskResponse is the body of GET /api/ask.
type AskResponse struct {
	Question string    `json:"question"`
	Here     Here      `json:"here"`
	Summary  string    `json:"summary"`
	Places   []Place   `json:"places"`
	Sessions []Session `json:"sessions"`
}

type Here struct {
	Asked string `json:"asked"`
	Room  string `json:"room,omitempty"`
	Venue string `json:"venue,omitempty"`
	Level string `json:"level,omitempty"`
	Side  string `json:"side,omitempty"`
	Via   string `json:"via,omitempty"`
}

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
	Side     string   `json:"side,omitempty"`
	LeadsTo  string   `json:"leads_to,omitempty"`

	AttendeeStatus string     `json:"attendee_status,omitempty"`
	ReportCount    int        `json:"report_count,omitempty"`
	LatestNote     string     `json:"latest_note,omitempty"`
	LatestAt       *time.Time `json:"latest_at,omitempty"`
}

type Session struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	URL      string    `json:"url"`
	Speakers []string  `json:"speakers"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Location string    `json:"location"`
	Venue    string    `json:"venue"`
	Level    string    `json:"level"`
	Room     string    `json:"room"`
}

// Error is the body of a non-2xx response.
type Error struct {
	Error string `json:"error"`
}
