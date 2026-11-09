package ingest

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const scope = "github.com/auroq/cnh/pkg/ingest"

const (
	attrSucceeded        = attribute.Key("succeeded")
	attrSessionsIndexed  = attribute.Key("sessions.indexed")
	attrPlacesIndexed    = attribute.Key("places.indexed")
	attrDescribeFailures = attribute.Key("describe.failures")
)

type instruments struct {
	sessionsIndexed metric.Int64Counter
	detailFailures  metric.Int64Counter
	duration        metric.Float64Histogram
}

func (r *Runner) tracer() trace.Tracer {
	if r.Tracer == nil {
		return otel.Tracer(scope)
	}
	return r.Tracer.Tracer(scope)
}

func (r *Runner) instruments() *instruments {
	r.once.Do(func() {
		mp := r.Meter
		if mp == nil {
			mp = otel.GetMeterProvider()
		}
		m := mp.Meter(scope)
		r.inst = &instruments{}
		r.inst.sessionsIndexed, _ = m.Int64Counter("widedoor.ingest.sessions_indexed",
			metric.WithDescription("Schedule sessions the store accepted"))
		r.inst.detailFailures, _ = m.Int64Counter("widedoor.ingest.detail_failures",
			metric.WithDescription("Session pages that could not be fetched, so the session was indexed without its description"))
		r.inst.duration, _ = m.Float64Histogram("widedoor.ingest.duration",
			metric.WithUnit("ms"),
			metric.WithDescription("How long one ingest run took"))
	})
	return r.inst
}

func step[T any](ctx context.Context, tracer trace.Tracer, name string, fn func(context.Context) (T, error)) (T, error) {
	ctx, span := tracer.Start(ctx, name)
	defer span.End()
	out, err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return out, err
}
