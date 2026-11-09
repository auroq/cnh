package model

// Lesson is one chip a question implied and the reason, in words the attendee reads next to the chip.
type Lesson struct {
	Chip    string `json:"chip"`
	Because string `json:"because"`
}
