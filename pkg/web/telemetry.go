package web

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	scope             = "github.com/auroq/cnh/pkg/web"
	chipUsedMetric    = "widedoor.profile.chip_used"
	chipLearnedMetric = "widedoor.profile.chip_learned"
	attrChip          = attribute.Key("chip")
)

// WithMeterProvider is where the profile metrics go; tests pass an in-memory one.
func WithMeterProvider(mp metric.MeterProvider) Option {
	return func(s *Server) { s.meter = mp }
}

func (s *Server) instruments() {
	s.once.Do(func() {
		mp := s.meter
		if mp == nil {
			mp = otel.GetMeterProvider()
		}
		meter := mp.Meter(scope)
		s.chipCounter, _ = meter.Int64Counter(chipUsedMetric,
			metric.WithDescription("Questions asked with this chip in the attendee's profile; one count per chip, never a combination"))
		s.chipLearnedCounter, _ = meter.Int64Counter(chipLearnedMetric,
			metric.WithDescription("Chips added to a profile because a question implied them; one count per chip, never a combination"))
	})
}

func (s *Server) chipUsed() metric.Int64Counter {
	s.instruments()
	return s.chipCounter
}

func (s *Server) chipLearned() metric.Int64Counter {
	s.instruments()
	return s.chipLearnedCounter
}
