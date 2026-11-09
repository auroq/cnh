// Package ingest is the use case that fills the search store: fetch the schedule, load the curated places, index both.
package ingest

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

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
// One span covers the run with a child per step, and the counters say what reached the store.
func (r *Runner) Run(ctx context.Context) (Report, error) {
	start := time.Now()
	ctx, span := tracer.Start(ctx, "ingest.Run")
	defer span.End()
	rep, err := r.run(ctx)
	span.SetAttributes(
		attrSessionsIndexed.Int(rep.Sessions.Count),
		attrPlacesIndexed.Int(rep.Places.Count),
		attrDescribeFailures.Int(rep.DescribeFailed),
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	ingestDuration.Record(ctx, float64(time.Since(start))/float64(time.Millisecond), metric.WithAttributes(attrSucceeded.Bool(err == nil && !rep.Failed())))
	return rep, err
}

func (r *Runner) run(ctx context.Context) (Report, error) {
	var rep Report
	if err := r.Indexer.EnsureIndexes(ctx); err != nil {
		return rep, fmt.Errorf("ensure indexes: %w", err)
	}
	sessions, err := step(ctx, "ingest.fetch-index", func(ctx context.Context) ([]schedule.Session, error) { return r.Source.Index(ctx) })
	if err != nil {
		return rep, fmt.Errorf("fetch schedule: %w", err)
	}
	r.log().InfoContext(ctx, "fetched index", "sessions", len(sessions))
	if !r.SkipDetails {
		rep.DescribeFailed, _ = step(ctx, "ingest.describe", func(ctx context.Context) (int, error) {
			failed := r.Source.Describe(ctx, sessions, func(id string, err error) {
				if err != nil {
					r.log().WarnContext(ctx, "describe failed", "session", id, "err", err.Error())
				}
			})
			trace.SpanFromContext(ctx).SetAttributes(attrDescribeFailures.Int(failed))
			return failed, nil
		})
		detailFailures.Add(ctx, int64(rep.DescribeFailed))
		r.log().InfoContext(ctx, "fetched details", "failed", rep.DescribeFailed)
	}
	rep.Sessions, err = step(ctx, "ingest.index-sessions", func(ctx context.Context) (Result, error) { return r.Indexer.IndexSessions(ctx, sessions) })
	if err != nil {
		return rep, fmt.Errorf("index sessions: %w", err)
	}
	sessionsIndexed.Add(ctx, int64(rep.Sessions.Sent-rep.Sessions.Failed))
	r.log().InfoContext(ctx, "indexed sessions", "sent", rep.Sessions.Sent, "failed", rep.Sessions.Failed, "count", rep.Sessions.Count)

	curated, err := r.Places.Places()
	if err != nil {
		return rep, fmt.Errorf("load places: %w", err)
	}
	rooms := venue.RoomsFromLocations(schedule.Locations(sessions))
	rep.Curated, rep.Rooms = len(curated), len(rooms)
	rep.Places, err = step(ctx, "ingest.index-places", func(ctx context.Context) (Result, error) {
		return r.Indexer.IndexPlaces(ctx, append(curated, rooms...))
	})
	if err != nil {
		return rep, fmt.Errorf("index places: %w", err)
	}
	r.log().InfoContext(ctx, "indexed places", "curated", rep.Curated, "rooms", rep.Rooms, "failed", rep.Places.Failed, "count", rep.Places.Count)
	return rep, nil
}

func step[T any](ctx context.Context, name string, fn func(context.Context) (T, error)) (T, error) {
	ctx, span := tracer.Start(ctx, name)
	defer span.End()
	out, err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return out, err
}
