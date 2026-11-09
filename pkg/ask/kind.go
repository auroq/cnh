package ask

import (
	"strings"
	"unicode"
)

var kindWords = []struct {
	kind  string
	words []string
}{
	{"elevator", []string{"elevator", "elevators", "lift", "lifts"}},
	{"all-gender-restroom", []string{"restroom", "restrooms", "toilet", "toilets", "bathroom", "bathrooms", "washroom", "washrooms"}},
	{"quiet-room", []string{"quiet room", "quiet rooms", "quiet space", "sensory"}},
	{"prayer-room", []string{"pray", "prayer"}},
	{"nursing-room", []string{"nursing", "lactation"}},
	{"step-free-entrance", []string{"step free", "stepfree", "accessible entrance"}},
	{"mobility-equipment", []string{"wheelchair", "wheelchairs", "scooter", "scooters"}},
	{"parking", []string{"parking"}},
	{"community-hub", []string{"community hub"}},
	{"coat-check", []string{"coat", "coats", "coat check", "bag check"}},
	{"info-desk", []string{"help desk", "info desk", "information desk", "registration"}},
}

// KindOf names the kind of place a question asks for, or "" when it names none.
// A kind named in the question outranks nearness: the store is asked for that kind first.
func KindOf(question string) string {
	words := " " + strings.Join(strings.FieldsFunc(strings.ToLower(question), isNotLetter), " ") + " "
	for _, k := range kindWords {
		for _, w := range k.words {
			if strings.Contains(words, " "+w+" ") {
				return k.kind
			}
		}
	}
	return ""
}

func isNotLetter(r rune) bool { return !unicode.IsLetter(r) }
