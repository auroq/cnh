package ask

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func observe(t *testing.T) (*sdkmetric.ManualReader, *tracetest.SpanRecorder) {
	t.Helper()
	reader := sdkmetric.NewManualReader(sdkmetric.WithTemporalitySelector(func(sdkmetric.InstrumentKind) metricdata.Temporality { return metricdata.DeltaTemporality }))
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(mp)
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		_ = mp.Shutdown(context.Background())
		_ = tp.Shutdown(context.Background())
	})
	return reader, recorder
}

func collect(t *testing.T, reader *sdkmetric.ManualReader) map[string]metricdata.Metrics {
	t.Helper()
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatal(err)
	}
	out := map[string]metricdata.Metrics{}
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			out[m.Name] = m
		}
	}
	return out
}

func histogramCount(m metricdata.Metrics, attrs attribute.Set) uint64 {
	h, ok := m.Data.(metricdata.Histogram[float64])
	if !ok {
		return 0
	}
	for _, dp := range h.DataPoints {
		if dp.Attributes.Equals(&attrs) {
			return dp.Count
		}
	}
	return 0
}

func sumValue(m metricdata.Metrics, attrs attribute.Set) int64 {
	s, ok := m.Data.(metricdata.Sum[int64])
	if !ok {
		return 0
	}
	for _, dp := range s.DataPoints {
		if dp.Attributes.Equals(&attrs) {
			return dp.Value
		}
	}
	return 0
}

func sumTotal(m metricdata.Metrics) int64 {
	s, ok := m.Data.(metricdata.Sum[int64])
	if !ok {
		return 0
	}
	var total int64
	for _, dp := range s.DataPoints {
		total += dp.Value
	}
	return total
}

func TestAskTelemetry(t *testing.T) {
	reader, recorder := observe(t)
	ctx := context.Background()
	unverifiedElevator := Place{ID: "sp-l1-elevator", Kind: "elevator", Name: "Elevator by registration", Venue: "Salt Palace", Level: "Level 1"}
	verifiedRestroom := Place{ID: "sp-l2-restroom", Kind: "all-gender-restroom", Name: "All-gender restroom outside 255", Venue: "Salt Palace", Level: "Level 2", Verified: true}
	in255 := Here{Room: "255 BC", Venue: "Salt Palace", Level: "Level 2"}

	cases := []struct {
		name       string
		store      *fakeStore
		question   string
		from       string
		duration   attribute.Set
		answers    attribute.Set
		unverified int64
	}{
		{
			"an attendee in 255 BC is sent to an elevator nobody has checked",
			&fakeStore{room: in255, places: []Place{unverifiedElevator}}, "nearest elevator", "255 BC",
			attribute.NewSet(attrHasLocation.Bool(true), attrAnswered.Bool(true)),
			attribute.NewSet(attrKind.String("elevator"), attrVerified.Bool(false)),
			1,
		},
		{
			"an attendee without a location is sent to a restroom the organiser published",
			&fakeStore{places: []Place{verifiedRestroom}}, "all-gender restroom", "",
			attribute.NewSet(attrHasLocation.Bool(false), attrAnswered.Bool(true)),
			attribute.NewSet(attrKind.String("all-gender-restroom"), attrVerified.Bool(true)),
			0,
		},
		{
			"only sessions match",
			&fakeStore{sessions: []Session{{ID: "1", Title: "Sign Language Crash Course"}}}, "sign language talk", "",
			attribute.NewSet(attrHasLocation.Bool(false), attrAnswered.Bool(true)),
			attribute.NewSet(attrKind.String("session"), attrVerified.Bool(true)),
			0,
		},
		{
			"nothing matches",
			&fakeStore{}, "a unicorn", "",
			attribute.NewSet(attrHasLocation.Bool(false), attrAnswered.Bool(false)),
			attribute.NewSet(attrKind.String("none"), attrVerified.Bool(false)),
			0,
		},
	}
	for _, c := range cases {
		t.Run("when "+c.name, func(t *testing.T) {
			collect(t, reader)
			if _, err := (&Service{Store: c.store}).Ask(ctx, c.question, c.from); err != nil {
				t.Fatal(err)
			}
			got := collect(t, reader)
			t.Run("it should record one duration with whether a location and an answer were present", func(t *testing.T) {
				if n := histogramCount(got["widedoor.ask.duration"], c.duration); n != 1 {
					t.Fatalf("count %d for %v in %+v", n, c.duration.ToSlice(), got["widedoor.ask.duration"].Data)
				}
			})
			t.Run("it should count the answer by the kind of the top place and whether it was verified", func(t *testing.T) {
				if n := sumValue(got["widedoor.ask.answers"], c.answers); n != 1 {
					t.Fatalf("count %d for %v in %+v", n, c.answers.ToSlice(), got["widedoor.ask.answers"].Data)
				}
			})
			t.Run("it should count how often an unverified answer was served", func(t *testing.T) {
				if n := sumTotal(got["widedoor.ask.unverified_served"]); n != c.unverified {
					t.Fatalf("unverified served %d, want %d", n, c.unverified)
				}
			})
		})
	}

	t.Run("when the question is empty", func(t *testing.T) {
		collect(t, reader)
		if _, err := (&Service{Store: &fakeStore{}}).Ask(ctx, "  ", ""); err != nil {
			t.Fatal(err)
		}
		got := collect(t, reader)
		t.Run("it should not count an answer", func(t *testing.T) {
			if _, ok := got["widedoor.ask.answers"]; ok {
				t.Fatalf("recorded %+v", got["widedoor.ask.answers"].Data)
			}
		})
	})

	t.Run("when an attendee in 255 BC asks for the nearest elevator", func(t *testing.T) {
		recorder.Reset()
		if _, err := (&Service{Store: &fakeStore{room: in255, places: []Place{unverifiedElevator}}}).Ask(ctx, "nearest elevator", "255 BC"); err != nil {
			t.Fatal(err)
		}
		spans := map[string]sdktrace.ReadOnlySpan{}
		for _, s := range recorder.Ended() {
			spans[s.Name()] = s
		}
		root, ok := spans["ask.Ask"]
		if !ok {
			t.Fatalf("spans %v", names(recorder.Ended()))
		}
		t.Run("it should record one span per store call under the ask span", func(t *testing.T) {
			for _, child := range []string{"ask.LocateRoom", "ask.FindPlaces", "ask.FindSessions"} {
				if s, ok := spans[child]; !ok || s.Parent().SpanID() != root.SpanContext().SpanID() {
					t.Fatalf("%s missing or not a child of ask.Ask in %v", child, names(recorder.Ended()))
				}
			}
		})
		want := []attribute.KeyValue{
			attrQuestionLength.Int(len("nearest elevator")),
			attrHereVenue.String("Salt Palace"),
			attrHereLevel.String("Level 2"),
			attrPlacesCount.Int(1),
			attrSessionsCount.Int(0),
			attrAnswerVerified.Bool(false),
		}
		for _, kv := range want {
			t.Run("it should describe the question and the answer with "+string(kv.Key), func(t *testing.T) {
				if !hasAttribute(root.Attributes(), kv) {
					t.Fatalf("attributes %v", root.Attributes())
				}
			})
		}
	})
}

func names(spans []sdktrace.ReadOnlySpan) []string {
	out := make([]string, 0, len(spans))
	for _, s := range spans {
		out = append(out, s.Name())
	}
	return out
}

func hasAttribute(attrs []attribute.KeyValue, want attribute.KeyValue) bool {
	for _, kv := range attrs {
		if kv.Key == want.Key && kv.Value == want.Value {
			return true
		}
	}
	return false
}
