package ask

import (
	"context"
	"strings"
	"testing"
	"time"
)

var noon = time.Date(2026, 11, 11, 19, 0, 0, 0, time.UTC)

type countingStore struct {
	fakeStore
	sessionsAsked int
}

func (f *countingStore) FindSessions(ctx context.Context, question string) ([]Session, error) {
	f.sessionsAsked++
	return f.fakeStore.FindSessions(ctx, question)
}

func TestAskSessions(t *testing.T) {
	ctx := context.Background()
	clock := func() time.Time { return noon }
	ids := func(sessions []Session) string {
		var out []string
		for _, s := range sessions {
			out = append(out, s.ID)
		}
		return strings.Join(out, ",")
	}
	soon := Session{ID: "soon", Title: "Sign Language Crash Course", Start: noon.Add(25 * time.Minute), End: noon.Add(85 * time.Minute), Venue: "Salt Palace", Level: "Level 1"}
	later := Session{ID: "later", Title: "Sign Language Crash Course", Start: noon.Add(48 * time.Hour), End: noon.Add(49 * time.Hour), Venue: "Salt Palace", Level: "Level 1"}
	over := Session{ID: "over", Title: "Captions Panel", Start: noon.Add(-2 * time.Hour), End: noon.Add(-time.Hour), Venue: "Salt Palace", Level: "Level 1"}
	yesterday := Session{ID: "yesterday", Title: "Captions Panel", Start: noon.Add(-26 * time.Hour), End: noon.Add(-25 * time.Hour)}
	networking := Session{ID: "net", Title: "Kubernetes Networking Deep Dive", Start: noon.Add(time.Hour), End: noon.Add(2 * time.Hour)}

	t.Run("when the question is about a session feature", func(t *testing.T) {
		store := &countingStore{fakeStore: fakeStore{room: Here{Room: "255 BC", Venue: "Salt Palace", Level: "Level 2"}, sessions: []Session{over, soon, later}}}
		ans, err := (&Service{Store: store, Now: clock}).Ask(ctx, "sign language", "255 BC")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should say why sessions are shown", func(t *testing.T) {
			if ans.SessionsShownBecause != ShownForSessionQuestion {
				t.Fatalf("reason = %q", ans.SessionsShownBecause)
			}
		})
		t.Run("it should keep only what is on in the next two hours", func(t *testing.T) {
			if ids(ans.Sessions) != "soon" {
				t.Fatalf("sessions = %s", ids(ans.Sessions))
			}
		})
		t.Run("it should say the course is upcoming", func(t *testing.T) {
			if len(ans.Sessions) != 1 || ans.Sessions[0].State != SessionUpcoming {
				t.Fatalf("sessions = %+v", ans.Sessions)
			}
		})
		t.Run("it should say the course is in the same building", func(t *testing.T) {
			if len(ans.Sessions) != 1 || ans.Sessions[0].Distance != DistanceSameBuilding {
				t.Fatalf("sessions = %+v", ans.Sessions)
			}
		})
	})

	t.Run("when nothing relevant is on in the next two hours", func(t *testing.T) {
		store := &countingStore{fakeStore: fakeStore{sessions: []Session{over, later}}}
		ans, err := (&Service{Store: store, Now: clock}).Ask(ctx, "sign language", "")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should show the next relevant sessions instead", func(t *testing.T) {
			if ids(ans.Sessions) != "later" {
				t.Fatalf("sessions = %s", ids(ans.Sessions))
			}
		})
	})

	t.Run("when the question is about a session that is over", func(t *testing.T) {
		store := &countingStore{fakeStore: fakeStore{sessions: []Session{yesterday, over, later}}}
		ans, err := (&Service{Store: store, Now: clock}).Ask(ctx, "the captions panel I just got out of", "")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should include today's ended sessions but not yesterday's or far-off ones", func(t *testing.T) {
			if ids(ans.Sessions) != "over" {
				t.Fatalf("sessions = %s", ids(ans.Sessions))
			}
		})
	})

	t.Run("when the question names a kind of place", func(t *testing.T) {
		store := &countingStore{fakeStore: fakeStore{sessions: []Session{soon}}}
		ans, err := (&Service{Store: store, Now: clock}).Ask(ctx, "nearest elevator", "")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should hide sessions", func(t *testing.T) {
			if len(ans.Sessions) != 0 || ans.SessionsShownBecause != SessionsHidden {
				t.Fatalf("reason = %q, sessions = %+v", ans.SessionsShownBecause, ans.Sessions)
			}
		})
		t.Run("it should not ask the store for sessions", func(t *testing.T) {
			if store.sessionsAsked != 0 {
				t.Fatalf("asked %d times", store.sessionsAsked)
			}
		})
	})

	t.Run("when the question matches a session title", func(t *testing.T) {
		store := &countingStore{fakeStore: fakeStore{sessions: []Session{networking}}}
		ans, err := (&Service{Store: store, Now: clock}).Ask(ctx, "kubernetes networking", "")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should show sessions because of the title", func(t *testing.T) {
			if ans.SessionsShownBecause != ShownForTitleMatch || ids(ans.Sessions) != "net" {
				t.Fatalf("reason = %q, sessions = %+v", ans.SessionsShownBecause, ans.Sessions)
			}
		})
	})

	t.Run("when the question matches no session title", func(t *testing.T) {
		store := &countingStore{fakeStore: fakeStore{sessions: []Session{networking}}}
		ans, err := (&Service{Store: store, Now: clock}).Ask(ctx, "where is coffee", "")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should hide the sessions it found", func(t *testing.T) {
			if ans.SessionsShownBecause != SessionsHidden || len(ans.Sessions) != 0 {
				t.Fatalf("reason = %q, sessions = %+v", ans.SessionsShownBecause, ans.Sessions)
			}
		})
	})
}

func TestSessionsShown(t *testing.T) {
	rest := Need{ID: "rest", Kinds: []string{"quiet-room"}}
	private := Need{ID: "private-space", Kinds: []string{"all-gender-restroom"}}
	cases := []struct {
		name     string
		question string
		kind     string
		need     *Need
		want     string
	}{
		{"a place kind is named", "nearest elevator", "elevator", nil, SessionsHidden},
		{"a place need is understood", "I need to fix my makeup", "", &private, SessionsHidden},
		{"a rest need is understood", "I'm overwhelmed", "", &rest, SessionsHidden},
		{"the words say rest even with a session word", "quiet talk-free place for a sensory break", "", nil, SessionsHidden},
		{"a session word is used", "talks about accessibility", "", nil, ShownForSessionQuestion},
		{"a session feature is named", "sign language", "", nil, ShownForSessionQuestion},
		{"captions are asked for", "which sessions have captions", "", nil, ShownForSessionQuestion},
		{"nothing says sessions or places", "kubernetes networking", "", nil, ShownIfTitleMatches},
	}
	for _, c := range cases {
		t.Run("when "+c.name, func(t *testing.T) {
			t.Run("it should decide "+c.want, func(t *testing.T) {
				if got := SessionsShown(c.question, c.kind, c.need); got != c.want {
					t.Fatalf("got %q", got)
				}
			})
		})
	}
}

func TestTitleMatches(t *testing.T) {
	cases := []struct {
		name     string
		question string
		title    string
		want     bool
	}{
		{"two long words of the question are in the title", "kubernetes networking", "Kubernetes Networking Deep Dive", true},
		{"half the long words are in the title", "scaling kubernetes clusters today", "Scaling Kubernetes to 10k Nodes", true},
		{"only one long word is in the title", "where is coffee", "Coffee With the Maintainers", false},
		{"no long word is in the title", "nearest exit", "Rego Without Borders", false},
		{"the question has no long words", "a b c", "Anything", false},
	}
	for _, c := range cases {
		t.Run("when "+c.name, func(t *testing.T) {
			t.Run("it should report a match only when they share enough words", func(t *testing.T) {
				if got := TitleMatches(c.question, c.title); got != c.want {
					t.Fatalf("got %v", got)
				}
			})
		})
	}
}

func TestWindow(t *testing.T) {
	t.Run("when the question is about now", func(t *testing.T) {
		from, to := Window(noon, "sign language")
		t.Run("it should start five minutes ago", func(t *testing.T) {
			if !from.Equal(noon.Add(-5 * time.Minute)) {
				t.Fatalf("from = %s", from)
			}
		})
		t.Run("it should end two hours from now", func(t *testing.T) {
			if !to.Equal(noon.Add(2 * time.Hour)) {
				t.Fatalf("to = %s", to)
			}
		})
	})
	for _, q := range []string{"the talk I just got out of", "what was on earlier", "did I miss the keynote"} {
		t.Run("when the question says "+q, func(t *testing.T) {
			from, _ := Window(noon, q)
			t.Run("it should start at the beginning of today in Mountain Time", func(t *testing.T) {
				if !from.Equal(time.Date(2026, 11, 11, 7, 0, 0, 0, time.UTC)) {
					t.Fatalf("from = %s", from)
				}
			})
		})
	}
}

func TestDescribe(t *testing.T) {
	here := Here{Venue: "Salt Palace", Level: "Level 2"}
	cases := []struct {
		name    string
		session Session
		here    Here
		state   string
		startIn int
		dist    string
	}{
		{"it starts in 25 minutes on my level", Session{Start: noon.Add(25 * time.Minute), End: noon.Add(55 * time.Minute), Venue: "Salt Palace", Level: "Level 2"}, here, SessionUpcoming, 25, DistanceSameLevel},
		{"it is running in my building", Session{Start: noon.Add(-10 * time.Minute), End: noon.Add(20 * time.Minute), Venue: "Salt Palace", Level: "Level 1"}, here, SessionRunning, 0, DistanceSameBuilding},
		{"it ended in the other building", Session{Start: noon.Add(-60 * time.Minute), End: noon.Add(-30 * time.Minute), Venue: "Hyatt Regency", Level: "Level 2"}, here, SessionEnded, 0, DistanceOtherBuilding},
		{"I did not say where I am", Session{Start: noon.Add(90 * time.Second), End: noon.Add(time.Hour), Venue: "Salt Palace", Level: "Level 2"}, Here{}, SessionUpcoming, 2, DistanceUnknown},
	}
	for _, c := range cases {
		t.Run("when "+c.name, func(t *testing.T) {
			got := Describe([]Session{c.session}, noon, c.here)[0]
			t.Run("it should say its state", func(t *testing.T) {
				if got.State != c.state {
					t.Fatalf("state = %q", got.State)
				}
			})
			t.Run("it should say how many minutes until it starts", func(t *testing.T) {
				if got.StartsInMinutes != c.startIn {
					t.Fatalf("starts in = %d", got.StartsInMinutes)
				}
			})
			t.Run("it should say how far it is in the same tiers as places", func(t *testing.T) {
				if got.Distance != c.dist {
					t.Fatalf("distance = %q", got.Distance)
				}
			})
		})
	}
}
