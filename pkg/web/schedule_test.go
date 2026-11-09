package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	v1 "github.com/auroq/cnh/pkg/api/v1"
	"github.com/auroq/cnh/pkg/ask"
	"github.com/auroq/cnh/pkg/model"
)

type scheduledAsker struct {
	fakeAsker
	byTitle   map[string]model.Session
	schedule  model.Schedule
	asked     []string
	askedIDs  []string
	askedFrom string
}

func (s *scheduledAsker) MatchSessions(_ context.Context, titles []string) ([]model.Match, error) {
	out := make([]model.Match, 0, len(titles))
	for _, t := range titles {
		m := model.Match{Title: t}
		if session, ok := s.byTitle[t]; ok {
			m.Session = &session
		}
		out = append(out, m)
	}
	return out, s.err
}

func (s *scheduledAsker) Agenda(_ context.Context, ids []string) (model.Schedule, error) {
	s.askedIDs = ids
	return s.schedule, s.err
}

func (s *scheduledAsker) AskScheduled(_ context.Context, question, from string, _ model.Preferences, ids []string) (ask.Answer, error) {
	s.asked = append(s.asked, question)
	s.askedIDs, s.askedFrom = ids, from
	ans := s.answer
	sched := s.schedule
	if from == "" && sched.JustLeft != nil {
		sched.AssumedLocation = sched.JustLeft.Room
		ans.Here = ask.Here{Asked: sched.JustLeft.Room, Room: sched.JustLeft.Room, Venue: sched.JustLeft.Venue, Level: sched.JustLeft.Level, Via: ask.ViaSchedule}
	}
	ans.Schedule = &sched
	return ans, s.err
}

func (c client) getWithSched(target, cookie string) (*http.Response, string) {
	c.t.Helper()
	req, _ := http.NewRequest(http.MethodGet, c.srv.URL+target, nil)
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: schedCookie, Value: cookie})
	}
	return c.do(req)
}

func (c client) postSchedule(form url.Values) (*http.Response, string) {
	c.t.Helper()
	req, _ := http.NewRequest(http.MethodPost, c.srv.URL+"/schedule", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.do(req)
}

func schedCookieOf(res *http.Response) *http.Cookie {
	for _, c := range res.Cookies() {
		if c.Name == schedCookie {
			return c
		}
	}
	return nil
}

func TestSchedulePage(t *testing.T) {
	day := time.Date(2026, 11, 11, 0, 0, 0, 0, time.UTC)
	at := func(h, m int) time.Time { return day.Add(time.Duration(h)*time.Hour + time.Duration(m)*time.Minute) }
	crash := model.Session{ID: "crash", Title: "Sign Language Crash Course", Start: at(16, 0), End: at(16, 35), Location: "Salt Palace | Level 1 | Grand Ballroom J", Room: "Grand Ballroom J", Venue: "Salt Palace", Level: "Level 1"}
	keynote := model.Session{ID: "keynote", Title: "Keynote: Welcome", Start: at(17, 0), End: at(17, 35), Location: "Salt Palace | Level 1 | Grand Ballroom A", Room: "Grand Ballroom A", Venue: "Salt Palace", Level: "Level 1"}
	quiet := ask.Place{Name: "All-attendee quiet room", Venue: "Salt Palace", Level: "Level 2", Verified: true}
	newAsker := func() *scheduledAsker {
		return &scheduledAsker{
			fakeAsker: fakeAsker{answer: ask.Answer{Summary: "All-attendee quiet room: Salt Palace, Level 2.", Places: []ask.Place{quiet}}},
			byTitle:   map[string]model.Session{"Sign Language Crash Course": crash, "Keynote: Welcome": keynote},
			schedule:  model.Schedule{Sessions: []model.Session{crash, keynote}, Next: &keynote, JustLeft: &crash},
		}
	}

	t.Run("when three titles are pasted and one is made up", func(t *testing.T) {
		res, _ := newClient(t, newAsker()).postSchedule(url.Values{"text": {"Sign Language Crash Course\nA talk nobody gave\nKeynote: Welcome"}, "prefs": {"step-free"}})
		location, _ := url.Parse(res.Header.Get("Location"))
		cookie := schedCookieOf(res)
		if cookie == nil {
			t.Fatalf("no %s cookie in %v", schedCookie, res.Header)
		}
		t.Run("it should send the attendee back to the schedule page", func(t *testing.T) {
			if res.StatusCode != http.StatusSeeOther || location.Path != "/schedule" {
				t.Fatalf("%d %s", res.StatusCode, location)
			}
		})
		t.Run("it should carry only the matched ids", func(t *testing.T) {
			if got := location.Query().Get("sched"); got != "crash,keynote" {
				t.Fatalf("sched %q", got)
			}
		})
		t.Run("it should carry the title it could not find", func(t *testing.T) {
			if got := location.Query().Get("missed"); got != "A talk nobody gave" {
				t.Fatalf("missed %q", got)
			}
		})
		t.Run("it should keep the profile", func(t *testing.T) {
			if got := location.Query().Get("prefs"); got != "step-free" {
				t.Fatalf("prefs %q", got)
			}
		})
		t.Run("it should remember the ids in the cookie", func(t *testing.T) {
			if cookie.Value != "crash%2Ckeynote" {
				t.Fatalf("cookie %q", cookie.Value)
			}
		})
		t.Run("it should keep the cookie first-party, strict and until the conference ends", func(t *testing.T) {
			if cookie.SameSite != http.SameSiteStrictMode || !cookie.Secure || !cookie.HttpOnly || !cookie.Expires.Equal(cookieExpiry) {
				t.Fatalf("cookie %+v", cookie)
			}
		})
	})

	t.Run("when the schedule page is opened with the ids in the query", func(t *testing.T) {
		asker := newAsker()
		_, body := newClient(t, asker).get("/schedule?sched=crash,keynote&missed=A+talk+nobody+gave&prefs=step-free", "")
		shows := map[string]string{
			"count the kept sessions in the heading":              "Your schedule: 2 sessions</h2>",
			"announce what was kept and what was not":             `<p class="status" role="status">Your schedule: 2 sessions kept, 1 title not found.</p>`,
			"list a kept session with its time and room":          "<strong>Keynote: Welcome</strong>",
			"list the title it could not find":                    "<h3>Not found</h3>",
			"label the textarea":                                  `<label for="sched-text">Paste your schedule</label>`,
			"describe the textarea":                               `<textarea id="sched-text" name="text" rows="6" aria-describedby="sched-help"`,
			"offer to forget the schedule":                        `name="forget" value="1" class="secondary">Forget my schedule</button>`,
			"carry the schedule in the search form":               `<input type="hidden" name="sched" value="crash,keynote">`,
			"carry the schedule and the profile on the home link": `<a href="/?prefs=step-free&amp;sched=crash%2ckeynote">Wide Door</a>`,
			"say the next session with its room":                  "Your next session: <strong>Keynote: Welcome</strong>, Wed 10:00 am in Grand Ballroom A, Salt Palace Level 1.",
			"say which session was just left":                     "You just left: <strong>Sign Language Crash Course</strong>.",
			"hand the ids to the page script":                     `data-sched="crash,keynote"`,
		}
		for name, want := range shows {
			t.Run("it should "+name, func(t *testing.T) {
				if !strings.Contains(body, want) {
					t.Fatalf("%q missing in %s", want, body)
				}
			})
		}
		t.Run("it should carry the schedule on every quick pick", func(t *testing.T) {
			if n := strings.Count(body, `&amp;prefs=step-free&amp;sched=crash%2ckeynote"`); n != len(quickKinds) {
				t.Fatalf("%d quick picks carry sched, want %d in %s", n, len(quickKinds), body)
			}
		})
		t.Run("it should load exactly the kept sessions", func(t *testing.T) {
			if strings.Join(asker.askedIDs, ",") != "crash,keynote" {
				t.Fatalf("ids %q", asker.askedIDs)
			}
		})
	})

	t.Run("when the page is opened with only the cookie", func(t *testing.T) {
		asker := newAsker()
		_, body := newClient(t, asker).getWithSched("/", "crash%2Ckeynote")
		t.Run("it should read the ids from the cookie", func(t *testing.T) {
			if strings.Join(asker.askedIDs, ",") != "crash,keynote" {
				t.Fatalf("ids %q", asker.askedIDs)
			}
		})
		t.Run("it should not announce anything", func(t *testing.T) {
			if strings.Contains(body, `role="status"`) {
				t.Fatalf("status in %s", body)
			}
		})
	})

	t.Run("when a question is asked with a schedule and no location", func(t *testing.T) {
		asker := newAsker()
		_, body := newClient(t, asker).get("/ask?q=nearest+quiet+room&sched=crash,keynote", "")
		t.Run("it should hand the schedule to the use case", func(t *testing.T) {
			if len(asker.asked) == 0 || asker.asked[0] != "nearest quiet room" || strings.Join(asker.askedIDs, ",") != "crash,keynote" {
				t.Fatalf("asked %q ids %q", asker.asked, asker.askedIDs)
			}
		})
		t.Run("it should say it assumed the room just left", func(t *testing.T) {
			if !strings.Contains(body, "Assuming you are still near <strong>Grand Ballroom J</strong>, the room of the session you just left.") {
				t.Fatalf("assumption missing in %s", body)
			}
		})
	})

	t.Run("when the API is asked with a schedule", func(t *testing.T) {
		_, body := newClient(t, newAsker()).get("/api/ask?q=nearest+quiet+room&sched=crash,keynote", "")
		var got v1.AskResponse
		if err := json.Unmarshal([]byte(body), &got); err != nil {
			t.Fatal(err)
		}
		t.Run("it should return the next session", func(t *testing.T) {
			if got.Schedule == nil || got.Schedule.Next == nil || got.Schedule.Next.ID != "keynote" {
				t.Fatalf("schedule %+v", got.Schedule)
			}
		})
		t.Run("it should return the session just left", func(t *testing.T) {
			if got.Schedule.JustLeft == nil || got.Schedule.JustLeft.ID != "crash" {
				t.Fatalf("schedule %+v", got.Schedule)
			}
		})
		t.Run("it should return the assumed location", func(t *testing.T) {
			if got.Schedule.AssumedLocation != "Grand Ballroom J" || got.Here.Via != ask.ViaSchedule {
				t.Fatalf("schedule %+v here %+v", got.Schedule, got.Here)
			}
		})
	})

	t.Run("when three kept sessions run back to back", func(t *testing.T) {
		asker := newAsker()
		asker.schedule.BackToBack = true
		_, body := newClient(t, asker).get("/?sched=crash,keynote,third", "")
		t.Run("it should offer the nearest quiet room", func(t *testing.T) {
			if !strings.Contains(body, "You have three sessions in a row. If you want a break, the nearest quiet room is All-attendee quiet room, Salt Palace, Level 2.") {
				t.Fatalf("offer missing in %s", body)
			}
		})
		t.Run("it should ask the use case for the quiet room", func(t *testing.T) {
			if len(asker.asked) != 1 || asker.asked[0] != quietRoomAsk {
				t.Fatalf("asked %q", asker.asked)
			}
		})
	})

	t.Run("when two kept sessions do not run back to back", func(t *testing.T) {
		_, body := newClient(t, newAsker()).get("/?sched=crash,keynote", "")
		t.Run("it should not mention a row of sessions", func(t *testing.T) {
			if strings.Contains(body, "sessions in a row") {
				t.Fatalf("offer in %s", body)
			}
		})
	})

	t.Run("when the schedule is forgotten", func(t *testing.T) {
		c := newClient(t, newAsker())
		res, _ := c.postSchedule(url.Values{"forget": {"1"}, "q": {"quiet room"}})
		cookie := schedCookieOf(res)
		if cookie == nil {
			t.Fatalf("no %s cookie in %v", schedCookie, res.Header)
		}
		location, _ := url.Parse(res.Header.Get("Location"))
		_, body := c.get(location.String(), "")
		t.Run("it should expire the cookie", func(t *testing.T) {
			if cookie.MaxAge != -1 || cookie.Value != "" {
				t.Fatalf("cookie %+v", cookie)
			}
		})
		t.Run("it should carry no ids", func(t *testing.T) {
			if location.Query().Has("sched") {
				t.Fatalf("location %s", location)
			}
		})
		t.Run("it should say the schedule is forgotten", func(t *testing.T) {
			if !strings.Contains(body, `<p class="status" role="status">`+scheduleForgot+"</p>") {
				t.Fatalf("status missing in %s", body)
			}
		})
		t.Run("it should tell the page script to forget too", func(t *testing.T) {
			if !strings.Contains(body, `data-forgot-sched="1"`) {
				t.Fatalf("flag missing in %s", body)
			}
		})
	})

	t.Run("when the use case does not understand schedules", func(t *testing.T) {
		res, body := newClient(t, fakeAsker{}).postSchedule(url.Values{"text": {"Keynote: Welcome"}})
		t.Run("it should say the search is not answering", func(t *testing.T) {
			if res.StatusCode != http.StatusBadGateway || !strings.Contains(body, searchDown) {
				t.Fatalf("%d %q", res.StatusCode, body)
			}
		})
	})

	t.Run("when the page is opened without a schedule", func(t *testing.T) {
		_, body := newClient(t, newAsker()).get("/", "")
		t.Run("it should say none is kept yet", func(t *testing.T) {
			if !strings.Contains(body, "Your schedule: none yet</h2>") {
				t.Fatalf("heading missing in %s", body)
			}
		})
		t.Run("it should not offer to forget", func(t *testing.T) {
			if strings.Contains(body, "Forget my schedule") {
				t.Fatalf("forget in %s", body)
			}
		})
	})
}
