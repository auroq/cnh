package acceptance

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/auroq/cnh/testing/helpers"
)

const (
	declinedCookie  = "wd_prefs_no"
	learnedStepFree = removeStepFree + `<span class="from">, from your question</span>`
	learnedSentence = "Added Step-free only to what Wide Door knows about you, because you asked for an elevator. Remove it if that is wrong."
)

func valueOf(c *http.Cookie) string {
	if c == nil {
		return ""
	}
	got, _ := url.QueryUnescape(c.Value)
	return got
}

func cookieNamed(res *http.Response, name string) *http.Cookie {
	for _, c := range res.Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func TestLearnedChips(t *testing.T) {
	t.Parallel()

	t.Run("when I ask for the nearest elevator with no profile", func(t *testing.T) {
		t.Parallel()
		res := request(t, http.MethodGet, "/ask", helpers.Query("q", "nearest elevator"), "", "")
		body := bodyOf(t, res)
		t.Run("it should mark Step-free only as learned from my question", func(t *testing.T) {
			if !strings.Contains(body, learnedStepFree) {
				t.Fatal("learned chip missing from the page")
			}
		})
		t.Run("it should say so in the live region", func(t *testing.T) {
			if !strings.Contains(body, learnedSentence) {
				t.Fatal("no status sentence")
			}
		})
		t.Run("it should remember the chip as learned", func(t *testing.T) {
			if got := valueOf(cookieNamed(res, prefsCookie)); got != "step-free~" {
				t.Fatalf("cookie value %q", got)
			}
		})

		t.Run("and I remove it through the form without JavaScript", func(t *testing.T) {
			removed := request(t, http.MethodPost, "/profile", nil, "", url.Values{"prefs": {"step-free~"}, "remove": {"step-free"}, "q": {"nearest elevator"}}.Encode())
			t.Run("it should remember not to add it again", func(t *testing.T) {
				if got := valueOf(cookieNamed(removed, declinedCookie)); got != "step-free" {
					t.Fatalf("cookie value %q", got)
				}
			})
			t.Run("and I ask the same question again", func(t *testing.T) {
				again := bodyOf(t, requestWith(t, http.MethodGet, "/ask", helpers.Query("q", "nearest elevator"), map[string]string{declinedCookie: "step-free"}, ""))
				t.Run("it should not add the chip back", func(t *testing.T) {
					if strings.Contains(again, `<span class="vh">Remove `) {
						t.Fatal("a chip is on")
					}
				})
			})
		})
	})

	t.Run("when I ask for the nearest elevator with Step-free only already chosen", func(t *testing.T) {
		t.Parallel()
		body := bodyOf(t, request(t, http.MethodGet, "/ask", helpers.Query("q", "nearest elevator", "prefs", "step-free"), "", ""))
		t.Run("it should keep the chip as my choice", func(t *testing.T) {
			if strings.Contains(body, "from your question") {
				t.Fatal("the chosen chip was downgraded to learned")
			}
		})
	})
}
