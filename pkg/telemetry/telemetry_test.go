package telemetry

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	lognoop "go.opentelemetry.io/otel/log/noop"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

type settings struct{ endpoint, name string }

func (s settings) OTLPEndpoint() string { return s.endpoint }
func (s settings) ServiceName() string  { return s.name }

func TestSetup(t *testing.T) {
	ctx := context.Background()

	t.Run("when no endpoint is configured", func(t *testing.T) {
		p, err := Setup(ctx, settings{"", "widedoor"})
		if err != nil {
			t.Fatal(err)
		}
		t.Run("it should hand back a noop tracer provider", func(t *testing.T) {
			if _, ok := p.Tracer.(tracenoop.TracerProvider); !ok {
				t.Fatalf("tracer provider is %T", p.Tracer)
			}
		})
		t.Run("it should hand back a noop meter provider", func(t *testing.T) {
			if _, ok := p.Meter.(metricnoop.MeterProvider); !ok {
				t.Fatalf("meter provider is %T", p.Meter)
			}
		})
		t.Run("it should hand back a noop logger provider", func(t *testing.T) {
			if _, ok := p.Logger.(lognoop.LoggerProvider); !ok {
				t.Fatalf("logger provider is %T", p.Logger)
			}
		})
		t.Run("it should shut down without error", func(t *testing.T) {
			if err := p.Shutdown(ctx); err != nil {
				t.Fatal(err)
			}
		})
	})

	t.Run("when an endpoint is configured", func(t *testing.T) {
		p, err := Setup(ctx, settings{"http://collector.invalid:4317", "widedoor"})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			cancelled, cancel := context.WithCancel(ctx)
			cancel()
			_ = p.Shutdown(cancelled)
		})
		t.Run("it should build an SDK tracer provider", func(t *testing.T) {
			if _, ok := p.Tracer.(*sdktrace.TracerProvider); !ok {
				t.Fatalf("tracer provider is %T", p.Tracer)
			}
		})
		t.Run("it should build an SDK meter provider", func(t *testing.T) {
			if _, ok := p.Meter.(*sdkmetric.MeterProvider); !ok {
				t.Fatalf("meter provider is %T", p.Meter)
			}
		})
		t.Run("it should build an SDK logger provider", func(t *testing.T) {
			if _, ok := p.Logger.(*sdklog.LoggerProvider); !ok {
				t.Fatalf("logger provider is %T", p.Logger)
			}
		})
		t.Run("it should name the service in the resource", func(t *testing.T) {
			if !hasAttribute(p.Resource.Attributes(), "service.name", "widedoor") {
				t.Fatalf("resource %v", p.Resource.Attributes())
			}
		})
		t.Run("it should put the build version in the resource", func(t *testing.T) {
			if !hasAttribute(p.Resource.Attributes(), "service.version", Version) {
				t.Fatalf("resource %v", p.Resource.Attributes())
			}
		})
	})
}

func TestInstall(t *testing.T) {
	t.Run("when the providers are installed", func(t *testing.T) {
		p, err := Setup(context.Background(), settings{"", "widedoor"})
		if err != nil {
			t.Fatal(err)
		}
		p.Install()
		fields := strings.Join(otel.GetTextMapPropagator().Fields(), " ")
		t.Run("it should propagate W3C trace context", func(t *testing.T) {
			if !strings.Contains(fields, "traceparent") {
				t.Fatalf("propagator fields %q", fields)
			}
		})
		t.Run("it should propagate W3C baggage", func(t *testing.T) {
			if !strings.Contains(fields, "baggage") {
				t.Fatalf("propagator fields %q", fields)
			}
		})
	})
}

type captureExporter struct{ records []sdklog.Record }

func (c *captureExporter) Export(_ context.Context, records []sdklog.Record) error {
	c.records = append(c.records, records...)
	return nil
}
func (c *captureExporter) Shutdown(context.Context) error   { return nil }
func (c *captureExporter) ForceFlush(context.Context) error { return nil }

func TestHandler(t *testing.T) {
	var buf bytes.Buffer
	exporter := &captureExporter{}
	lp := sdklog.NewLoggerProvider(sdklog.WithProcessor(sdklog.NewSimpleProcessor(exporter)))
	t.Cleanup(func() { _ = lp.Shutdown(context.Background()) })
	log := slog.New(handler(slog.NewJSONHandler(&buf, nil), lp))
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{0x4b, 0xf9, 0x2f, 0x35, 0x77, 0xb3, 0x4d, 0xa6, 0xa3, 0xce, 0x92, 0x9d, 0x0e, 0x0e, 0x47, 0x36},
		SpanID:     trace.SpanID{0x00, 0xf0, 0x67, 0xaa, 0x0b, 0xa9, 0x02, 0xb7},
		TraceFlags: trace.FlagsSampled,
	})

	t.Run("when a record is logged inside a span", func(t *testing.T) {
		buf.Reset()
		exporter.records = nil
		log.InfoContext(trace.ContextWithSpanContext(context.Background(), sc), "request", "path", "/api/ask")
		t.Run("it should write the trace id beside the record on the base handler", func(t *testing.T) {
			if !strings.Contains(buf.String(), `"trace_id":"4bf92f3577b34da6a3ce929d0e0e4736"`) {
				t.Fatalf("logged %s", buf.String())
			}
		})
		t.Run("it should write the span id beside the record on the base handler", func(t *testing.T) {
			if !strings.Contains(buf.String(), `"span_id":"00f067aa0ba902b7"`) {
				t.Fatalf("logged %s", buf.String())
			}
		})
		t.Run("it should send exactly one record to OpenTelemetry", func(t *testing.T) {
			if len(exporter.records) != 1 {
				t.Fatalf("exported %d records", len(exporter.records))
			}
		})
		t.Run("it should keep the record in the same trace", func(t *testing.T) {
			if len(exporter.records) == 0 || exporter.records[0].TraceID() != sc.TraceID() {
				t.Fatalf("exported %+v", exporter.records)
			}
		})
	})

	t.Run("when a record is logged outside a span", func(t *testing.T) {
		buf.Reset()
		log.Info("listening", "addr", ":8080")
		t.Run("it should not add a trace id", func(t *testing.T) {
			if strings.Contains(buf.String(), "trace_id") {
				t.Fatalf("logged %s", buf.String())
			}
		})
	})
}

func TestTransport(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	t.Cleanup(upstream.Close)

	client := &http.Client{Transport: Transport("elasticsearch", nil, otelhttp.WithTracerProvider(tp))}
	call := func(t *testing.T, ctx context.Context) {
		t.Helper()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, upstream.URL+"/places/_search", nil)
		if err != nil {
			t.Fatal(err)
		}
		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
	}

	t.Run("when Elasticsearch is called inside a trace", func(t *testing.T) {
		recorder.Reset()
		ctx, parent := tp.Tracer("test").Start(context.Background(), "ask.FindPlaces")
		call(t, ctx)
		parent.End()
		spans := recorder.Ended()
		if len(spans) != 2 {
			t.Fatalf("%d spans: %v", len(spans), spans)
		}
		child := spans[0]
		t.Run("it should name the span after the peer, the method and the path", func(t *testing.T) {
			if child.Name() != "elasticsearch GET /places/_search" {
				t.Fatalf("span %q", child.Name())
			}
		})
		t.Run("it should tag the span with the peer service", func(t *testing.T) {
			if !hasAttribute(child.Attributes(), "peer.service", "elasticsearch") {
				t.Fatalf("attributes %v", child.Attributes())
			}
		})
		t.Run("it should make the call a child of the caller's span", func(t *testing.T) {
			if child.Parent().SpanID() != parent.SpanContext().SpanID() {
				t.Fatalf("parent %s", child.Parent().SpanID())
			}
		})
	})

	t.Run("when Elasticsearch is pinged outside any trace", func(t *testing.T) {
		recorder.Reset()
		call(t, context.Background())
		t.Run("it should record no span", func(t *testing.T) {
			if n := len(recorder.Ended()); n != 0 {
				t.Fatalf("%d spans", n)
			}
		})
	})
}

func hasAttribute(attrs []attribute.KeyValue, key, value string) bool {
	for _, kv := range attrs {
		if string(kv.Key) == key && kv.Value.AsString() == value {
			return true
		}
	}
	return false
}
