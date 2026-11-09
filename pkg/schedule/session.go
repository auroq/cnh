// Package schedule is the model of a conference session as the app sees it, whatever the source.
package schedule

import "time"

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

// Locations returns every session's "Venue | Level | Room" string, in order, duplicates included.
func Locations(sessions []Session) []string {
	out := make([]string, 0, len(sessions))
	for _, s := range sessions {
		out = append(out, s.Location)
	}
	return out
}
