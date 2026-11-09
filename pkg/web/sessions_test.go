package web

import (
	"strings"
	"testing"
	"time"

	"github.com/auroq/cnh/pkg/ask"
)

func TestSessionsOnThePage(t *testing.T) {
	noon := time.Date(2026, 11, 11, 19, 0, 0, 0, time.UTC)
	course := ask.Session{Title: "Sign Language Crash Course", URL: "https://example.org/1", Start: noon.Add(25 * time.Minute), End: noon.Add(85 * time.Minute), Location: "Salt Palace | Level 1 | Grand Ballroom J"}

	t.Run("when the answer lists sessions", func(t *testing.T) {
		_, body := get(t, fakeAsker{answer: ask.Answer{Summary: "s", Sessions: []ask.Session{course}, SessionsShownBecause: ask.ShownForSessionQuestion}}, "/ask?q=sign+language")
		t.Run("it should head them as sessions matching the question", func(t *testing.T) {
			if !strings.Contains(body, "<h3>Sessions matching your question</h3>") {
				t.Fatalf("heading missing in %s", body)
			}
		})
	})

	t.Run("when the answer left sessions out", func(t *testing.T) {
		_, body := get(t, fakeAsker{answer: ask.Answer{Summary: "s", SessionsShownBecause: ask.SessionsHidden}}, "/ask?q=quiet+room")
		t.Run("it should have no sessions heading", func(t *testing.T) {
			if strings.Contains(body, "Sessions matching your question") {
				t.Fatalf("sessions heading in %s", body)
			}
		})
	})

	timeCases := []struct {
		name    string
		session ask.Session
		want    string
	}{
		{"it starts in 25 minutes", withTime(course, ask.SessionUpcoming, 25), "Starts in 25 minutes"},
		{"it starts in 1 minute", withTime(course, ask.SessionUpcoming, 1), "Starts in 1 minute"},
		{"it starts in two days", withTime(course, ask.SessionUpcoming, 2880), "Starts Wednesday 12:25 pm"},
		{"it is running", withTime(course, ask.SessionRunning, 0), "Happening now, ends 1:25 pm"},
		{"it has ended", withTime(course, ask.SessionEnded, 0), "Ended at 1:25 pm"},
	}
	for _, c := range timeCases {
		t.Run("when "+c.name, func(t *testing.T) {
			_, body := get(t, fakeAsker{answer: ask.Answer{Summary: "s", Sessions: []ask.Session{c.session}, SessionsShownBecause: ask.ShownForSessionQuestion}}, "/ask?q=sign+language")
			t.Run("it should say "+c.want, func(t *testing.T) {
				if !strings.Contains(body, `<span class="when">`+c.want+`</span>`) {
					t.Fatalf("%q missing in %s", c.want, body)
				}
			})
		})
	}

	distanceCases := []struct {
		tier string
		want string
	}{
		{ask.DistanceSameLevel, `<span class="distance">Same level as you</span>`},
		{ask.DistanceSameBuilding, `<span class="distance">Same building, another level</span>`},
		{ask.DistanceOtherBuilding, `<span class="distance">Other building</span>`},
	}
	for _, c := range distanceCases {
		t.Run("when the session is "+c.tier, func(t *testing.T) {
			s := course
			s.Distance = c.tier
			_, body := get(t, fakeAsker{answer: ask.Answer{Summary: "s", Sessions: []ask.Session{s}, SessionsShownBecause: ask.ShownForSessionQuestion}}, "/ask?q=sign+language")
			t.Run("it should say so in the tier words", func(t *testing.T) {
				if !strings.Contains(body, c.want) {
					t.Fatalf("%q missing in %s", c.want, body)
				}
			})
		})
	}

	t.Run("when the attendee's location is unknown", func(t *testing.T) {
		s := course
		s.Distance = ask.DistanceUnknown
		_, body := get(t, fakeAsker{answer: ask.Answer{Summary: "s", Sessions: []ask.Session{s}, SessionsShownBecause: ask.ShownForSessionQuestion}}, "/ask?q=sign+language")
		t.Run("it should say nothing about distance", func(t *testing.T) {
			if strings.Contains(body, `class="distance"`) {
				t.Fatalf("distance line in %s", body)
			}
		})
	})
}

func withTime(s ask.Session, state string, startsIn int) ask.Session {
	s.State = state
	s.StartsInMinutes = startsIn
	return s
}
