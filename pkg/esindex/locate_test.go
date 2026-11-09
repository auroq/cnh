package esindex

import (
	"context"
	"encoding/json"
	"testing"
)

const grandJ = `{"hits":{"hits":[
  {"_source":{"id":"room-salt-palace-level-1-grand-ballroom-j","kind":"room","name":"Grand Ballroom J","venue":"Salt Palace","level":"Level 1","verified":true}}
]}}`

func TestLocatePlace(t *testing.T) {
	t.Run("when a door code's id is looked up", func(t *testing.T) {
		store, got := fakeES(t, grandJ)
		here, err := store.LocatePlace(context.Background(), "room-salt-palace-level-1-grand-ballroom-j")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should search the places index", func(t *testing.T) {
			if got.path != "/places/_search" {
				t.Fatalf("path %s", got.path)
			}
		})
		t.Run("it should ask for exactly that id", func(t *testing.T) {
			b, _ := json.Marshal(got.body["query"])
			if string(b) != `{"term":{"id":"room-salt-palace-level-1-grand-ballroom-j"}}` {
				t.Fatalf("query %s", b)
			}
		})
		t.Run("it should return where the place is", func(t *testing.T) {
			if here.Room != "Grand Ballroom J" || here.Venue != "Salt Palace" || here.Level != "Level 1" {
				t.Fatalf("here %+v", here)
			}
		})
	})

	t.Run("when the id is unknown", func(t *testing.T) {
		store, _ := fakeES(t, `{"hits":{"hits":[]}}`)
		here, err := store.LocatePlace(context.Background(), "room-nowhere")
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should return nowhere", func(t *testing.T) {
			if here.Venue != "" || here.Room != "" {
				t.Fatalf("here %+v", here)
			}
		})
	})
}
