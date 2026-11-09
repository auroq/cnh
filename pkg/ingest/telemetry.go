package ingest

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const scope = "github.com/auroq/cnh/pkg/ingest"

var (
	tracer = otel.Tracer(scope)
	meter  = otel.Meter(scope)

	sessionsIndexed, _ = meter.Int64Counter("widedoor.ingest.sessions_indexed",
		metric.WithDescription("Schedule sessions the store accepted"))
	detailFailures, _ = meter.Int64Counter("widedoor.ingest.detail_failures",
		metric.WithDescription("Session pages that could not be fetched, so the session was indexed without its description"))
	ingestDuration, _ = meter.Float64Histogram("widedoor.ingest.duration",
		metric.WithUnit("ms"),
		metric.WithDescription("How long one ingest run took"))
)

const (
	attrSucceeded        = attribute.Key("succeeded")
	attrSessionsIndexed  = attribute.Key("sessions.indexed")
	attrPlacesIndexed    = attribute.Key("places.indexed")
	attrDescribeFailures = attribute.Key("describe.failures")
)
