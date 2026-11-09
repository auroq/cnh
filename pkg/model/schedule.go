package model

// Match is one pasted title and the session it named, nil when nothing on the schedule was close enough.
type Match struct {
	Title   string   `json:"title"`
	Session *Session `json:"session,omitempty"`
}

// Schedule is what the attendee's own sessions say about now: the next one, the one just left,
// the room assumed from it when no location was given, and whether three run back to back.
type Schedule struct {
	Next            *Session `json:"next,omitempty"`
	JustLeft        *Session `json:"just_left,omitempty"`
	AssumedLocation string   `json:"assumed_location,omitempty"`
	BackToBack      bool     `json:"back_to_back,omitempty"`
}
