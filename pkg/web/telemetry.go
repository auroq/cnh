package web

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	scope          = "github.com/auroq/cnh/pkg/web"
	chipUsedMetric = "widedoor.profile.chip_used"
	attrChip       = attribute.Key("chip")
)

// WithMeterProvider is where the profile metric goes; tests pass an in-memory one.
func WithMeterProvider(mp metric.MeterProvider) Option {
	return func(s *Server) { s.meter = mp }
}

func (s *Server) chipUsed() metric.Int64Counter {
	s.once.Do(func() {
		mp := s.meter
		if mp == nil {
			mp = otel.GetMeterProvider()
		}
		s.chipCounter, _ = mp.Meter(scope).Int64Counter(chipUsedMetric,
			metric.WithDescription("Questions asked with this chip in the attendee's profile; one count per chip, never a combination"))
	})
	return s.chipCounter
}
