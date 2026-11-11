package web

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/auroq/cnh/pkg/ask"
	"github.com/auroq/cnh/pkg/model"
)

type profiledAsker struct {
	fakeAsker
	prefs model.Preferences
}

func (p *profiledAsker) AskFor(_ context.Context, _, _ string, prefs model.Preferences) (ask.Answer, error) {
	p.prefs = prefs
	return p.answer, p.err
}

type client struct {
	t   *testing.T
	srv *httptest.Server
	log *bytes.Buffer
}

func newClient(t *testing.T, a Asker, opts ...Option) client {
	t.Helper()
	log := &bytes.Buffer{}
	srv := httptest.NewServer(New(a, slog.New(slog.NewTextHandler(log, nil)), opts...).Handler())
	t.Cleanup(srv.Close)
	return client{t: t, srv: srv, log: log}
}

func (c client) do(req *http.Request) (*http.Response, string) {
	c.t.Helper()
	res, err := (&http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}).Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	c.t.Cleanup(func() { res.Body.Close() })
	body, _ := io.ReadAll(res.Body)
	return res, string(body)
}

func (c client) get(target string, cookie string) (*http.Response, string) {
	c.t.Helper()
	req, _ := http.NewRequest(http.MethodGet, c.srv.URL+target, nil)
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: prefsCookie, Value: cookie})
	}
	return c.do(req)
}

func (c client) post(form url.Values) (*http.Response, string) {
	c.t.Helper()
	req, _ := http.NewRequest(http.MethodPost, c.srv.URL+"/profile", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.do(req)
}

func prefsCookieOf(res *http.Response) *http.Cookie {
	for _, c := range res.Cookies() {
		if c.Name == prefsCookie {
			return c
		}
	}
	return nil
}

const (
	removeStepFree = `<span class="vh">Remove </span><span aria-hidden="true">&#10003; </span><span>Step-free only`
	addQuiet       = `<span class="vh">Add </span><span>Quiet spaces matter`
	profileHeading = "What Wide Door knows about you"
)

func TestProfileBar(t *testing.T) {
	answered := fakeAsker{answer: ask.Answer{Summary: "Quiet room 260A: Salt Palace, Level 2."}}

	t.Run("when the page is opened with step-free and wheels in the query", func(t *testing.T) {
		res, body := newClient(t, answered).get("/ask?q=quiet+room&from=255+BC&prefs=step-free,wheels", "")
		cookie := prefsCookieOf(res)
		if cookie == nil {
			t.Fatalf("no %s cookie in %v", prefsCookie, res.Header)
		}
		shows := map[string]string{
			"show the profile heading":                        profileHeading,
			"offer to remove a chip that is on":               removeStepFree,
			"offer to add a chip that is off":                 addQuiet,
			"offer to be forgotten":                           `name="forget"`,
			"offer the profile as a link with only the chips": `/?prefs=step-free,wheels`,
			"carry the profile in the search form":            `<input type="hidden" name="prefs" value="step-free,wheels">`,
			"carry the profile on the home link":              `<a href="/?prefs=step-free%2cwheels">Wide Door</a>`,
			"state the new disclosure sentence":               "Stored on this device until Thursday, sent with every page, never kept by the server. We never ask what you have, only what you need.",
		}
		for name, want := range shows {
			t.Run("it should "+name, func(t *testing.T) {
				if !strings.Contains(body, want) {
					t.Fatalf("%q missing in %s", want, body)
				}
			})
		}
		t.Run("it should carry the profile on every quick pick", func(t *testing.T) {
			if n := strings.Count(body, `&amp;prefs=step-free%2cwheels"`); n != len(quickKinds) {
				t.Fatalf("%d quick picks carry prefs, want %d in %s", n, len(quickKinds), body)
			}
		})
		t.Run("it should remember the chips in the cookie", func(t *testing.T) {
			if got, _ := url.QueryUnescape(cookie.Value); got != "step-free,wheels" {
				t.Fatalf("cookie value %q", cookie.Value)
			}
		})
		t.Run("it should keep the cookie first-party and off the wire in clear", func(t *testing.T) {
			if !cookie.Secure || cookie.SameSite != http.SameSiteStrictMode || cookie.Path != "/" {
				t.Fatalf("cookie %+v", cookie)
			}
		})
		t.Run("it should let the cookie expire when the conference ends", func(t *testing.T) {
			if !cookie.Expires.Equal(cookieExpiry) {
				t.Fatalf("expires %s", cookie.Expires)
			}
		})
	})

	t.Run("when the page is opened with only the cookie", func(t *testing.T) {
		res, body := newClient(t, answered).get("/", url.QueryEscape("step-free"))
		t.Run("it should render the chip from the cookie", func(t *testing.T) {
			if !strings.Contains(body, removeStepFree) {
				t.Fatalf("chip missing in %s", body)
			}
		})
		t.Run("it should not set the cookie again", func(t *testing.T) {
			if prefsCookieOf(res) != nil {
				t.Fatalf("cookie set: %v", res.Header)
			}
		})
	})

	t.Run("when the query says no chips and the cookie says one", func(t *testing.T) {
		res, body := newClient(t, answered).get("/?prefs=", url.QueryEscape("step-free"))
		t.Run("it should let the query win", func(t *testing.T) {
			if strings.Contains(body, "<span class=\"vh\">Remove ") {
				t.Fatalf("a chip is on in %s", body)
			}
		})
		t.Run("it should clear the cookie", func(t *testing.T) {
			if c := prefsCookieOf(res); c == nil || c.MaxAge >= 0 {
				t.Fatalf("cookie %+v", c)
			}
		})
	})

	t.Run("when the page is opened with no profile anywhere", func(t *testing.T) {
		res, body := newClient(t, answered).get("/ask?q=quiet+room", "")
		absent := map[string]string{
			"carry no prefs on the home link":  "?prefs=",
			"carry no prefs on any quick pick": "&amp;prefs=",
			"turn no chip on":                  `<span class="vh">Remove `,
			"offer nothing to forget":          `name="forget"`,
			"offer no link to save":            "Save my profile",
		}
		for name, unwanted := range absent {
			t.Run("it should "+name, func(t *testing.T) {
				if strings.Contains(body, unwanted) {
					t.Fatalf("%q present in %s", unwanted, body)
				}
			})
		}
		t.Run("it should still offer every chip", func(t *testing.T) {
			if n := strings.Count(body, `<span class="vh">Add </span>`); n != len(model.Chips) {
				t.Fatalf("%d chips offered, want %d", n, len(model.Chips))
			}
		})
		t.Run("it should set no cookie", func(t *testing.T) {
			if prefsCookieOf(res) != nil {
				t.Fatalf("cookie set: %v", res.Header)
			}
		})
	})
}

func TestProfileForm(t *testing.T) {
	cases := []struct {
		name     string
		form     url.Values
		location string
		cookie   string
	}{
		{"a chip is added while asking a question", url.Values{"prefs": {"step-free"}, "add": {"wheels"}, "q": {"quiet room"}, "from": {"255 BC"}}, "/profile?added=wheels&from=255+BC&prefs=step-free%2Cwheels&q=quiet+room", "step-free,wheels"},
		{"a chip is removed on the home page", url.Values{"prefs": {"step-free,wheels"}, "remove": {"wheels"}}, "/profile?prefs=step-free&removed=wheels", "step-free"},
		{"the last chip is removed", url.Values{"prefs": {"wheels"}, "remove": {"wheels"}}, "/profile?removed=wheels", ""},
		{"an unknown chip is added", url.Values{"prefs": {"wheels"}, "add": {"diagnosis"}}, "/profile?prefs=wheels", "wheels"},
		{"the attendee asks to be forgotten", url.Values{"prefs": {"step-free,wheels"}, "forget": {"1"}, "q": {"quiet room"}}, "/profile?forgot=1&q=quiet+room", ""},
	}
	for _, c := range cases {
		t.Run("when "+c.name, func(t *testing.T) {
			res, _ := newClient(t, fakeAsker{}).post(c.form)
			cookie := prefsCookieOf(res)
			t.Run("it should redirect back to the page with a GET", func(t *testing.T) {
				if res.StatusCode != http.StatusSeeOther {
					t.Fatalf("status %d", res.StatusCode)
				}
			})
			t.Run("it should send the attendee to "+c.location, func(t *testing.T) {
				if got := res.Header.Get("Location"); got != c.location {
					t.Fatalf("location %q", got)
				}
			})
			if c.cookie == "" {
				t.Run("it should clear the cookie", func(t *testing.T) {
					if cookie == nil || cookie.MaxAge >= 0 {
						t.Fatalf("cookie %+v", cookie)
					}
				})
			} else {
				t.Run("it should remember "+c.cookie+" in the cookie", func(t *testing.T) {
					if got, _ := url.QueryUnescape(cookie.Value); cookie == nil || got != c.cookie {
						t.Fatalf("cookie %+v", cookie)
					}
				})
			}
		})
	}

	t.Run("when the page is reopened after being forgotten", func(t *testing.T) {
		_, body := newClient(t, fakeAsker{}).get("/profile?forgot=1", "")
		t.Run("it should say so where a screen reader hears it", func(t *testing.T) {
			if !strings.Contains(body, `<p class="status" role="status">Forgotten. Nothing about you is on this device now.</p>`) {
				t.Fatalf("no status in %s", body)
			}
		})
		t.Run("it should tell the browser to forget too", func(t *testing.T) {
			if !strings.Contains(body, `data-forgot="1"`) {
				t.Fatalf("no forgot marker in %s", body)
			}
		})
	})
}

func TestProfileAPI(t *testing.T) {
	applied := ask.Answer{Summary: "Step-free entrance, 200 South: Salt Palace, Level 1. No stairs.", Applied: []model.Applied{{Chip: "step-free", Effect: "elevators and step-free entrances first"}}}

	t.Run("when the API is asked with a profile", func(t *testing.T) {
		asker := &profiledAsker{fakeAsker: fakeAsker{answer: applied}}
		_, body := newClient(t, asker).get("/api/ask?q=nearest+entrance&prefs=step-free,wheels", "")
		var got struct {
			Profile struct {
				Applied []model.Applied `json:"applied"`
			} `json:"profile"`
		}
		if err := json.Unmarshal([]byte(body), &got); err != nil {
			t.Fatal(err)
		}
		t.Run("it should hand the chips to the use case", func(t *testing.T) {
			if asker.prefs.String() != "step-free,wheels" {
				t.Fatalf("prefs %q", asker.prefs)
			}
		})
		t.Run("it should say which chips changed the answer and how", func(t *testing.T) {
			if len(got.Profile.Applied) != 1 || got.Profile.Applied[0] != applied.Applied[0] {
				t.Fatalf("applied %+v in %s", got.Profile.Applied, body)
			}
		})
	})

	t.Run("when the API is asked without a profile", func(t *testing.T) {
		asker := &profiledAsker{fakeAsker: fakeAsker{answer: ask.Answer{Summary: "x"}}}
		_, body := newClient(t, asker).get("/api/ask?q=nearest+entrance", "")
		t.Run("it should answer exactly as before", func(t *testing.T) {
			if strings.Contains(body, "profile") {
				t.Fatalf("profile in %s", body)
			}
		})
	})

	t.Run("when the use case knows nothing about profiles", func(t *testing.T) {
		_, body := newClient(t, fakeAsker{answer: ask.Answer{Summary: "x"}}).get("/api/ask?q=nearest+entrance&prefs=step-free", "")
		t.Run("it should still answer", func(t *testing.T) {
			if !strings.Contains(body, `"summary":"x"`) {
				t.Fatalf("body %s", body)
			}
		})
	})
}

func TestProfileTelemetry(t *testing.T) {
	reader := sdkmetric.NewManualReader(sdkmetric.WithTemporalitySelector(func(sdkmetric.InstrumentKind) metricdata.Temporality { return metricdata.DeltaTemporality }))
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { _ = mp.Shutdown(context.Background()) })
	c := newClient(t, fakeAsker{answer: ask.Answer{Summary: "x"}}, WithMeterProvider(mp))
	c.get("/ask?q=nearest+entrance&from=255+BC&prefs=step-free,wheels", "")
	c.get("/?prefs=step-free,wheels", "")

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatal(err)
	}
	points := map[string]int64{}
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != chipUsedMetric {
				continue
			}
			for _, dp := range m.Data.(metricdata.Sum[int64]).DataPoints {
				chip, _ := dp.Attributes.Value(attrChip)
				points[chip.AsString()] = dp.Value
			}
		}
	}

	t.Run("when a question is asked with two chips and the home page is opened with the same", func(t *testing.T) {
		t.Run("it should count each chip once, for the question only", func(t *testing.T) {
			if len(points) != 2 || points["step-free"] != 1 || points["wheels"] != 1 {
				t.Fatalf("points %v", points)
			}
		})
		t.Run("it should keep the chips out of the request log", func(t *testing.T) {
			if strings.Contains(c.log.String(), "step-free") {
				t.Fatalf("log %s", c.log.String())
			}
		})
		t.Run("it should still log the request", func(t *testing.T) {
			if !strings.Contains(c.log.String(), "route=/ask") {
				t.Fatalf("log %s", c.log.String())
			}
		})
	})
}

var _ = time.Now
