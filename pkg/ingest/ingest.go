// Package ingest is the use case that fills the search store: fetch the schedule, load the curated places, index both.
package ingest

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/auroq/cnh/pkg/schedule"
	"github.com/auroq/cnh/pkg/venue"
)

// Source is where the schedule comes from.
type Source interface {
	Index(ctx context.Context) ([]schedule.Session, error)
	// Describe fills each session's Description in place and returns how many could not be fetched.
	Describe(ctx context.Context, sessions []schedule.Session, report func(id string, err error)) int
}

// Places is where the curated access features come from.
type Places interface {
	Places() ([]venue.Place, error)
}

// Indexer is the store's write side.
type Indexer interface {
	EnsureIndexes(ctx context.Context) error
	IndexSessions(ctx context.Context, sessions []schedule.Session) (Result, error)
	IndexPlaces(ctx context.Context, places []venue.Place) (Result, error)
}

// Result is what one bulk write did.
type Result struct {
	Sent   int
	Failed int
	Count  int
}

// Report is what one run did.
type Report struct {
	Sessions       Result
	Places         Result
	DescribeFailed int
	Curated        int
	Rooms          int
}

// Failed is true when any document was rejected by the store.
func (r Report) Failed() bool { return r.Sessions.Failed+r.Places.Failed > 0 }

type Runner struct {
	Source      Source
	Places      Places
	Indexer     Indexer
	SkipDetails bool
	Log         *slog.Logger
}

func (r *Runner) log() *slog.Logger {
	if r.Log == nil {
		return slog.Default()
	}
	return r.Log
}

// Run indexes the schedule, then the curated places together with the rooms the schedule names.
func (r *Runner) Run(ctx context.Context) (Report, error) {
	var rep Report
	if err := r.Indexer.EnsureIndexes(ctx); err != nil {
		return rep, fmt.Errorf("ensure indexes: %w", err)
	}
	sessions, err := r.Source.Index(ctx)
	if err != nil {
		return rep, fmt.Errorf("fetch schedule: %w", err)
	}
	r.log().Info("fetched index", "sessions", len(sessions))
	if !r.SkipDetails {
		rep.DescribeFailed = r.Source.Describe(ctx, sessions, func(id string, err error) {
			if err != nil {
				r.log().Warn("describe failed", "session", id, "err", err.Error())
			}
		})
		r.log().Info("fetched details", "failed", rep.DescribeFailed)
	}
	if rep.Sessions, err = r.Indexer.IndexSessions(ctx, sessions); err != nil {
		return rep, fmt.Errorf("index sessions: %w", err)
	}
	r.log().Info("indexed sessions", "sent", rep.Sessions.Sent, "failed", rep.Sessions.Failed, "count", rep.Sessions.Count)

	curated, err := r.Places.Places()
	if err != nil {
		return rep, fmt.Errorf("load places: %w", err)
	}
	rooms := venue.RoomsFromLocations(schedule.Locations(sessions))
	rep.Curated, rep.Rooms = len(curated), len(rooms)
	if rep.Places, err = r.Indexer.IndexPlaces(ctx, append(curated, rooms...)); err != nil {
		return rep, fmt.Errorf("index places: %w", err)
	}
	r.log().Info("indexed places", "curated", rep.Curated, "rooms", rep.Rooms, "failed", rep.Places.Failed, "count", rep.Places.Count)
	return rep, nil
}
