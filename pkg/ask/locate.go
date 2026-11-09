package ask

import (
	"context"
	"strings"
)

// IDLocator is the optional half of a Store that resolves a place id, the kind a door code carries, to where it is.
// A Store without it is still complete; Locate then starts at the room name.
type IDLocator interface {
	LocatePlace(ctx context.Context, id string) (Here, error)
}

// The Via values: which rule resolved what the attendee said.
const (
	ViaID        = "id"
	ViaRoom      = "room"
	ViaAlias     = "alias"
	ViaSession   = "session"
	ViaFuzzyRoom = "fuzzy-room"
	ViaNearby    = "nearby"
)

const stepFreeEntrance = "step-free-entrance"

type rule struct {
	via  string
	span string
	find func(ctx context.Context) (Here, error)
}

// rules is the resolution order: a door code's id, the exact room name, a phrase people use for a place,
// the title of the session they just left, a misspelt room, and last what the words most resemble.
func (s *Service) rules(asked string) []rule {
	var out []rule
	if byID, ok := s.Store.(IDLocator); ok && looksLikeID(asked) {
		out = append(out, rule{ViaID, "LocatePlace", func(ctx context.Context) (Here, error) { return byID.LocatePlace(ctx, asked) }})
	}
	now := s.now()
	return append(out,
		rule{ViaRoom, "LocateRoom", func(ctx context.Context) (Here, error) { return s.Store.LocateRoom(ctx, asked) }},
		rule{ViaAlias, "LocateAlias", func(ctx context.Context) (Here, error) { return s.Store.LocateAlias(ctx, asked) }},
		rule{ViaSession, "LocateSession", func(ctx context.Context) (Here, error) { return s.Store.LocateSession(ctx, asked, now) }},
		rule{ViaFuzzyRoom, "LocateRoomFuzzy", func(ctx context.Context) (Here, error) { return s.Store.LocateRoomFuzzy(ctx, asked) }},
		rule{ViaNearby, "LocateNearby", func(ctx context.Context) (Here, error) { return s.Store.LocateNearby(ctx, asked) }},
	)
}

// looksLikeID is true for the slugs places are keyed by: lowercase letters, digits and at least one hyphen.
func looksLikeID(s string) bool {
	hyphen := false
	for _, r := range s {
		switch {
		case r == '-':
			hyphen = true
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
		default:
			return false
		}
	}
	return hyphen
}

var mobilityWords = []string{"stairs", "stair", "steps", "step", "wheelchair", "mobility", "ramp", "walker", "crutches", "scooter"}

// mentionsMobility is true when the question says the attendee cannot take steps, so a door must be step-free.
func mentionsMobility(question string) bool {
	words := " " + strings.Join(strings.FieldsFunc(strings.ToLower(question), isNotLetter), " ") + " "
	for _, w := range mobilityWords {
		if strings.Contains(words, " "+w+" ") {
			return true
		}
	}
	return false
}
