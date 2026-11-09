package web

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/metric"

	"github.com/auroq/cnh/pkg/ask"
	"github.com/auroq/cnh/pkg/model"
)

// Profiled is the optional half of an Asker that shapes the answer by the attendee's profile.
// An Asker without it still answers; the chips then only travel with the page.
type Profiled interface {
	AskFor(ctx context.Context, question, from string, prefs model.Preferences) (ask.Answer, error)
}

const (
	prefsCookie  = "wd_prefs"
	prefsParam   = "prefs"
	forgotParam  = "forgot"
	addedParam   = "added"
	removedParam = "removed"
	profilePath  = "/profile"
	forgotten    = "Forgotten. Nothing about you is on this device now."
)

// cookieExpiry is the end of the conference, Thursday 2026-11-12 23:59 Mountain Time.
var cookieExpiry = time.Date(2026, 11, 12, 23, 59, 0, 0, time.FixedZone("MST", -7*3600))

type chip struct {
	ID      string
	Label   string
	On      bool
	Learned bool
}

// preferences reads the profile for this request: the query wins and refreshes the cookie, else the cookie.
func (s *Server) preferences(w http.ResponseWriter, r *http.Request) model.Preferences {
	if r.URL.Query().Has(prefsParam) {
		prefs := model.ParsePreferences(r.URL.Query().Get(prefsParam))
		remember(w, prefs)
		return prefs
	}
	c, err := r.Cookie(prefsCookie)
	if err != nil {
		return nil
	}
	raw, _ := url.QueryUnescape(c.Value)
	return model.ParsePreferences(raw)
}

func remember(w http.ResponseWriter, prefs model.Preferences) {
	c := &http.Cookie{Name: prefsCookie, Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode}
	if len(prefs) == 0 {
		c.MaxAge = -1
	} else {
		c.Value = url.QueryEscape(prefs.String())
		c.Expires = cookieExpiry
	}
	setCookie(w, c)
}

func chips(prefs model.Preferences) []chip {
	out := make([]chip, 0, len(model.Chips))
	for _, c := range model.Chips {
		out = append(out, chip{ID: c.ID, Label: c.Label, On: prefs.Has(c.ID), Learned: prefs.IsLearned(c.ID)})
	}
	return out
}

// profile is the no-JavaScript path: one POST per chip change, then back to the page with the panel open.
func (s *Server) profile(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	prefs := model.ParsePreferences(r.PostForm.Get(prefsParam))
	values := url.Values{}
	for _, key := range []string{"q", "from"} {
		if v := strings.TrimSpace(r.PostForm.Get(key)); v != "" {
			values.Set(key, v)
		}
	}
	no := declined(r)
	if id := r.PostForm.Get("add"); id != "" && (!prefs.Has(id) || prefs.IsLearned(id)) {
		was := prefs.IsLearned(id)
		prefs = prefs.With(id)
		no = no.Without(id)
		switch {
		case was:
			values.Set(keptParam, id)
		case prefs.Has(id):
			values.Set(addedParam, id)
		}
	}
	if id := r.PostForm.Get("remove"); prefs.Has(id) {
		prefs = prefs.Without(id)
		no = no.With(id)
		values.Set(removedParam, id)
	}
	if r.PostForm.Get("forget") != "" {
		prefs, no = nil, nil
		values.Set(forgotParam, "1")
	}
	if len(prefs) > 0 {
		values.Set(prefsParam, prefs.String())
	}
	remember(w, prefs)
	rememberDeclined(w, no)
	http.Redirect(w, r, profilePath+"?"+values.Encode(), http.StatusSeeOther)
}

// describeProfile fills the page with what the attendee sees of their profile and what just changed.
func (s *Server) describeProfile(p *page, r *http.Request, prefs model.Preferences) {
	p.Prefs = prefs.String()
	p.Chips = chips(prefs)
	p.Count = countWords(len(prefs))
	p.ProfileOpen = r.URL.Path == profilePath
	p.PrefsKnown = profileKnown(r)
	p.EditURL = editURL(r, prefs)
	if len(prefs) > 0 {
		p.ProfileURL = profileURL(r, prefs)
	}
	q := r.URL.Query()
	switch {
	case q.Get(forgotParam) != "":
		p.Status, p.Forgot = forgotten, true
	case q.Get(addedParam) != "":
		p.Status = "Added " + labelOf(q.Get(addedParam)) + " to what Wide Door knows about you."
	case q.Get(keptParam) != "":
		p.Status = "Kept " + labelOf(q.Get(keptParam)) + " as your own choice, not just from your question."
	case q.Get(removedParam) != "":
		p.Status = "Removed " + labelOf(q.Get(removedParam)) + " from what Wide Door knows about you."
	}
}

// profileKnown says the server rendered the profile from what the browser sent or from an edit it just made,
// so the browser's own copy must follow the page and never restore over it.
func profileKnown(r *http.Request) bool {
	q := r.URL.Query()
	if q.Has(prefsParam) {
		return true
	}
	if r.URL.Path == profilePath {
		for _, p := range []string{addedParam, keptParam, removedParam, forgotParam} {
			if q.Has(p) {
				return true
			}
		}
	}
	for _, name := range []string{prefsCookie, declinedCookie} {
		if _, err := r.Cookie(name); err == nil {
			return true
		}
	}
	return false
}

func countWords(n int) string {
	switch n {
	case 0:
		return "nothing yet"
	case 1:
		return "1 thing"
	default:
		return strconv.Itoa(n) + " things"
	}
}

func labelOf(id string) string {
	for _, c := range model.Chips {
		if c.ID == id {
			return c.Label
		}
	}
	return id
}

func editURL(r *http.Request, prefs model.Preferences) string {
	values := url.Values{}
	for _, key := range []string{"q", "from"} {
		if v := strings.TrimSpace(r.URL.Query().Get(key)); v != "" {
			values.Set(key, v)
		}
	}
	if len(prefs) > 0 {
		values.Set(prefsParam, prefs.String())
	}
	if len(values) == 0 {
		return profilePath
	}
	return profilePath + "?" + values.Encode()
}

func profileURL(r *http.Request, prefs model.Preferences) string {
	scheme := "https"
	if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") == "" {
		scheme = "http"
	}
	return scheme + "://" + r.Host + "/?" + prefsParam + "=" + prefs.String()
}

// askFor hands the profile to a use case that understands it, or asks plainly when it does not.
func (s *Server) askFor(ctx context.Context, question, from string, prefs model.Preferences) (ask.Answer, error) {
	if p, ok := s.ask.(Profiled); ok && len(prefs) > 0 {
		return p.AskFor(ctx, question, from, prefs)
	}
	return s.ask.Ask(ctx, question, from)
}

func (s *Server) countChips(ctx context.Context, prefs model.Preferences) {
	for _, id := range prefs.IDs() {
		s.chipUsed().Add(ctx, 1, metric.WithAttributes(attrChip.String(id)))
	}
}
