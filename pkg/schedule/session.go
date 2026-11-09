// Package schedule is the model of a conference session as the app sees it, whatever the source.
package schedule

import "github.com/auroq/cnh/pkg/model"

// Session is the model type; this package only adds schedule-wide helpers.
type Session = model.Session

// Locations returns every session's "Venue | Level | Room" string, in order, duplicates included.
func Locations(sessions []Session) []string {
	out := make([]string, 0, len(sessions))
	for _, s := range sessions {
		out = append(out, s.Location)
	}
	return out
}
