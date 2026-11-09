package ask

import (
	"context"
	"strings"

	"github.com/auroq/cnh/pkg/model"
)

// Preferences is the attendee's profile, the chips they chose; it lives in pkg/model like the rest of the vocabulary.
type Preferences = model.Preferences

// AskFor is Ask with the attendee's profile: the chips decide which kinds are offered first and what the summary adds.
// Each effect is one row below, and the answer names every chip that changed it.
func (s *Service) AskFor(ctx context.Context, question, from string, prefs Preferences) (Answer, error) {
	return s.ask(ctx, question, from, prefs)
}

type kindEffect struct {
	chip  string
	when  func(question, kind string) bool
	kinds []string
	says  string
}

type summaryEffect struct {
	chip string
	when func(a Answer) bool
	note string
	says string
}

var stepFreeKinds = []string{"step-free-entrance", "elevator"}

var kindEffects = []kindEffect{
	{"step-free", entranceQuestion, stepFreeKinds, "elevators and step-free entrances first"},
	{"wheels", entranceQuestion, stepFreeKinds, "elevators and step-free entrances first"},
	{"quiet-matters", restQuestion, []string{"quiet-room"}, "quiet rooms first for a rest"},
	{"all-gender", func(_, kind string) bool { return kind == "all-gender-restroom" }, []string{"all-gender-restroom"}, "all-gender restrooms first"},
}

var summaryEffects = []summaryEffect{
	{"step-free", topKindIn(stepFreeKinds...), "Step-free route, no stairs.", "no stairs on the way"},
	{"wheels", topKindIn(stepFreeKinds...), "Step-free route, no stairs.", "no stairs on the way"},
	{"avoid-crowds", topPlaceIn("Solutions Showcase", "Grand Ballroom"), "Heads up: this is in a busy area, the Solutions Showcase and Grand Ballroom draw crowds and noise.", "a warning when the place is in a crowded area"},
	{"captions-asl", func(a Answer) bool { return looksLikeSessionQuestion(a.Question) && len(a.Sessions) > 0 }, "Sessions are listed below; Wordly AI captions run in every session room, ask the room host to switch them on.", "sessions kept for talk questions, with where captions come from"},
	{"dietary", func(a Answer) bool { return foodQuestion(a.Question) }, "For dietary needs, ask at the information desk at registration, 200 South entrance; they know what each caterer labels.", "the information desk for food questions"},
}

// preferredKinds is the hook after KindOf: the kinds the profile puts first for this question, the inferred kind last.
func preferredKinds(ans *Answer, question, kind string, prefs Preferences) []string {
	var kinds []string
	for _, e := range kindEffects {
		if prefs.Has(e.chip) && e.when(question, kind) {
			kinds = appendMissing(kinds, e.kinds...)
			addApplied(ans, e.chip, e.says)
		}
	}
	if len(kinds) > 0 && kind != "" {
		kinds = appendMissing(kinds, kind)
	}
	return kinds
}

// shapeSummary is the hook after Summarise: the sentences the profile adds, each a row above.
func shapeSummary(ans *Answer, prefs Preferences) {
	for _, e := range summaryEffects {
		if prefs.Has(e.chip) && e.when(*ans) && !strings.Contains(ans.Summary, e.note) {
			ans.Summary += " " + e.note
			addApplied(ans, e.chip, e.says)
		}
	}
}

func addApplied(ans *Answer, chip, says string) {
	for i, a := range ans.Applied {
		if a.Chip == chip {
			if !strings.Contains(a.Effect, says) {
				ans.Applied[i].Effect += "; " + says
			}
			return
		}
	}
	ans.Applied = append(ans.Applied, model.Applied{Chip: chip, Effect: says})
}

func appendMissing(kinds []string, more ...string) []string {
	for _, k := range more {
		found := false
		for _, have := range kinds {
			found = found || have == k
		}
		if !found {
			kinds = append(kinds, k)
		}
	}
	return kinds
}

func topKindIn(kinds ...string) func(Answer) bool {
	return func(a Answer) bool {
		if len(a.Places) == 0 {
			return false
		}
		for _, k := range kinds {
			if a.Places[0].Kind == k {
				return true
			}
		}
		return false
	}
}

func topPlaceIn(areas ...string) func(Answer) bool {
	return func(a Answer) bool {
		if len(a.Places) == 0 {
			return false
		}
		text := strings.ToLower(a.Places[0].Name + " " + a.Places[0].Near)
		for _, area := range areas {
			if strings.Contains(text, strings.ToLower(area)) {
				return true
			}
		}
		return false
	}
}

func anyWord(question string, phrases ...string) bool {
	words := wordsOf(question)
	for _, p := range phrases {
		if strings.Contains(words, " "+p+" ") {
			return true
		}
	}
	return false
}

func entranceQuestion(question, _ string) bool {
	return anyWord(question, "entrance", "entrances", "entry", "enter", "way in", "get in", "door", "doors")
}

func restQuestion(question, _ string) bool {
	return anyWord(question, "rest", "sit", "sit down", "break", "lie down", "calm", "decompress", "overwhelmed", "recharge")
}

func foodQuestion(question string) bool {
	return anyWord(question, "food", "eat", "lunch", "breakfast", "dinner", "meal", "meals", "snack", "snacks", "coffee", "vegan", "vegetarian", "gluten", "halal", "kosher", "allergy", "allergies", "dietary")
}
