package reports_test

import (
	"context"
	"testing"
	"time"

	"github.com/auroq/cnh/pkg/esindex"
	"github.com/auroq/cnh/pkg/model"
	"github.com/auroq/cnh/pkg/reports"
	"github.com/auroq/cnh/testing/helpers"
)

var placesMapping = map[string]any{
	"mappings": map[string]any{
		"properties": map[string]any{
			"id":   map[string]any{"type": "keyword"},
			"kind": map[string]any{"type": "keyword"},
		},
	},
}

func store(t *testing.T) *esindex.Reports {
	t.Helper()
	es := helpers.ESClient(t)
	s := &esindex.Reports{Client: es, ReportsIndex: helpers.NewIndex(t, esindex.ReportsMapping), PlacesIndex: helpers.NewIndex(t, placesMapping)}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := es.Put(ctx, s.PlacesIndex, "elevators-hyatt", map[string]any{"id": "elevators-hyatt", "kind": "elevator"}); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestReportsIndex(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC().Truncate(time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)

	t.Run("when two attendees say the Hyatt elevators work and one is later hidden", func(t *testing.T) {
		t.Parallel()
		s := store(t)
		svc := &reports.Service{Store: s, Now: func() time.Time { return now }}
		for _, token := range []string{"a", "b"} {
			if _, err := svc.Submit(ctx, reports.Submission{PlaceID: "elevators-hyatt", Verdict: model.VerdictWorks, Token: token, IP: "198.51.100." + token}); err != nil {
				t.Fatal(err)
			}
		}
		before, err := svc.StatusFor(ctx, []string{"elevators-hyatt"})
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should read back as verified by two attendees", func(t *testing.T) {
			if got := before["elevators-hyatt"]; got.Status != model.StatusVerified || got.Count != 2 {
				t.Fatalf("status %+v", got)
			}
		})
		t.Run("and a volunteer hides one of them", func(t *testing.T) {
			recent, err := s.Recent(ctx, now.Add(-time.Hour))
			if err != nil || len(recent) != 2 {
				t.Fatalf("recent %+v, %v", recent, err)
			}
			if err := s.Hide(ctx, recent[0].ID); err != nil {
				t.Fatal(err)
			}
			after, err := svc.StatusFor(ctx, []string{"elevators-hyatt"})
			if err != nil {
				t.Fatal(err)
			}
			t.Run("it should fall back to one attendee, not yet verified", func(t *testing.T) {
				if got := after["elevators-hyatt"]; got.Status != model.StatusNone || got.Count != 1 {
					t.Fatalf("status %+v", got)
				}
			})
		})
	})

	t.Run("when a report names a place that is not in the places index", func(t *testing.T) {
		t.Parallel()
		svc := &reports.Service{Store: store(t), Now: func() time.Time { return now }}
		_, err := svc.Submit(ctx, reports.Submission{PlaceID: "narnia", Verdict: model.VerdictWorks, Token: "z", IP: "198.51.100.9"})
		t.Run("it should be dropped as an unknown place", func(t *testing.T) {
			if err == nil || err.Error() != reports.ErrUnknownPlace.Error() {
				t.Fatalf("err %v", err)
			}
		})
	})
}
