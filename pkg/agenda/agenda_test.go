package agenda

import (
	"strings"
	"testing"
	"time"

	"github.com/auroq/cnh/pkg/model"
)

var day = time.Date(2026, 11, 11, 0, 0, 0, 0, time.UTC)

func at(h, m int) time.Time {
	return day.Add(time.Duration(h)*time.Hour + time.Duration(m)*time.Minute)
}

func talk(id string, start, end time.Time) model.Session {
	return model.Session{ID: id, Title: "Talk " + id, Start: start, End: end, Room: "Room " + id, Venue: "Salt Palace", Level: "Level 1"}
}

func TestNextAndJustLeft(t *testing.T) {
	first := talk("a", at(9, 0), at(9, 35))
	second := talk("b", at(10, 0), at(10, 35))
	third := talk("c", at(11, 0), at(11, 35))
	sessions := []model.Session{third, first, second}

	cases := []struct {
		name     string
		now      time.Time
		next     string
		justLeft string
	}{
		{"before the day starts", at(8, 0), "a", ""},
		{"ten minutes after the first talk ended", at(9, 45), "b", "a"},
		{"exactly thirty minutes after the first talk ended", at(10, 5), "c", "a"},
		{"inside the second talk", at(10, 20), "c", ""},
		{"forty minutes after the second talk ended", at(11, 15), "", ""},
		{"after the last talk ended", at(11, 40), "", "c"},
		{"the next morning", at(24+9, 0), "", ""},
	}
	for _, c := range cases {
		t.Run("when it is "+c.name, func(t *testing.T) {
			t.Run("it should say the next session is "+orNone(c.next), func(t *testing.T) {
				if got := idOf(Next(sessions, c.now)); got != c.next {
					t.Fatalf("got %q, want %q", got, c.next)
				}
			})
			t.Run("it should say the session just left is "+orNone(c.justLeft), func(t *testing.T) {
				if got := idOf(JustLeft(sessions, c.now)); got != c.justLeft {
					t.Fatalf("got %q, want %q", got, c.justLeft)
				}
			})
		})
	}

	t.Run("when a session has no end time", func(t *testing.T) {
		open := model.Session{ID: "open", Start: at(9, 0)}
		t.Run("it should never count as just left", func(t *testing.T) {
			if got := JustLeft([]model.Session{open}, at(9, 10)); got != nil {
				t.Fatalf("got %+v", got)
			}
		})
	})
}

func TestBackToBack(t *testing.T) {
	cases := []struct {
		name     string
		sessions []model.Session
		want     bool
	}{
		{"two sessions in a row", []model.Session{talk("a", at(9, 0), at(9, 35)), talk("b", at(9, 40), at(10, 15))}, false},
		{"three sessions with five minutes between each", []model.Session{talk("c", at(10, 20), at(10, 55)), talk("a", at(9, 0), at(9, 35)), talk("b", at(9, 40), at(10, 15))}, true},
		{"three sessions with exactly ten minutes between each", []model.Session{talk("a", at(9, 0), at(9, 35)), talk("b", at(9, 45), at(10, 20)), talk("c", at(10, 30), at(11, 5))}, false},
		{"three sessions with a lunch break in the middle", []model.Session{talk("a", at(9, 0), at(9, 35)), talk("b", at(9, 40), at(10, 15)), talk("c", at(12, 0), at(12, 35))}, false},
		{"four sessions where the last three run together", []model.Session{talk("a", at(9, 0), at(9, 35)), talk("b", at(13, 0), at(13, 35)), talk("c", at(13, 40), at(14, 15)), talk("d", at(14, 20), at(14, 55))}, true},
		{"three overlapping sessions", []model.Session{talk("a", at(9, 0), at(9, 35)), talk("b", at(9, 30), at(10, 5)), talk("c", at(10, 0), at(10, 35))}, true},
		{"no sessions", nil, false},
	}
	for _, c := range cases {
		t.Run("when the schedule has "+c.name, func(t *testing.T) {
			t.Run("it should "+strings.TrimSpace(map[bool]string{true: "", false: "not "}[c.want])+"call it back to back", func(t *testing.T) {
				if got := BackToBack(c.sessions); got != c.want {
					t.Fatalf("got %v, want %v", got, c.want)
				}
			})
		})
	}
}

func TestIDs(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"nothing", "", ""},
		{"three ids", "a,b,c", "a,b,c"},
		{"spaces and an empty id", " a, ,b ", "a,b"},
		{"an id twice", "a,a,b", "a,b"},
		{"an id with a character no session id has", "a,<script>,b", "a,b"},
	}
	for _, c := range cases {
		t.Run("when the sched parameter holds "+c.name, func(t *testing.T) {
			t.Run("it should keep "+quoted(c.want), func(t *testing.T) {
				if got := JoinIDs(ParseIDs(c.in)); got != c.want {
					t.Fatalf("got %q, want %q", got, c.want)
				}
			})
		})
	}
}

func idOf(s *model.Session) string {
	if s == nil {
		return ""
	}
	return s.ID
}

func orNone(id string) string {
	if id == "" {
		return "none"
	}
	return id
}

func quoted(s string) string {
	if s == "" {
		return "nothing"
	}
	return `"` + s + `"`
}
