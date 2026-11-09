package sessionize

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/auroq/cnh/pkg/schedule"
	"github.com/auroq/cnh/pkg/venue"
)

// Session is the schedule model; this package only parses into it.
type Session = schedule.Session

type Detail struct {
	ID          string
	Description string
}

var (
	mst        = time.FixedZone("MST", -7*3600)
	dayHeading = regexp.MustCompile(`^### (.+)$`)
	bullet     = regexp.MustCompile(`^- \[(.+)\]\((\S+/session/(\d+))\)\s*$`)
	timeRange  = regexp.MustCompile(`^(\d{1,2}:\d{2} [ap]m) - (\d{1,2}:\d{2} [ap]m)$`)
	sessionID  = regexp.MustCompile(`^- \*\*Session ID:\*\* (\d+)`)
)

func ParseIndex(r io.Reader) ([]Session, error) {
	var (
		sessions []Session
		day      time.Time
		inList   bool
		pending  *Session
	)
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 1024*1024), 4*1024*1024)
	for sc.Scan() {
		line := sc.Text()
		switch {
		case line == "## Sessions":
			inList = true
			continue
		case strings.HasPrefix(line, "## "):
			if inList {
				return sessions, nil
			}
			continue
		case !inList:
			continue
		}
		if m := dayHeading.FindStringSubmatch(line); m != nil {
			d, err := time.ParseInLocation("Monday, January 2, 2006", m[1], mst)
			if err != nil {
				return nil, fmt.Errorf("day heading %q: %w", m[1], err)
			}
			day = d
			continue
		}
		if m := bullet.FindStringSubmatch(line); m != nil {
			sessions = append(sessions, Session{ID: m[3], Title: unescape(m[1]), URL: m[2]})
			pending = &sessions[len(sessions)-1]
			continue
		}
		if pending != nil && strings.HasPrefix(line, "  ") {
			if err := parseMeta(pending, strings.TrimSpace(line), day); err != nil {
				return nil, fmt.Errorf("session %s: %w", pending.ID, err)
			}
			pending = nil
		}
	}
	return sessions, sc.Err()
}

func parseMeta(s *Session, meta string, day time.Time) error {
	parts := strings.Split(meta, " · ")
	var timePart string
	switch len(parts) {
	case 2:
		timePart, s.Location = parts[0], parts[1]
	case 3:
		s.Speakers = splitSpeakers(parts[0])
		timePart, s.Location = parts[1], parts[2]
	default:
		return fmt.Errorf("unexpected meta line %q", meta)
	}
	m := timeRange.FindStringSubmatch(timePart)
	if m == nil {
		return fmt.Errorf("unexpected time range %q", timePart)
	}
	var err error
	if s.Start, err = atTime(day, m[1]); err != nil {
		return err
	}
	if s.End, err = atTime(day, m[2]); err != nil {
		return err
	}
	if r, ok := venue.SplitLocation(s.Location); ok {
		s.Venue, s.Level, s.Room = r.Venue, r.Level, r.Room
	}
	return nil
}

func atTime(day time.Time, clock string) (time.Time, error) {
	t, err := time.ParseInLocation("3:04 pm", clock, mst)
	if err != nil {
		return time.Time{}, fmt.Errorf("time %q: %w", clock, err)
	}
	return time.Date(day.Year(), day.Month(), day.Day(), t.Hour(), t.Minute(), 0, 0, mst), nil
}

func splitSpeakers(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func unescape(s string) string {
	return strings.NewReplacer(`\[`, "[", `\]`, "]", `\*`, "*", `\_`, "_").Replace(s)
}

func ParseDetail(r io.Reader) (Detail, error) {
	var (
		d      Detail
		inDesc bool
		desc   []string
		sc     = bufio.NewScanner(r)
	)
	sc.Buffer(make([]byte, 0, 1024*1024), 4*1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if m := sessionID.FindStringSubmatch(line); m != nil {
			d.ID = m[1]
			continue
		}
		if line == "## Description" {
			inDesc = true
			continue
		}
		if inDesc && strings.HasPrefix(line, "## ") {
			break
		}
		if inDesc {
			desc = append(desc, line)
		}
	}
	d.Description = strings.TrimSpace(strings.Join(desc, "\n"))
	return d, sc.Err()
}
