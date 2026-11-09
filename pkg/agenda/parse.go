// Package agenda reads the schedule an attendee pasted and answers what comes next in it.
// It is pure: every rule takes the clock as an argument.
package agenda

import (
	"strings"
)

const summaryLine = "SUMMARY"

// Parse reads a Sched iCal export or a plain list of titles, one per line, and returns the distinct titles in order.
func Parse(text string) []string {
	lines := unfold(strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n"))
	if isICal(lines) {
		return distinct(summaries(lines))
	}
	return distinct(lines)
}

func isICal(lines []string) bool {
	for _, l := range lines {
		if strings.HasPrefix(l, "BEGIN:VCALENDAR") {
			return true
		}
	}
	return false
}

func unfold(lines []string) []string {
	var out []string
	for _, l := range lines {
		if len(out) > 0 && (strings.HasPrefix(l, " ") || strings.HasPrefix(l, "\t")) && strings.HasPrefix(strings.ToUpper(out[len(out)-1]), summaryLine) {
			out[len(out)-1] += l[1:]
			continue
		}
		out = append(out, l)
	}
	return out
}

func summaries(lines []string) []string {
	var out []string
	for _, l := range lines {
		name, value, ok := strings.Cut(l, ":")
		if !ok {
			continue
		}
		name, _, _ = strings.Cut(name, ";")
		if strings.ToUpper(name) == summaryLine {
			out = append(out, unescape(value))
		}
	}
	return out
}

var unescaper = strings.NewReplacer(`\,`, ",", `\;`, ";", `\n`, " ", `\N`, " ", `\\`, `\`)

func unescape(s string) string {
	return unescaper.Replace(s)
}

func distinct(lines []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" || seen[l] {
			continue
		}
		seen[l] = true
		out = append(out, l)
	}
	return out
}
