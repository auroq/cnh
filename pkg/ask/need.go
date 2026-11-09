package ask

import (
	"context"
	"strings"
	"unicode"
)

// NeedMatch is a need the store thinks the question is about, with how sure it is.
type NeedMatch struct {
	Need  Need
	Score float64
}

// MinNeedScore is the semantic score below which a need match is a guess and the question falls through to the hybrid place search.
const MinNeedScore = 7.0

// NeedOf finds the first need one of whose phrases appears in the question as whole words, in the order the needs are listed.
func NeedOf(question string, needs []Need) (Need, bool) {
	q := wordsOf(question)
	for _, n := range needs {
		for _, p := range n.Phrases {
			if p = wordsOf(p); p != "  " && strings.Contains(q, p) {
				return n, true
			}
		}
	}
	return Need{}, false
}

func wordsOf(s string) string {
	return " " + strings.Join(strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsNumber(r) }), " ") + " "
}

// placesFor runs the inference in order: a kind word in the question, a need phrase, the store's semantic guess, then the plain hybrid search.
// A need is offered kind by kind until one has a place.
func (s *Service) placesFor(ctx context.Context, ans *Answer, question string) ([]Place, error) {
	if kind := KindOf(question); kind != "" {
		return s.findPlaces(ctx, question, kind)
	}
	need, ok := NeedOf(question, s.Needs)
	if !ok {
		matches, err := step(ctx, s.tracer(), "ask.FindNeeds", func(ctx context.Context) ([]NeedMatch, error) { return s.Store.FindNeeds(ctx, question) })
		if err != nil {
			return nil, err
		}
		need, ok = bestNeed(matches)
	}
	if ok {
		ans.Need = &need
		for _, kind := range need.Kinds {
			places, err := s.findPlaces(ctx, question, kind)
			if err != nil || len(places) > 0 {
				return places, err
			}
		}
	}
	return s.findPlaces(ctx, question, "")
}

func (s *Service) findPlaces(ctx context.Context, question, kind string) ([]Place, error) {
	return step(ctx, s.tracer(), "ask.FindPlaces", func(ctx context.Context) ([]Place, error) { return s.Store.FindPlaces(ctx, question, kind) })
}

// bestNeed keeps the top semantic match only when the store is sure enough.
func bestNeed(matches []NeedMatch) (Need, bool) {
	if len(matches) == 0 || matches[0].Score < MinNeedScore {
		return Need{}, false
	}
	return matches[0].Need, true
}
