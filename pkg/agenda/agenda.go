package agenda

import (
	"sort"
	"strings"
	"time"

	"github.com/auroq/cnh/pkg/model"
)

const (
	// JustLeftWindow is how long after a session ends the attendee is assumed to be still near its room.
	JustLeftWindow = 30 * time.Minute
	// BackToBackGap is the longest break that still counts as no break between sessions.
	BackToBackGap = 10 * time.Minute
	// BackToBackRun is how many sessions without a break earn the offer of a quiet room.
	BackToBackRun = 3
)

// Next is the first session that has not started yet, or nil.
func Next(sessions []model.Session, now time.Time) *model.Session {
	var next *model.Session
	for i := range sessions {
		s := &sessions[i]
		if !s.Start.After(now) {
			continue
		}
		if next == nil || s.Start.Before(next.Start) {
			next = s
		}
	}
	return next
}

// JustLeft is the session that ended most recently within JustLeftWindow, or nil.
func JustLeft(sessions []model.Session, now time.Time) *model.Session {
	var left *model.Session
	for i := range sessions {
		s := &sessions[i]
		if s.End.IsZero() || s.End.After(now) || now.Sub(s.End) > JustLeftWindow {
			continue
		}
		if left == nil || s.End.After(left.End) {
			left = s
		}
	}
	return left
}

// BackToBack is true when BackToBackRun sessions follow each other with less than BackToBackGap between them.
func BackToBack(sessions []model.Session) bool {
	ordered := byStart(sessions)
	run := 1
	for i := 1; i < len(ordered); i++ {
		if ordered[i].Start.Sub(ordered[i-1].End) < BackToBackGap {
			run++
		} else {
			run = 1
		}
		if run >= BackToBackRun {
			return true
		}
	}
	return false
}

func byStart(sessions []model.Session) []model.Session {
	out := append([]model.Session(nil), sessions...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Start.Before(out[j].Start) })
	return out
}

// ParseIDs reads the sched parameter: comma-separated session ids, each once, anything that is not an id dropped.
func ParseIDs(s string) []string {
	seen := map[string]bool{}
	var out []string
	for _, raw := range strings.Split(s, ",") {
		id := strings.TrimSpace(raw)
		if id == "" || seen[id] || !looksLikeID(id) {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

// JoinIDs is the inverse of ParseIDs.
func JoinIDs(ids []string) string {
	return strings.Join(ids, ",")
}

func looksLikeID(s string) bool {
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
		default:
			return false
		}
	}
	return true
}
