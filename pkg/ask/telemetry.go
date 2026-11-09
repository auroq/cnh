package ask

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const scope = "github.com/auroq/cnh/pkg/ask"

var (
	tracer = otel.Tracer(scope)
	meter  = otel.Meter(scope)

	askDuration, _ = meter.Float64Histogram("widedoor.ask.duration",
		metric.WithUnit("ms"),
		metric.WithDescription("How long answering one question took"))
	askAnswers, _ = meter.Int64Counter("widedoor.ask.answers",
		metric.WithDescription("Answers served, by the kind of the top place and whether it was verified"))
	unverifiedServed, _ = meter.Int64Counter("widedoor.ask.unverified_served",
		metric.WithDescription("Answers whose top place nobody has verified on the floor yet"))
)

const (
	attrHasLocation    = attribute.Key("has_location")
	attrAnswered       = attribute.Key("answered")
	attrKind           = attribute.Key("kind")
	attrVerified       = attribute.Key("verified")
	attrQuestionLength = attribute.Key("question.length")
	attrHereVenue      = attribute.Key("here.venue")
	attrHereLevel      = attribute.Key("here.level")
	attrPlacesCount    = attribute.Key("places.count")
	attrSessionsCount  = attribute.Key("sessions.count")
	attrAnswerVerified = attribute.Key("answer.verified")
)

// topKind names what the attendee was told first: a place's kind, "session", or "none".
func topKind(a Answer) (kind string, verified bool) {
	switch {
	case len(a.Places) > 0:
		return a.Places[0].Kind, a.Places[0].Verified
	case len(a.Sessions) > 0:
		return "session", true
	default:
		return "none", false
	}
}
