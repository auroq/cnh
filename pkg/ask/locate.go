package ask

import "context"

// IDLocator is the optional half of a Store that resolves a place id, the kind a door code carries, to where it is.
// A Store without it is still complete; Locate then only knows the room search.
type IDLocator interface {
	LocatePlace(ctx context.Context, id string) (Here, error)
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
