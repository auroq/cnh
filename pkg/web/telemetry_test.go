package web

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/auroq/cnh/pkg/ask"
)

func traced(t *testing.T, a Asker) (*tracetest.SpanRecorder, func(target string)) {
	t.Helper()
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	srv := httptest.NewServer(New(a, slog.New(slog.NewTextHandler(io.Discard, nil)), WithTracerProvider(tp)).Handler())
	t.Cleanup(srv.Close)
	return recorder, func(target string) {
		res, err := http.Get(srv.URL + target)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
	}
}

func TestRequestSpans(t *testing.T) {
	t.Run("when the API answers a question", func(t *testing.T) {
		recorder, get := traced(t, fakeAsker{answer: ask.Answer{Summary: "Quiet room 260A: Salt Palace, Level 2."}})
		get("/api/ask?q=quiet+room&from=Grand+Ballroom+J")
		spans := recorder.Ended()
		if len(spans) == 0 {
			t.Fatal("no spans")
		}
		t.Run("it should record exactly one span", func(t *testing.T) {
			if len(spans) != 1 {
				t.Fatalf("spans %v", spanNames(spans))
			}
		})
		t.Run("it should name the span after the route pattern", func(t *testing.T) {
			if spans[0].Name() != "GET /api/ask" {
				t.Fatalf("span %q", spans[0].Name())
			}
		})
		t.Run("it should tag the span with the route, not the query", func(t *testing.T) {
			if !hasAttribute(spans[0], "http.route", "/api/ask") {
				t.Fatalf("attributes %v", spans[0].Attributes())
			}
		})
	})

	t.Run("when the probes are answered by the real use case", func(t *testing.T) {
		recorder, get := traced(t, &ask.Service{Store: pingStore{}})
		get("/healthz")
		get("/readyz")
		t.Run("it should record no span", func(t *testing.T) {
			if n := len(recorder.Ended()); n != 0 {
				t.Fatalf("%d spans: %v", n, spanNames(recorder.Ended()))
			}
		})
	})
}

type pingStore struct{}

func (pingStore) Ping(context.Context) error                                  { return nil }
func (pingStore) LocateRoom(context.Context, string) (ask.Here, error)        { return ask.Here{}, nil }
func (pingStore) FindPlaces(context.Context, string) ([]ask.Place, error)     { return nil, nil }
func (pingStore) FindSessions(context.Context, string) ([]ask.Session, error) { return nil, nil }

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
