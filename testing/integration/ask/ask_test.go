package ask_test

import (
	"context"
	"testing"
	"time"

	"github.com/auroq/cnh/pkg/ask"
	"github.com/auroq/cnh/pkg/esindex"
	"github.com/auroq/cnh/testing/helpers"
)

func service(t *testing.T) *ask.Service {
	t.Helper()
	es := helpers.ESClient(t)
	places := helpers.NewIndex(t, esindex.PlacesMapping)
	sessions := helpers.NewIndex(t, esindex.SessionsMapping)
	helpers.LoadPlaces(t, es, places)
	helpers.LoadSessions(t, es, sessions)
	return &ask.Service{Store: &esindex.Store{Client: es, PlacesIndex: places, SessionsIndex: sessions}}
}

func askFrom(t *testing.T, svc *ask.Service, question, from string) ask.Answer {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	a, err := svc.Ask(ctx, question, from)
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	return a
}

func TestAsk(t *testing.T) {
	t.Parallel()

	t.Run("when a room is resolved", func(t *testing.T) {
		t.Parallel()
		a := askFrom(t, service(t), "nearest all-gender restroom", "255 BC")
		t.Run("it should resolve 255 BC to Salt Palace Level 2", func(t *testing.T) {
			if a.Here.Venue != "Salt Palace" || a.Here.Level != "Level 2" {
				t.Fatalf("here = %+v", a.Here)
			}
		})
		t.Run("it should put places on its level first", func(t *testing.T) {
			if len(a.Places) == 0 || a.Places[0].Venue != "Salt Palace" || a.Places[0].Level != "Level 2" {
				t.Fatalf("places = %+v", a.Places)
			}
		})
	})

	t.Run("when the attendee in 255 BC asks for the nearest elevator", func(t *testing.T) {
		t.Parallel()
		a := askFrom(t, service(t), "nearest elevator", "255 BC")
		t.Run("it should answer the elevator, not the restroom on the same level", func(t *testing.T) {
			if len(a.Places) == 0 || a.Places[0].ID != "elevators-salt-palace" {
				t.Fatalf("places = %+v", a.Places)
			}
		})
		t.Run("it should list only elevators", func(t *testing.T) {
			for _, p := range a.Places {
				if p.Kind != "elevator" {
					t.Fatalf("places = %+v", a.Places)
				}
			}
		})
	})
}
