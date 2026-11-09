package ask_test

import (
	"context"
	"testing"
	"time"

	"github.com/auroq/cnh/pkg/ask"
	"github.com/auroq/cnh/pkg/esindex"
	"github.com/auroq/cnh/testing/helpers"
)

func TestAsk(t *testing.T) {
	t.Parallel()

	t.Run("when a room is resolved", func(t *testing.T) {
		t.Parallel()
		es := helpers.ESClient(t)
		places := helpers.NewIndex(t, esindex.PlacesMapping)
		sessions := helpers.NewIndex(t, esindex.SessionsMapping)
		helpers.LoadPlaces(t, es, places)
		helpers.LoadSessions(t, es, sessions)
		svc := &ask.Service{Store: &esindex.Store{Client: es, PlacesIndex: places, SessionsIndex: sessions}}
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		a, err := svc.Ask(ctx, "nearest all-gender restroom", "255 BC")
		if err != nil {
			t.Fatalf("Ask: %v", err)
		}
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
}
