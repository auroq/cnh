package web

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/auroq/cnh/pkg/ask"
)

func TestRequestSpans(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	asker := fakeAsker{answer: ask.Answer{Summary: "Quiet room 260A: Salt Palace, Level 2."}}

	t.Run("when the API answers a question", func(t *testing.T) {
		recorder.Reset()
		get(t, asker, "/api/ask?q=quiet+room&from=Grand+Ballroom+J")
		spans := recorder.Ended()
		t.Run("it should record one span named after the route pattern", func(t *testing.T) {
			if len(spans) != 1 || spans[0].Name() != "GET /api/ask" {
				t.Fatalf("spans %v", spanNames(spans))
			}
		})
		t.Run("it should tag the span with the route, not the query", func(t *testing.T) {
			if !hasAttribute(spans[0], "http.route", "/api/ask") {
				t.Fatalf("attributes %v", spans[0].Attributes())
			}
		})
	})

	t.Run("when a probe is answered", func(t *testing.T) {
		recorder.Reset()
		get(t, asker, "/healthz")
		get(t, asker, "/readyz")
		t.Run("it should not record a span", func(t *testing.T) {
			if n := len(recorder.Ended()); n != 0 {
				t.Fatalf("%d spans: %v", n, spanNames(recorder.Ended()))
			}
		})
	})
}

func spanNames(spans []sdktrace.ReadOnlySpan) []string {
	out := make([]string, 0, len(spans))
	for _, s := range spans {
		out = append(out, s.Name())
	}
	return out
}

func hasAttribute(span sdktrace.ReadOnlySpan, key, value string) bool {
	for _, kv := range span.Attributes() {
		if string(kv.Key) == key && kv.Value.AsString() == value {
			return true
		}
	}
	return false
}
