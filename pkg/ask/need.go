package ask

import (
	"strings"
	"unicode"
)

// NeedMatch is a need the store thinks the question is about, with how sure it is.
type NeedMatch struct {
	Need  Need
	Score float64
}

// MinNeedScore is the semantic score below which a need match is a guess and the question falls through to the hybrid place search.
const MinNeedScore = 8.0

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

// bestNeed keeps the top semantic match only when the store is sure enough.
func bestNeed(matches []NeedMatch) (Need, bool) {
	if len(matches) == 0 || matches[0].Score < MinNeedScore {
		return Need{}, false
	}
	return matches[0].Need, true
}
