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
		Here:     v1.Here{Asked: a.Here.Asked, Room: a.Here.Room, Venue: a.Here.Venue, Level: a.Here.Level},
		Summary:  a.Summary,
	}
	if a.Places != nil {
		out.Places = make([]v1.Place, 0, len(a.Places))
		for _, p := range a.Places {
			out.Places = append(out.Places, v1.Place{
				ID: p.ID, Kind: p.Kind, Name: p.Name, Venue: p.Venue, Level: p.Level, Near: p.Near,
				Notes: p.Notes, For: p.For, Route: p.Route, Verified: p.Verified, Source: p.Source,
			})
		}
	}
	if a.Sessions != nil {
		out.Sessions = make([]v1.Session, 0, len(a.Sessions))
		for _, s := range a.Sessions {
			out.Sessions = append(out.Sessions, v1.Session{
				ID: s.ID, Title: s.Title, URL: s.URL, Speakers: s.Speakers, Start: s.Start, End: s.End,
				Location: s.Location, Venue: s.Venue, Level: s.Level, Room: s.Room,
			})
		}
	}
	if a.Applied != nil {
		out.Profile = &v1.Profile{Applied: make([]v1.Applied, 0, len(a.Applied))}
		for _, ap := range a.Applied {
			out.Profile.Applied = append(out.Profile.Applied, v1.Applied{Chip: ap.Chip, Effect: ap.Effect})
		}
	}
	return out
}
