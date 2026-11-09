package web

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"

	"github.com/auroq/cnh/pkg/ask"
	"github.com/auroq/cnh/pkg/model"
)

// Scheduled is the optional half of an Asker that understands the attendee's own schedule.
// An Asker without it still answers; the schedule page then says it cannot match titles.
type Scheduled interface {
	MatchSessions(ctx context.Context, titles []string) ([]model.Match, error)
	Agenda(ctx context.Context, ids []string) (model.Schedule, error)
	AskScheduled(ctx context.Context, question, from string, prefs model.Preferences, ids []string) (ask.Answer, error)
}

const (
	schedCookie      = "wd_sched"
	schedParam       = "sched"
	missedParam      = "missed"
	forgotSchedParam = "forgot_sched"
	schedulePath     = "/schedule"
	scheduleForgot   = "Your schedule is forgotten. Nothing about it is on this device now."
	quietRoomAsk     = "nearest quiet room"
	matchedMetric    = "widedoor.schedule.matched"
	maxTitles        = 100
	maxPasteBytes    = 64 << 10
)

// scheduleIDs reads the kept session ids for this request: the query wins and refreshes the cookie, else the cookie.
func (s *Server) scheduleIDs(w http.ResponseWriter, r *http.Request) []string {
	if r.URL.Query().Has(schedParam) {
		ids := ask.ParseScheduleIDs(r.URL.Query().Get(schedParam))
		rememberSchedule(w, ids)
		return ids
	}
	c, err := r.Cookie(schedCookie)
	if err != nil {
		return nil
	}
	raw, _ := url.QueryUnescape(c.Value)
	return ask.ParseScheduleIDs(raw)
}

func rememberSchedule(w http.ResponseWriter, ids []string) {
	c := &http.Cookie{Name: schedCookie, Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode}
	if len(ids) == 0 {
		c.MaxAge = -1
	} else {
		c.Value = url.QueryEscape(ask.JoinScheduleIDs(ids))
		c.Expires = cookieExpiry
	}
	http.SetCookie(w, c)
}

// schedule is the paste: titles or an iCal export in, the matched ids kept on the device, then back to the page.
func (s *Server) schedule(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxPasteBytes)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	values := url.Values{}
	for _, key := range []string{"q", "from", prefsParam} {
		if v := strings.TrimSpace(r.PostForm.Get(key)); v != "" {
			values.Set(key, v)
		}
	}
	if r.PostForm.Get("forget") != "" {
		rememberSchedule(w, nil)
		values.Set(forgotSchedParam, "1")
		http.Redirect(w, r, schedulePath+"?"+values.Encode(), http.StatusSeeOther)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), askTimeout)
	defer cancel()
	titles := ask.ParseTitles(r.PostForm.Get("text"))
	if len(titles) > maxTitles {
		titles = titles[:maxTitles]
	}
	matches, err := s.matchSessions(ctx, titles)
	if err != nil {
		s.log.ErrorContext(ctx, "schedule", "err", err.Error())
		http.Error(w, searchDown, http.StatusBadGateway)
		return
	}
	var ids []string
	missed := 0
	for _, m := range matches {
		if m.Session != nil {
			ids = append(ids, m.Session.ID)
		} else {
			missed++
		}
	}
	ids = ask.ParseScheduleIDs(ask.JoinScheduleIDs(ids))
	s.matched().Add(ctx, int64(len(ids)))
	rememberSchedule(w, ids)
	if len(ids) > 0 {
		values.Set(schedParam, ask.JoinScheduleIDs(ids))
	}
	if missed > 0 {
		values.Set(missedParam, strconv.Itoa(missed))
	}
	http.Redirect(w, r, schedulePath+"?"+values.Encode(), http.StatusSeeOther)
}

func (s *Server) matchSessions(ctx context.Context, titles []string) ([]model.Match, error) {
	sch, ok := s.ask.(Scheduled)
	if !ok {
		return nil, ask.ErrNoSessionFinder
	}
	return sch.MatchSessions(ctx, titles)
}

// askScheduled hands the schedule to a use case that understands it, or falls back to the profile alone.
func (s *Server) askScheduled(ctx context.Context, question, from string, prefs model.Preferences, ids []string) (ask.Answer, error) {
	if sch, ok := s.ask.(Scheduled); ok && len(ids) > 0 {
		return sch.AskScheduled(ctx, question, from, prefs, ids)
	}
	return s.askFor(ctx, question, from, prefs)
}

// describeSchedule fills the page with what the attendee sees of their schedule: the lines at the top, the kept list, what just changed.
func (s *Server) describeSchedule(ctx context.Context, p *page, r *http.Request, ids []string, prefs model.Preferences) {
	p.Sched = ask.JoinScheduleIDs(ids)
	p.ScheduleOpen = r.URL.Path == schedulePath
	q := r.URL.Query()
	if q.Get(forgotSchedParam) != "" {
		p.Status, p.ForgotSched = scheduleForgot, true
	}
	p.Missed, _ = strconv.Atoi(q.Get(missedParam))
	if len(ids) == 0 {
		p.ScheduleCount = "none yet"
		return
	}
	if p.Answer != nil && p.Answer.Schedule != nil {
		p.Schedule = p.Answer.Schedule
	} else if sch, ok := s.ask.(Scheduled); ok {
		sched, err := sch.Agenda(ctx, ids)
		if err != nil {
			s.log.ErrorContext(ctx, "agenda", "err", err.Error())
		} else {
			p.Schedule = &sched
		}
	}
	if p.Schedule == nil {
		p.ScheduleCount = countSessions(len(ids))
		return
	}
	p.ScheduleCount = countSessions(len(p.Schedule.Sessions))
	if p.ScheduleOpen && q.Has(schedParam) {
		p.Status = "Your schedule: " + p.ScheduleCount + " kept" + missedWords(p.Missed) + "."
	}
	if p.Schedule.BackToBack {
		p.QuietRoom = s.quietRoom(ctx, p.From, prefs, ids)
	}
}

func (s *Server) quietRoom(ctx context.Context, from string, prefs model.Preferences, ids []string) string {
	ans, err := s.askScheduled(ctx, quietRoomAsk, from, prefs, ids)
	if err != nil || len(ans.Places) == 0 {
		return ""
	}
	p := ans.Places[0]
	where := p.Venue
	if p.Level != "" {
		where += ", " + p.Level
	}
	return p.Name + ", " + where
}

func countSessions(n int) string {
	switch n {
	case 0:
		return "none yet"
	case 1:
		return "1 session"
	default:
		return fmt.Sprintf("%d sessions", n)
	}
}

func missedWords(n int) string {
	switch n {
	case 0:
		return ""
	case 1:
		return ", 1 title not found"
	default:
		return fmt.Sprintf(", %d titles not found", n)
	}
}

func (s *Server) matched() metric.Int64Counter {
	s.schedOnce.Do(func() {
		mp := s.meter
		if mp == nil {
			mp = otel.GetMeterProvider()
		}
		s.matchedCounter, _ = mp.Meter(scope).Int64Counter(matchedMetric,
			metric.WithDescription("Sessions matched from pasted schedules; a count only, never which sessions or whose"))
	})
	return s.matchedCounter
}
