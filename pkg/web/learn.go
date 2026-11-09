package web

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"go.opentelemetry.io/otel/metric"

	"github.com/auroq/cnh/pkg/ask"
	"github.com/auroq/cnh/pkg/model"
)

const (
	declinedCookie = "wd_prefs_no"
	keptParam      = "kept"
)

// declined is the per-browser list of chips the attendee removed, which a question may never add back.
func declined(r *http.Request) model.Preferences {
	c, err := r.Cookie(declinedCookie)
	if err != nil {
		return nil
	}
	raw, _ := url.QueryUnescape(c.Value)
	return model.ParsePreferences(raw)
}

func rememberDeclined(w http.ResponseWriter, no model.Preferences) {
	c := &http.Cookie{Name: declinedCookie, Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode}
	if len(no) == 0 {
		c.MaxAge = -1
	} else {
		c.Value = url.QueryEscape(strings.Join(no.IDs(), ","))
		c.Expires = cookieExpiry
	}
	setCookie(w, c)
}

// learn adds the chips the question implied, unless the attendee already holds or has removed them,
// says so in the live region and counts each one.
func (s *Server) learn(ctx context.Context, w http.ResponseWriter, r *http.Request, p *page, prefs model.Preferences, lessons []ask.Lesson) model.Preferences {
	no := declined(r)
	var added []ask.Lesson
	for _, l := range lessons {
		if prefs.Has(l.Chip) || no.Has(l.Chip) {
			continue
		}
		if next := prefs.WithLearned(l.Chip); len(next) > len(prefs) {
			prefs = next
			added = append(added, l)
			s.chipLearned().Add(ctx, 1, metric.WithAttributes(attrChip.String(l.Chip)))
		}
	}
	if len(added) == 0 {
		return prefs
	}
	remember(w, prefs)
	p.Status = learnedSentence(added)
	return prefs
}

func learnedSentence(added []ask.Lesson) string {
	var labels, reasons []string
	for _, l := range added {
		labels = append(labels, labelOf(l.Chip))
		if !hasString(reasons, l.Because) {
			reasons = append(reasons, l.Because)
		}
	}
	them := "it"
	if len(added) > 1 {
		them = "them"
	}
	return "Added " + strings.Join(labels, " and ") + " to what Wide Door knows about you, because " + strings.Join(reasons, " and ") + ". Remove " + them + " if that is wrong."
}

func hasString(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

// setCookie replaces an earlier Set-Cookie for the same name, so a response names each cookie once.
func setCookie(w http.ResponseWriter, c *http.Cookie) {
	var kept []string
	for _, v := range w.Header().Values("Set-Cookie") {
		if !strings.HasPrefix(v, c.Name+"=") {
			kept = append(kept, v)
		}
	}
	w.Header()["Set-Cookie"] = append(kept, c.String())
}
