package ask

import "github.com/auroq/cnh/pkg/model"

// Lesson is a chip a question implied; it lives in pkg/model with the rest of the profile vocabulary.
type Lesson = model.Lesson

type lesson struct {
	chip    string
	because string
	kinds   []string
	needs   []string
	words   []string
}

var wheelWords = []string{"wheelchair", "wheelchairs", "scooter", "scooters"}

var lessons = []lesson{
	{"step-free", "you asked for an elevator", []string{"elevator"}, nil, nil},
	{"step-free", "you asked for a step-free way in", []string{"step-free-entrance"}, []string{"mobility-accessible"}, nil},
	{"step-free", "you said no stairs", nil, nil, []string{"no stairs", "without stairs", "avoid stairs", "avoid the stairs", "not the stairs", "can t do stairs", "cannot do stairs", "can t use stairs", "can t climb stairs", "can t walk up the stairs", "no steps", "without steps"}},
	{"step-free", "you mentioned a wheelchair or scooter", []string{"mobility-equipment"}, nil, wheelWords},
	{"wheels", "you mentioned a wheelchair or scooter", []string{"mobility-equipment"}, nil, wheelWords},
	{"quiet-matters", "you asked for a quiet room", []string{"quiet-room"}, nil, nil},
	{"quiet-matters", "you said you need a rest", nil, nil, []string{"overwhelmed", "sensory", "need a break", "take a break", "need a rest"}},
	{"avoid-crowds", "you mentioned crowds or noise", nil, nil, []string{"overstimulated", "crowd", "crowds", "crowded", "too loud", "so loud", "noise", "noisy"}},
	{"all-gender", "you asked for an all-gender restroom", nil, nil, []string{"all gender", "all genders", "gender neutral", "gender inclusive"}},
	{"captions-asl", "you asked about captions or sign language", nil, nil, []string{"sign language", "asl", "caption", "captions", "captioned", "captioning", "interpreter", "interpreters", "deaf", "hard of hearing"}},
	{"dietary", "you mentioned dietary needs", nil, nil, []string{"dietary", "diet", "allergy", "allergies", "allergen", "allergens", "halal", "kosher", "vegan", "vegetarian", "gluten", "lactose", "dairy free", "nut free"}},
}

// Learned maps what the question implied, through KindOf, the need it was understood as and a few plain words, to the chips it adds to the profile.
// Each chip is learned once, with the first reason that fits.
func Learned(question string, need *Need) []Lesson {
	kind := KindOf(question)
	needID := ""
	if need != nil {
		needID = need.ID
	}
	var out []Lesson
	for _, l := range lessons {
		if hasLesson(out, l.chip) || !l.applies(question, kind, needID) {
			continue
		}
		out = append(out, Lesson{Chip: l.chip, Because: l.because})
	}
	return out
}

func (l lesson) applies(question, kind, needID string) bool {
	return contains(l.kinds, kind) || contains(l.needs, needID) || anyWord(question, l.words...)
}

func contains(list []string, s string) bool {
	for _, item := range list {
		if s != "" && item == s {
			return true
		}
	}
	return false
}

func hasLesson(list []Lesson, chip string) bool {
	for _, l := range list {
		if l.Chip == chip {
			return true
		}
	}
	return false
}
