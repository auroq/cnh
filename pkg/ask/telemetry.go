package ask

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const scope = "github.com/auroq/cnh/pkg/ask"

const (
	attrHasLocation    = attribute.Key("has_location")
	attrAnswered       = attribute.Key("answered")
	attrKind           = attribute.Key("kind")
	attrNeed           = attribute.Key("need")
	attrVerified       = attribute.Key("verified")
	attrQuestionLength = attribute.Key("question.length")
	attrHereVenue      = attribute.Key("here.venue")
	attrHereLevel      = attribute.Key("here.level")
	attrPlacesCount    = attribute.Key("places.count")
	attrSessionsCount  = attribute.Key("sessions.count")
	attrAnswerVerified = attribute.Key("answer.verified")
)

type instruments struct {
	duration         metric.Float64Histogram
	answers          metric.Int64Counter
	unverifiedServed metric.Int64Counter
}

func (s *Service) tracer() trace.Tracer {
	if s.Tracer == nil {
		return otel.Tracer(scope)
	}
	return s.Tracer.Tracer(scope)
}

func (s *Service) instruments() *instruments {
	s.once.Do(func() {
		mp := s.Meter
		if mp == nil {
			mp = otel.GetMeterProvider()
		}
		m := mp.Meter(scope)
		s.inst = &instruments{}
		s.inst.duration, _ = m.Float64Histogram("widedoor.ask.duration",
			metric.WithUnit("ms"),
			metric.WithDescription("How long answering one question took"))
		s.inst.answers, _ = m.Int64Counter("widedoor.ask.answers",
			metric.WithDescription("Answers served, by the kind of the top place and whether it was verified"))
		s.inst.unverifiedServed, _ = m.Int64Counter("widedoor.ask.unverified_served",
			metric.WithDescription("Answers whose top place nobody has verified on the floor yet"))
	})
	return s.inst
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

func needID(a Answer) string {
	if a.Need == nil {
		return ""
	}
	return a.Need.ID
}

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
