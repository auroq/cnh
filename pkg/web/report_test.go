package web

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/auroq/cnh/pkg/ask"
	"github.com/auroq/cnh/pkg/model"
	"github.com/auroq/cnh/pkg/reports"
)

type fakeReporter struct {
	got reports.Submission
	err error
}

func (f *fakeReporter) Submit(_ context.Context, sub reports.Submission) (reports.Outcome, error) {
	f.got = sub
	return reports.Outcome{ID: "r1", Kind: "elevator"}, f.err
}

func postReport(t *testing.T, rep Reporter, form url.Values, cookie *http.Cookie) *http.Response {
	t.Helper()
	srv := httptest.NewServer(New(fakeAsker{}, slog.New(slog.NewTextHandler(io.Discard, nil)), WithReporter(rep)).Handler())
	t.Cleanup(srv.Close)
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/report", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Forwarded-For", "203.0.113.9, 10.0.0.1")
	if cookie != nil {
		req.AddCookie(cookie)
	}
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { res.Body.Close() })
	return res
}

func cookieNamed(res *http.Response, name string) *http.Cookie {
	for _, c := range res.Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func TestReport(t *testing.T) {
	form := url.Values{"place": {"elevators-hyatt"}, "verdict": {"problem"}, "note": {"out of order"}, "q": {"nearest elevator"}, "from": {"Grand Ballroom J"}}

	t.Run("when a first-time attendee posts a problem report", func(t *testing.T) {
		rep := &fakeReporter{}
		res := postReport(t, rep, form, nil)
		cookie := cookieNamed(res, TokenCookie)
		t.Run("it should hand the use case the form, the new token and the client IP", func(t *testing.T) {
			want := reports.Submission{PlaceID: "elevators-hyatt", Verdict: model.VerdictProblem, Note: "out of order", Token: cookie.Value, IP: "203.0.113.9"}
			if rep.got != want {
				t.Fatalf("got %+v, want %+v", rep.got, want)
			}
		})
		t.Run("it should set the wd_id cookie HttpOnly and SameSite strict, expiring 2026-11-12", func(t *testing.T) {
			if cookie == nil || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode || cookie.Expires.UTC().Format("2006-01-02") != "2026-11-12" || len(cookie.Value) != 32 {
				t.Fatalf("cookie %+v", cookie)
			}
		})
		t.Run("it should redirect back to the answer with the result announced", func(t *testing.T) {
			if res.StatusCode != http.StatusSeeOther || res.Header.Get("Location") != "/ask?from=Grand+Ballroom+J&q=nearest+elevator&reported=ok#answer" {
				t.Fatalf("%d %s", res.StatusCode, res.Header.Get("Location"))
			}
		})
	})

	t.Run("when the attendee already has a token", func(t *testing.T) {
		rep := &fakeReporter{}
		res := postReport(t, rep, form, &http.Cookie{Name: TokenCookie, Value: "existing-token"})
		t.Run("it should reuse it", func(t *testing.T) {
			if rep.got.Token != "existing-token" {
				t.Fatalf("token %q", rep.got.Token)
			}
		})
		t.Run("it should not set a new cookie", func(t *testing.T) {
			if cookieNamed(res, TokenCookie) != nil {
				t.Fatalf("cookies %v", res.Cookies())
			}
		})
	})

	refusals := []struct {
		err  error
		code string
	}{
		{reports.ErrRateLimited, "rate_limited"},
		{reports.ErrNoteLength, "note_length"},
		{reports.ErrNoteURL, "note_url"},
		{reports.ErrNoteBlocked, "note_blocked"},
		{reports.ErrUnknownPlace, "unknown_place"},
		{errors.New("elasticsearch down"), "error"},
	}
	for _, c := range refusals {
		t.Run("when the use case refuses with "+c.err.Error(), func(t *testing.T) {
			res := postReport(t, &fakeReporter{err: c.err}, form, nil)
			t.Run("it should redirect with the reason "+c.code, func(t *testing.T) {
				if !strings.Contains(res.Header.Get("Location"), "reported="+c.code) {
					t.Fatalf("location %s", res.Header.Get("Location"))
				}
			})
		})
	}

	t.Run("when the honeypot caught a bot", func(t *testing.T) {
		res := postReport(t, &fakeReporter{err: reports.ErrHoneypot}, form, nil)
		t.Run("it should pretend the report was accepted", func(t *testing.T) {
			if !strings.Contains(res.Header.Get("Location"), "reported=ok") {
				t.Fatalf("location %s", res.Header.Get("Location"))
			}
		})
	})

	t.Run("when no reporter is wired", func(t *testing.T) {
		srv := httptest.NewServer(New(fakeAsker{}, slog.New(slog.NewTextHandler(io.Discard, nil))).Handler())
		t.Cleanup(srv.Close)
		res, err := http.PostForm(srv.URL+"/report", form)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { res.Body.Close() })
		t.Run("it should answer not found", func(t *testing.T) {
			if res.StatusCode != http.StatusNotFound {
				t.Fatalf("%d", res.StatusCode)
			}
		})
	})
}

func TestReportOnThePage(t *testing.T) {
	reportedAt := time.Date(2026, 11, 9, 21, 52, 0, 0, time.UTC)
	answered := fakeAsker{answer: ask.Answer{
		Summary: "Elevators: Hyatt Regency. Unverified: read off a plan, nobody has checked it yet. Reported 2:52 pm by one attendee: out of order.",
		Places: []ask.Place{{ID: "elevators-hyatt", Name: "Elevators", Venue: "Hyatt Regency",
			Attendee: &model.AttendeeStatus{Status: model.StatusProblem, Count: 1, LatestNote: "out of order", LatestAt: reportedAt}}},
	}}
	withForm := func(t *testing.T, target string) string {
		t.Helper()
		srv := httptest.NewServer(New(answered, slog.New(slog.NewTextHandler(io.Discard, nil)), WithReporter(&fakeReporter{})).Handler())
		t.Cleanup(srv.Close)
		res, err := http.Get(srv.URL + target)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { res.Body.Close() })
		body, _ := io.ReadAll(res.Body)
		return string(body)
	}

	t.Run("when a place has a problem report", func(t *testing.T) {
		body := withForm(t, "/ask?q=nearest+elevator")
		t.Run("it should show the report on the place as a badge", func(t *testing.T) {
			if !strings.Contains(body, `<span class="badge problem">Reported 2:52 pm by one attendee: out of order.</span>`) {
				t.Fatalf("badge missing in %s", body)
			}
		})
		t.Run("it should offer a form that posts without JavaScript", func(t *testing.T) {
			if !strings.Contains(body, `<form method="post" action="/report"`) {
				t.Fatalf("form missing in %s", body)
			}
		})
		controls := map[string]string{
			"the works radio":   `<input type="radio" id="verdict-works-elevators-hyatt" name="verdict" value="works" required>`,
			"its label":         `<label for="verdict-works-elevators-hyatt">Works as described</label>`,
			"the problem radio": `<input type="radio" id="verdict-problem-elevators-hyatt" name="verdict" value="problem">`,
			"the note textarea": `<textarea id="note-elevators-hyatt" name="note"`,
			"the note label":    `<label for="note-elevators-hyatt">`,
			"the honeypot":      `name="website" tabindex="-1"`,
			"the place id":      `<input type="hidden" name="place" value="elevators-hyatt">`,
		}
		for name, want := range controls {
			t.Run("it should label "+name, func(t *testing.T) {
				if !strings.Contains(body, want) {
					t.Fatalf("%s missing in %s", want, body)
				}
			})
		}
	})

	t.Run("when the attendee comes back from posting a report", func(t *testing.T) {
		body := withForm(t, "/ask?q=nearest+elevator&reported=ok")
		t.Run("it should announce the result inside the live region", func(t *testing.T) {
			if !strings.Contains(body, `<p class="reported" role="status">Thank you. Your report is live and other attendees can see it now.</p>`) {
				t.Fatalf("message missing in %s", body)
			}
		})
	})

	t.Run("when the report was rate limited", func(t *testing.T) {
		body := withForm(t, "/ask?q=nearest+elevator&reported=rate_limited")
		t.Run("it should say so politely", func(t *testing.T) {
			if !strings.Contains(body, "Please wait a little before sending another.") {
				t.Fatalf("message missing in %s", body)
			}
		})
	})

	t.Run("when the reported code is not one we know", func(t *testing.T) {
		body := withForm(t, "/ask?q=nearest+elevator&reported=%3Cscript%3E")
		t.Run("it should fall back to the generic message and never echo the code", func(t *testing.T) {
			if strings.Contains(body, "<script>") || !strings.Contains(body, "The report could not be saved right now.") {
				t.Fatalf("body %s", body)
			}
		})
	})
}
