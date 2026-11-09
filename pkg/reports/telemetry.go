package reports

import (
	"context"
	"errors"

	"github.com/auroq/cnh/pkg/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const scope = "github.com/auroq/cnh/pkg/reports"

const (
	attrVerdict = attribute.Key("verdict")
	attrKind    = attribute.Key("kind")
	attrReason  = attribute.Key("reason")
)

type instruments struct {
	received metric.Int64Counter
	rejected metric.Int64Counter
}

func (s *Service) instruments() *instruments {
	if s.inst != nil {
		return s.inst
	}
	mp := s.Meter
	if mp == nil {
		mp = otel.GetMeterProvider()
	}
	m := mp.Meter(scope)
	in := &instruments{}
	in.received, _ = m.Int64Counter("widedoor.reports.received", metric.WithDescription("Attendee reports accepted, by verdict and the kind of place"))
	in.rejected, _ = m.Int64Counter("widedoor.reports.rejected", metric.WithDescription("Attendee reports refused, by reason"))
	s.inst = in
	return in
}

func (s *Service) received(ctx context.Context, verdict model.Verdict, kind string) {
	s.instruments().received.Add(ctx, 1, metric.WithAttributes(attrVerdict.String(string(verdict)), attrKind.String(kind)))
}

func (s *Service) rejected(ctx context.Context, err error) {
	s.instruments().rejected.Add(ctx, 1, metric.WithAttributes(attrReason.String(Reason(err))))
}

// Reason names a rejection for a metric label: one word, never the note.
func Reason(err error) string {
	for _, r := range reasons {
		if errors.Is(err, r.err) {
			return r.name
		}
	}
	return "error"
}

var reasons = []struct {
	err  error
	name string
}{
	{ErrUnknownPlace, "unknown_place"},
	{ErrVerdict, "verdict"},
	{ErrNoteRequired, "note_required"},
	{ErrNoteLength, "note_length"},
	{ErrNoteURL, "note_url"},
	{ErrNoteBlocked, "note_blocked"},
	{ErrHoneypot, "honeypot"},
	{ErrToken, "token"},
	{ErrRateLimited, "rate_limited"},
}
