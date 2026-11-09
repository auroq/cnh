package ingest

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/auroq/cnh/pkg/schedule"
	"github.com/auroq/cnh/pkg/venue"
)

func TestRunTelemetry(t *testing.T) {
	t.Run("when a run indexes three sessions and one session page fails to fetch", func(t *testing.T) {
		reader := sdkmetric.NewManualReader()
		mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
		recorder := tracetest.NewSpanRecorder()
		tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
		t.Cleanup(func() {
			_ = mp.Shutdown(context.Background())
			_ = tp.Shutdown(context.Background())
		})
		sessions := []schedule.Session{
			{ID: "1", Title: "Keynote", Location: "Salt Palace | Level 1 | Hall 1"},
			{ID: "2", Title: "Signs", Location: "Salt Palace | Level 1 | Hall 1"},
			{ID: "3", Title: "Coat check", Location: "Salt Palace | Level 2 | East Registration"},
		}
		curated := []venue.Place{{ID: "quiet-260a", Kind: "quiet-room", Name: "260A", Venue: "Salt Palace", Source: "organiser"}}
		runner := &Runner{Source: &fakeSource{sessions: sessions}, Places: fakePlaces{places: curated}, Indexer: &fakeIndexer{}, Log: quiet, Tracer: tp, Meter: mp}
		if _, err := runner.Run(context.Background()); err != nil {
			t.Fatal(err)
		}
		var rm metricdata.ResourceMetrics
		if err := reader.Collect(context.Background(), &rm); err != nil {
			t.Fatal(err)
		}
		metrics := map[string]metricdata.Metrics{}
		for _, sm := range rm.ScopeMetrics {
			for _, m := range sm.Metrics {
				metrics[m.Name] = m
			}
		}
		spans := map[string]sdktrace.ReadOnlySpan{}
		for _, s := range recorder.Ended() {
			spans[s.Name()] = s
		}

		counters := map[string]int64{
			"widedoor.ingest.sessions_indexed": 3,
			"widedoor.ingest.detail_failures":  1,
		}
		for name, want := range counters {
			t.Run("it should count "+name, func(t *testing.T) {
				if got := total(metrics[name]); got != want {
					t.Fatalf("%s = %d, want %d", name, got, want)
				}
			})
		}
		durations, _ := metrics["widedoor.ingest.duration"].Data.(metricdata.Histogram[float64])
		t.Run("it should record one run duration", func(t *testing.T) {
			if len(durations.DataPoints) != 1 || durations.DataPoints[0].Count != 1 {
				t.Fatalf("duration %+v", durations.DataPoints)
			}
		})
		t.Run("it should mark the run duration as succeeded", func(t *testing.T) {
			want := attribute.NewSet(attrSucceeded.Bool(true))
			if len(durations.DataPoints) != 1 || !durations.DataPoints[0].Attributes.Equals(&want) {
				t.Fatalf("duration %+v", durations.DataPoints)
			}
		})
		root, ok := spans["ingest.Run"]
		if !ok {
			t.Fatalf("no run span in %v", spanNames(recorder.Ended()))
		}
		for _, step := range []string{"ingest.fetch-index", "ingest.describe", "ingest.index-sessions", "ingest.index-places"} {
			t.Run("it should record "+step+" as a child of the run span", func(t *testing.T) {
				if s, ok := spans[step]; !ok || s.Parent().SpanID() != root.SpanContext().SpanID() {
					t.Fatalf("%s missing or not a child of ingest.Run in %v", step, spanNames(recorder.Ended()))
				}
			})
		}
		want := []attribute.KeyValue{attrSessionsIndexed.Int(3), attrPlacesIndexed.Int(3), attrDescribeFailures.Int(1)}
		for _, kv := range want {
			t.Run("it should put "+string(kv.Key)+" on the run span", func(t *testing.T) {
				if !has(root.Attributes(), kv) {
					t.Fatalf("attributes %v", root.Attributes())
				}
			})
		}
	})
}

func total(m metricdata.Metrics) int64 {
	s, ok := m.Data.(metricdata.Sum[int64])
	if !ok {
		return 0
	}
	var out int64
	for _, dp := range s.DataPoints {
		out += dp.Value
	}
	return out
}

func spanNames(spans []sdktrace.ReadOnlySpan) []string {
	out := make([]string, 0, len(spans))
	for _, s := range spans {
		out = append(out, s.Name())
	}
	return out
}

func has(attrs []attribute.KeyValue, want attribute.KeyValue) bool {
	for _, kv := range attrs {
		if kv.Key == want.Key && kv.Value == want.Value {
			return true
		}
	}
	return false
}
