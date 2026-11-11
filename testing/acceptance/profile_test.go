package acceptance

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/auroq/cnh/testing/helpers"
)

const (
	prefsCookie    = "wd_prefs"
	removeStepFree = `<span class="vh">Remove </span><span aria-hidden="true">&#10003; </span><span>Step-free only`
)

func request(t *testing.T, method, path string, query url.Values, cookie, form string) *http.Response {
	t.Helper()
	cookies := map[string]string{}
	if cookie != "" {
		cookies[prefsCookie] = cookie
	}
	return requestWith(t, method, path, query, cookies, form)
}

func requestWith(t *testing.T, method, path string, query url.Values, cookies map[string]string, form string) *http.Response {
	t.Helper()
	u := helpers.BaseURL(t) + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)
	req, err := http.NewRequestWithContext(ctx, method, u, strings.NewReader(form))
	if err != nil {
		t.Fatalf("%s %s: %v", method, u, err)
	}
	if form != "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}
	client := &http.Client{Timeout: 20 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, u, err)
	}
	t.Cleanup(func() { res.Body.Close() })
	return res
}

func bodyOf(t *testing.T, res *http.Response) string {
	t.Helper()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func cookieOf(res *http.Response) *http.Cookie {
	for _, c := range res.Cookies() {
		if c.Name == prefsCookie {
			return c
		}
	}
	return nil
}

func TestProfile(t *testing.T) {
	t.Parallel()

	t.Run("when I add Step-free only through the form without JavaScript", func(t *testing.T) {
		t.Parallel()
		res := request(t, http.MethodPost, "/profile", nil, "", url.Values{"add": {"step-free"}}.Encode())
		cookie := cookieOf(res)
		if cookie == nil {
			t.Fatalf("no %s cookie: %v", prefsCookie, res.Header)
		}
		t.Run("it should send me back to the profile page with the chip", func(t *testing.T) {
			if res.StatusCode != http.StatusSeeOther || res.Header.Get("Location") != "/profile?added=step-free&prefs=step-free" {
				t.Fatalf("%d %q", res.StatusCode, res.Header.Get("Location"))
			}
		})
		t.Run("it should put nothing but the chip ids in the cookie", func(t *testing.T) {
			if got, _ := url.QueryUnescape(cookie.Value); got != "step-free" {
				t.Fatalf("cookie value %q", cookie.Value)
			}
		})
		t.Run("it should keep the cookie first-party and secure", func(t *testing.T) {
			if !cookie.Secure || cookie.SameSite != http.SameSiteStrictMode || cookie.Path != "/" {
				t.Fatalf("cookie %+v", cookie)
			}
		})
		t.Run("and I open a fresh page with only the cookie", func(t *testing.T) {
			body := bodyOf(t, request(t, http.MethodGet, "/", nil, cookie.Value, ""))
			t.Run("it should render the chip", func(t *testing.T) {
				if !strings.Contains(body, removeStepFree) {
					t.Fatal("chip missing from the page")
				}
			})
		})
	})

	t.Run("when I ask to be forgotten", func(t *testing.T) {
		t.Parallel()
		res := request(t, http.MethodPost, "/profile", nil, url.QueryEscape("step-free"), url.Values{"prefs": {"step-free"}, "forget": {"1"}}.Encode())
		t.Run("it should expire the cookie", func(t *testing.T) {
			if c := cookieOf(res); c == nil || c.MaxAge >= 0 {
				t.Fatalf("cookie %+v", c)
			}
		})
		t.Run("and I follow the redirect", func(t *testing.T) {
			body := bodyOf(t, request(t, http.MethodGet, res.Header.Get("Location"), nil, "", ""))
			t.Run("it should show no chip as on", func(t *testing.T) {
				if strings.Contains(body, `<span class="vh">Remove `) {
					t.Fatal("a chip is still on")
				}
			})
			t.Run("it should say I am forgotten", func(t *testing.T) {
				if !strings.Contains(body, "Forgotten. Nothing about you is on this device now.") {
					t.Fatal("no status sentence")
				}
			})
		})
	})

	t.Run("when I open the saved link on a client with no cookie", func(t *testing.T) {
		t.Parallel()
		body := bodyOf(t, request(t, http.MethodGet, "/", helpers.Query("prefs", "step-free,wheels"), "", ""))
		t.Run("it should reproduce the chips", func(t *testing.T) {
			if !strings.Contains(body, removeStepFree) || !strings.Contains(body, "Remove </span><span aria-hidden=\"true\">&#10003; </span><span>Wide doors and ramps") {
				t.Fatal("chips missing from the page")
			}
		})
	})

	effects := []struct {
		prefs, question, from, firstPlace, summaryHas string
	}{
		{"step-free", "nearest entrance", "Grand Ballroom J", "step-free-200-south", "no stairs"},
		{"wheels", "nearest entrance", "", "step-free-200-south", "no stairs"},
		{"avoid-crowds", "nearest all-gender restroom", "Grand Ballroom J", "agr-sp-l1-showcase", "crowds and noise"},
	}
	for _, e := range effects {
		t.Run("when my profile says "+e.prefs+" and I ask for "+e.question+" "+whereIAm(e.from), func(t *testing.T) {
			t.Parallel()
			var a answer
			helpers.GetJSON(t, "/api/ask", helpers.Query("q", e.question, "from", e.from, "prefs", e.prefs), &a)
			t.Run("it should answer "+e.firstPlace+" first", func(t *testing.T) {
				if len(a.Places) == 0 || a.Places[0].ID != e.firstPlace {
					t.Fatalf("places = %+v", a.Places)
				}
			})
			t.Run("it should say "+e.summaryHas, func(t *testing.T) {
				if !strings.Contains(strings.ToLower(a.Summary), e.summaryHas) {
					t.Fatalf("summary = %q", a.Summary)
				}
			})
			t.Run("it should name the chip that changed the answer", func(t *testing.T) {
				if len(a.Profile.Applied) != 1 || a.Profile.Applied[0].Chip != e.prefs {
					t.Fatalf("profile = %+v", a.Profile)
				}
			})
		})
	}

	t.Run("when I ask with no profile", func(t *testing.T) {
		t.Parallel()
		status, body := helpers.Get(t, "/api/ask", helpers.Query("q", "nearest quiet room"), "application/json")
		t.Run("it should answer as before, without a profile field", func(t *testing.T) {
			if status != http.StatusOK || strings.Contains(body, `"profile"`) {
				t.Fatalf("%d %s", status, body)
			}
		})
	})
}
