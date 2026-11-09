package web

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/auroq/cnh/pkg/ask"
	"github.com/auroq/cnh/pkg/model"
)

const (
	learnedStepFree = `<span class="vh">Remove </span><span aria-hidden="true">&#10003; </span>Step-free only<span class="from">, from your question</span>`
	keepStepFree    = `<span class="vh">Keep Step-free only</span>`
	learnedStatus   = `<p class="status" role="status">Added Step-free only to what Wide Door knows about you, because you asked for an elevator. Remove it if that is wrong.</p>`
)

func (c client) getWith(target string, cookies map[string]string) (*http.Response, string) {
	c.t.Helper()
	req, _ := http.NewRequest(http.MethodGet, c.srv.URL+target, nil)
	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}
	return c.do(req)
}

func valueOf(c *http.Cookie) string {
	if c == nil {
		return ""
	}
	got, _ := url.QueryUnescape(c.Value)
	return got
}

func cookieOf(res *http.Response, name string) *http.Cookie {
	for _, c := range res.Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func counts(t *testing.T, reader *sdkmetric.ManualReader, name string) map[string]int64 {
	t.Helper()
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatal(err)
	}
	points := map[string]int64{}
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != name {
				continue
			}
			for _, dp := range m.Data.(metricdata.Sum[int64]).DataPoints {
				chip, _ := dp.Attributes.Value(attrChip)
				points[chip.AsString()] = dp.Value
			}
		}
	}
	return points
}

func TestLearnedChips(t *testing.T) {
	elevator := fakeAsker{answer: ask.Answer{
		Summary: "Elevators: Salt Palace. Nobody has verified this yet.",
		Learned: []model.Lesson{{Chip: "step-free", Because: "you asked for an elevator"}},
	}}

	t.Run("when the nearest elevator is asked for with no profile", func(t *testing.T) {
		reader := sdkmetric.NewManualReader(sdkmetric.WithTemporalitySelector(func(sdkmetric.InstrumentKind) metricdata.Temporality { return metricdata.DeltaTemporality }))
		mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
		t.Cleanup(func() { _ = mp.Shutdown(context.Background()) })
		res, body := newClient(t, elevator, WithMeterProvider(mp)).get("/ask?q=nearest+elevator", "")
		shows := map[string]string{
			"mark the step-free chip as from your question": learnedStepFree,
			"offer to keep the chip as a choice":            keepStepFree,
			"say so where a screen reader hears it":         learnedStatus,
			"carry the learned chip in the search form":     `<input type="hidden" name="prefs" value="step-free~">`,
			"count one thing known":                         "What Wide Door knows about you: 1 thing",
			"still answer the question":                     elevator.answer.Summary,
			"tell the browser to keep the chip":             `data-prefs="step-free~"`,
			"carry the learned chip on the home link":       `<a href="/?prefs=step-free~">Wide Door</a>`,
		}
		for name, want := range shows {
			t.Run("it should "+name, func(t *testing.T) {
				if !strings.Contains(body, want) {
					t.Fatalf("%q missing in %s", want, body)
				}
			})
		}
		t.Run("it should remember the chip as learned in the cookie", func(t *testing.T) {
			if got := valueOf(cookieOf(res, prefsCookie)); got != "step-free~" {
				t.Fatalf("cookie value %q", got)
			}
		})
		t.Run("it should set the profile cookie once", func(t *testing.T) {
			if n := strings.Count(strings.Join(res.Header.Values("Set-Cookie"), "\n"), prefsCookie+"="); n != 1 {
				t.Fatalf("%d %s cookies in %v", n, prefsCookie, res.Header.Values("Set-Cookie"))
			}
		})
		t.Run("it should count the learned chip", func(t *testing.T) {
			if points := counts(t, reader, chipLearnedMetric); len(points) != 1 || points["step-free"] != 1 {
				t.Fatalf("points %v", points)
			}
		})
	})

	t.Run("when the same question comes with the chip already chosen", func(t *testing.T) {
		res, body := newClient(t, elevator).get("/ask?q=nearest+elevator&prefs=step-free", "")
		t.Run("it should keep the chip chosen", func(t *testing.T) {
			if strings.Contains(body, "from your question") {
				t.Fatalf("learned marker in %s", body)
			}
		})
		t.Run("it should announce nothing", func(t *testing.T) {
			if strings.Contains(body, `role="status"`) {
				t.Fatalf("status in %s", body)
			}
		})
		t.Run("it should leave the cookie chosen", func(t *testing.T) {
			if got := valueOf(cookieOf(res, prefsCookie)); got != "step-free" {
				t.Fatalf("cookie value %q", got)
			}
		})
	})

	t.Run("when the same question comes after the chip was removed", func(t *testing.T) {
		res, body := newClient(t, elevator).getWith("/ask?q=nearest+elevator", map[string]string{declinedCookie: "step-free"})
		t.Run("it should not add the chip again", func(t *testing.T) {
			if strings.Contains(body, `<span class="vh">Remove `) {
				t.Fatalf("a chip is on in %s", body)
			}
		})
		t.Run("it should announce nothing", func(t *testing.T) {
			if strings.Contains(body, `role="status"`) {
				t.Fatalf("status in %s", body)
			}
		})
		t.Run("it should set no profile cookie", func(t *testing.T) {
			if cookieOf(res, prefsCookie) != nil {
				t.Fatalf("cookie set: %v", res.Header)
			}
		})
	})

	t.Run("when the question is shown again after a profile edit", func(t *testing.T) {
		_, body := newClient(t, elevator).get("/profile?q=nearest+elevator&removed=step-free", "")
		t.Run("it should learn nothing", func(t *testing.T) {
			if strings.Contains(body, "from your question") {
				t.Fatalf("learned marker in %s", body)
			}
		})
		t.Run("it should report the removal instead", func(t *testing.T) {
			if !strings.Contains(body, "Removed Step-free only from what Wide Door knows about you.") {
				t.Fatalf("no removal status in %s", body)
			}
		})
	})

	t.Run("when the question implies two chips", func(t *testing.T) {
		wheels := fakeAsker{answer: ask.Answer{Summary: "x", Learned: []model.Lesson{
			{Chip: "step-free", Because: "you asked about a wheelchair or scooter"},
			{Chip: "wheels", Because: "you asked about a wheelchair or scooter"},
		}}}
		_, body := newClient(t, wheels).get("/ask?q=wheelchair+ramp", "")
		t.Run("it should say both in one sentence", func(t *testing.T) {
			want := "Added Step-free only and Wide doors and ramps to what Wide Door knows about you, because you asked about a wheelchair or scooter. Remove them if that is wrong."
			if !strings.Contains(body, want) {
				t.Fatalf("%q missing in %s", want, body)
			}
		})
	})
}

func TestLearnedChipForm(t *testing.T) {
	cases := []struct {
		name     string
		form     url.Values
		declined string
		location string
		cookie   string
		noCookie string
	}{
		{"a learned chip is removed", url.Values{"prefs": {"step-free~"}, "remove": {"step-free"}, "q": {"nearest elevator"}}, "", "/profile?q=nearest+elevator&removed=step-free", "", "step-free"},
		{"a chosen chip is removed", url.Values{"prefs": {"step-free,wheels"}, "remove": {"wheels"}}, "", "/profile?prefs=step-free&removed=wheels", "step-free", "wheels"},
		{"a second chip is removed", url.Values{"prefs": {"wheels~"}, "remove": {"wheels"}}, "step-free", "/profile?removed=wheels", "", "step-free,wheels"},
		{"a learned chip is kept", url.Values{"prefs": {"step-free~"}, "add": {"step-free"}}, "", "/profile?kept=step-free&prefs=step-free", "step-free", ""},
		{"a removed chip is added back by hand", url.Values{"prefs": {""}, "add": {"step-free"}}, "step-free,wheels", "/profile?added=step-free&prefs=step-free", "step-free", "wheels"},
		{"the attendee asks to be forgotten", url.Values{"prefs": {"step-free~"}, "forget": {"1"}}, "wheels", "/profile?forgot=1", "", ""},
	}
	for _, c := range cases {
		t.Run("when "+c.name, func(t *testing.T) {
			cl := newClient(t, fakeAsker{})
			req, _ := http.NewRequest(http.MethodPost, cl.srv.URL+"/profile", strings.NewReader(c.form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			if c.declined != "" {
				req.AddCookie(&http.Cookie{Name: declinedCookie, Value: c.declined})
			}
			res, _ := cl.do(req)
			t.Run("it should send the attendee to "+c.location, func(t *testing.T) {
				if got := res.Header.Get("Location"); got != c.location {
					t.Fatalf("location %q", got)
				}
			})
			t.Run("it should leave the profile cookie as "+quoted(c.cookie), func(t *testing.T) {
				cookie := cookieOf(res, prefsCookie)
				if cookie == nil || valueOf(cookie) != c.cookie || (c.cookie == "" && cookie.MaxAge >= 0) {
					t.Fatalf("cookie %+v", cookie)
				}
			})
			t.Run("it should leave the do-not-re-add cookie as "+quoted(c.noCookie), func(t *testing.T) {
				cookie := cookieOf(res, declinedCookie)
				if cookie == nil || valueOf(cookie) != c.noCookie || (c.noCookie == "" && cookie.MaxAge >= 0) {
					t.Fatalf("cookie %+v", cookie)
				}
			})
		})
	}

	t.Run("when the page is reopened after a learned chip was kept", func(t *testing.T) {
		_, body := newClient(t, fakeAsker{}).get("/profile?kept=step-free&prefs=step-free", "")
		t.Run("it should say so where a screen reader hears it", func(t *testing.T) {
			want := `<p class="status" role="status">Kept Step-free only as your own choice, not just from your question.</p>`
			if !strings.Contains(body, want) {
				t.Fatalf("%q missing in %s", want, body)
			}
		})
	})
}

func quoted(s string) string {
	if s == "" {
		return "nothing"
	}
	return `"` + s + `"`
}
