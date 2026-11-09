package web

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/auroq/cnh/pkg/ask"
	"github.com/auroq/cnh/pkg/model"
	"github.com/auroq/cnh/pkg/reports"
)

// Reporter is what the web layer needs to take "Were you here?" reports.
type Reporter interface {
	Submit(ctx context.Context, sub reports.Submission) (reports.Outcome, error)
}

// WithReporter turns the report form on; without it the page shows no form and POST /report answers 404.
func WithReporter(r Reporter) Option {
	return func(s *Server) { s.reporter = r }
}

const (
	TokenCookie   = "wd_id"
	reportTimeout = 10 * time.Second
	reportedOK    = "ok"
)

var tokenExpires = time.Date(2026, 11, 12, 23, 59, 59, 0, time.UTC)

// reportMessages is what the live region says after a report, by the reason the use case gave; never the note.
var reportMessages = map[string]string{
	reportedOK:      "Thank you. Your report is live and other attendees can see it now.",
	"rate_limited":  "You have sent several reports in the last few minutes. Please wait a little before sending another.",
	"note_required": "Please say what you found, in 10 to 280 characters, so others know what to expect.",
	"note_length":   "Please keep the note between 10 and 280 characters.",
	"note_url":      "Please leave links out of the note.",
	"note_blocked":  "That note could not be posted. Please describe the place, not people.",
	"unknown_place": "We could not find that place, so the report was not saved.",
	"verdict":       "Please pick one of the three options.",
	"error":         "The report could not be saved right now. Please try again in a minute.",
}

type reportView struct {
	Place    model.Place
	Question string
	From     string
}

func reportFor(p page, place model.Place) reportView {
	return reportView{Place: place, Question: p.Question, From: p.From}
}

func (s *Server) report(w http.ResponseWriter, r *http.Request) {
	if s.reporter == nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	token := s.token(w, r)
	ctx, cancel := context.WithTimeout(r.Context(), reportTimeout)
	defer cancel()
	_, err := s.reporter.Submit(ctx, reports.Submission{
		PlaceID:  strings.TrimSpace(r.PostForm.Get("place")),
		Verdict:  model.Verdict(r.PostForm.Get("verdict")),
		Note:     r.PostForm.Get("note"),
		Token:    token,
		IP:       clientIP(r),
		Honeypot: r.PostForm.Get("website"),
	})
	code := reportedOK
	switch {
	case errors.Is(err, reports.ErrHoneypot):
	case err != nil:
		code = reports.Reason(err)
		if code == "error" {
			s.log.ErrorContext(ctx, "report", "err", err.Error())
		}
	}
	q := url.Values{"q": {r.PostForm.Get("q")}, "from": {r.PostForm.Get("from")}, "reported": {code}}
	http.Redirect(w, r, "/ask?"+q.Encode()+"#answer", http.StatusSeeOther)
}

func (s *Server) token(w http.ResponseWriter, r *http.Request) string {
	if c, err := r.Cookie(TokenCookie); err == nil && c.Value != "" {
		return c.Value
	}
	token := reports.NewToken()
	http.SetCookie(w, &http.Cookie{
		Name:     TokenCookie,
		Value:    token,
		Path:     "/",
		Expires:  tokenExpires,
		HttpOnly: true,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteStrictMode,
	})
	return token
}

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.TrimSpace(strings.Split(fwd, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func reportMessage(code string) string {
	if code == "" {
		return ""
	}
	if msg, ok := reportMessages[code]; ok {
		return msg
	}
	return reportMessages["error"]
}

func attendeeBadge(p model.Place) string {
	if p.Attendee == nil {
		return ""
	}
	switch p.Attendee.Status {
	case model.StatusProblem:
		return "problem"
	case model.StatusDisputed:
		return "warn"
	case model.StatusVerified:
		return "ok"
	default:
		return "note"
	}
}

var _ = ask.AttendeeSentence
