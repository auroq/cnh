package web

import (
	v1 "github.com/auroq/cnh/pkg/api/v1"
	"github.com/auroq/cnh/pkg/model"
)

// toV1 maps the domain answer onto the frozen wire contract.
// The two look alike today on purpose; when the model moves, this is where the contract stays still.
func toV1(a model.Answer) v1.AskResponse {
	out := v1.AskResponse{
		Question: a.Question,
		Here:     v1.Here{Asked: a.Here.Asked, Room: a.Here.Room, Venue: a.Here.Venue, Level: a.Here.Level, Side: a.Here.Side, Via: a.Here.Via},
		Summary:  a.Summary,

		SessionsShownBecause: a.SessionsShownBecause,
	}
	if a.Places != nil {
		out.Places = make([]v1.Place, 0, len(a.Places))
		for _, p := range a.Places {
			place := v1.Place{
				ID: p.ID, Kind: p.Kind, Name: p.Name, Venue: p.Venue, Level: p.Level, Near: p.Near,
				Notes: p.Notes, For: p.For, Route: p.Route, Verified: p.Verified, Source: p.Source, Side: p.Side, LeadsTo: p.LeadsTo,
			}
			if st := p.Attendee; st != nil {
				at := st.LatestAt
				place.AttendeeStatus, place.ReportCount, place.LatestNote, place.LatestAt = st.Status, st.Count, st.LatestNote, &at
			}
			out.Places = append(out.Places, place)
		}
	}
	if a.Sessions != nil {
		out.Sessions = make([]v1.Session, 0, len(a.Sessions))
		for _, s := range a.Sessions {
			out.Sessions = append(out.Sessions, sessionV1(s))
		}
	}
	if a.Schedule != nil {
		out.Schedule = &v1.Schedule{AssumedLocation: a.Schedule.AssumedLocation, BackToBack: a.Schedule.BackToBack}
		if a.Schedule.Next != nil {
			next := sessionV1(*a.Schedule.Next)
			out.Schedule.Next = &next
		}
		if a.Schedule.JustLeft != nil {
			left := sessionV1(*a.Schedule.JustLeft)
			out.Schedule.JustLeft = &left
		}
	}
	if a.Applied != nil {
		out.Profile = &v1.Profile{Applied: make([]v1.Applied, 0, len(a.Applied))}
		for _, ap := range a.Applied {
			out.Profile.Applied = append(out.Profile.Applied, v1.Applied{Chip: ap.Chip, Effect: ap.Effect, Strength: string(ap.Strength)})
		}
	}
	return out
}

func sessionV1(s model.Session) v1.Session {
	return v1.Session{
		ID: s.ID, Title: s.Title, URL: s.URL, Speakers: s.Speakers, Start: s.Start, End: s.End,
		Location: s.Location, Venue: s.Venue, Level: s.Level, Room: s.Room,
		State: s.State, StartsInMinutes: s.StartsInMinutes, Distance: s.Distance,
	}
}
