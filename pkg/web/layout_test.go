package web

import (
	"strings"
	"testing"

	"github.com/auroq/cnh/pkg/ask"
)

const (
	barOpen    = `<p class="profile-bar">`
	asideOpen  = `<aside id="profile" class="profile`
	mainOpen   = `<main id="main">`
	footerOpen = `<footer>`
)

func indexOrFail(t *testing.T, body, marker string) int {
	t.Helper()
	i := strings.Index(body, marker)
	if i < 0 {
		t.Fatalf("%q missing in %s", marker, body)
	}
	return i
}

func TestProfileLayout(t *testing.T) {
	answered := fakeAsker{answer: ask.Answer{Summary: "Quiet room 260A: Salt Palace, Level 2."}}

	t.Run("when a page that is not the profile page is rendered with two chips", func(t *testing.T) {
		_, body := newClient(t, answered).get("/ask?q=quiet+room&from=255+BC&prefs=step-free,wheels", "")
		shows := map[string]string{
			"sum the profile up in one line under the header": `<p class="profile-bar">What Wide Door knows about you: 2 things. `,
			"point the wide Edit link at the aside":           `<a class="edit wide" href="#profile">Edit`,
			"point the narrow Edit link at the profile page":  `<a class="edit narrow" href="/profile?from=255&#43;BC&amp;prefs=step-free%2Cwheels&amp;q=quiet&#43;room">Edit`,
			"label the aside by its heading":                  `<aside id="profile" class="profile" aria-labelledby="profile-heading">`,
			"head the aside with the count":                   `<h2 id="profile-heading">What Wide Door knows about you: 2 things</h2>`,
		}
		for name, want := range shows {
			t.Run("it should "+name, func(t *testing.T) {
				if !strings.Contains(body, want) {
					t.Fatalf("%q missing in %s", want, body)
				}
			})
		}
		t.Run("it should render no collapsed profile panel", func(t *testing.T) {
			if strings.Contains(body, `<details class="profile"`) {
				t.Fatalf("details.profile in %s", body)
			}
		})
		t.Run("it should put the bar before main", func(t *testing.T) {
			if indexOrFail(t, body, barOpen) > indexOrFail(t, body, mainOpen) {
				t.Fatalf("bar after main in %s", body)
			}
		})
		t.Run("it should put the aside after main and before the footer", func(t *testing.T) {
			if a, m, f := indexOrFail(t, body, asideOpen), indexOrFail(t, body, mainOpen), indexOrFail(t, body, footerOpen); a < m || a > f {
				t.Fatalf("aside at %d, main at %d, footer at %d in %s", a, m, f, body)
			}
		})
		t.Run("it should keep the chips out of main", func(t *testing.T) {
			if indexOrFail(t, body, `<ul class="chips">`) < indexOrFail(t, body, `</main>`) {
				t.Fatalf("chips inside main in %s", body)
			}
		})
	})

	t.Run("when the profile page is rendered", func(t *testing.T) {
		_, body := newClient(t, answered).get("/profile?prefs=step-free", "")
		t.Run("it should render no summary bar", func(t *testing.T) {
			if strings.Contains(body, barOpen) {
				t.Fatalf("bar in %s", body)
			}
		})
		t.Run("it should put the aside before main", func(t *testing.T) {
			if indexOrFail(t, body, asideOpen) > indexOrFail(t, body, mainOpen) {
				t.Fatalf("aside after main in %s", body)
			}
		})
		t.Run("it should mark the aside as shown at every width", func(t *testing.T) {
			if !strings.Contains(body, `<aside id="profile" class="profile shown"`) {
				t.Fatalf("aside not shown in %s", body)
			}
		})
	})

	t.Run("when the home page is rendered with no profile", func(t *testing.T) {
		_, body := newClient(t, answered).get("/", "")
		t.Run("it should say it knows nothing yet in the bar", func(t *testing.T) {
			if !strings.Contains(body, `<p class="profile-bar">What Wide Door knows about you: nothing yet. `) {
				t.Fatalf("bar missing in %s", body)
			}
		})
		t.Run("it should send the narrow Edit link to the bare profile page", func(t *testing.T) {
			if !strings.Contains(body, `<a class="edit narrow" href="/profile">Edit`) {
				t.Fatalf("edit link missing in %s", body)
			}
		})
	})
}
