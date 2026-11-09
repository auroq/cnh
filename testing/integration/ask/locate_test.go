package ask_test

import (
	"context"
	"testing"
	"time"

	"github.com/auroq/cnh/pkg/ask"
	"github.com/auroq/cnh/pkg/esindex"
	"github.com/auroq/cnh/testing/helpers"
)

func TestLocateInMyWords(t *testing.T) {
	t.Parallel()
	es := helpers.ESClient(t)
	places := helpers.NewIndex(t, esindex.PlacesMapping)
	sessions := helpers.NewIndex(t, esindex.SessionsMapping)
	helpers.LoadPlaces(t, es, places)
	helpers.LoadSessions(t, es, sessions)
	svc := &ask.Service{Store: &esindex.Store{Client: es, PlacesIndex: places, SessionsIndex: sessions}}

	cases := []struct {
		said string
		want ask.Here
	}{
		{"I just got out of Sign Language Crash Course", ask.Here{Room: "Grand Ballroom J", Venue: "Salt Palace", Level: "Level 1", Via: ask.ViaSession}},
		{"where I registered", ask.Here{Room: "Registration and badge pick-up", Venue: "Salt Palace", Via: ask.ViaAlias}},
		{"the 100 S side", ask.Here{Room: "100 South entrance", Venue: "Salt Palace", Side: "100 South", Via: ask.ViaAlias}},
		{"Grand Ballroom J", ask.Here{Room: "Grand Ballroom J", Venue: "Salt Palace", Level: "Level 1", Via: ask.ViaRoom}},
	}
	for _, c := range cases {
		t.Run("when the attendee says "+c.said, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			got, err := svc.Locate(ctx, c.said)
			if err != nil {
				t.Fatalf("Locate: %v", err)
			}
			t.Run("it should resolve to "+c.want.Room+" via "+c.want.Via, func(t *testing.T) {
				want := c.want
				want.Asked = c.said
				if got != want {
					t.Fatalf("\n got %+v\nwant %+v", got, want)
				}
			})
		})
	}
}
