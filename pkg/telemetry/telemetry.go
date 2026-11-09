// Package telemetry is the OpenTelemetry adapter: it builds the SDK providers that ship traces, metrics
// and logs over OTLP, and the slog handler and HTTP transport the wiring hands to the other layers.
package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	lognoop "go.opentelemetry.io/otel/log/noop"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	"github.com/auroq/cnh/pkg/config"
)

// Version is the build's version, set with -ldflags "-X github.com/auroq/cnh/pkg/telemetry.Version=...".
var Version = "dev"

const (
	scope          = "github.com/auroq/cnh/pkg/telemetry"
	metricInterval = 15 * time.Second
)

// Providers is what Setup built; each is a noop when no endpoint is configured.
type Providers struct {
	Tracer   trace.TracerProvider
	Meter    metric.MeterProvider
	Logger   otellog.LoggerProvider
	Resource *resource.Resource
	shutdown []func(context.Context) error
}

// Shutdown flushes and stops every provider, returning every error that occurred.
func (p *Providers) Shutdown(ctx context.Context) error {
	var errs []error
	for _, fn := range p.shutdown {
		if err := fn(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Setup builds the SDK providers for cfg, installs them as the OpenTelemetry globals, and installs the
// W3C trace context and baggage propagators. With an empty endpoint every provider is a noop.
func Setup(ctx context.Context, cfg config.Telemetry) (*Providers, error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	p := &Providers{Tracer: tracenoop.NewTracerProvider(), Meter: metricnoop.NewMeterProvider(), Logger: lognoop.NewLoggerProvider()}
	endpoint := strings.TrimSpace(cfg.OTLPEndpoint())
	if endpoint != "" {
		res, err := resource.New(ctx,
			resource.WithFromEnv(),
			resource.WithHost(),
			resource.WithAttributes(
				attribute.String("service.name", cfg.ServiceName()),
				attribute.String("service.version", Version),
			),
		)
		if err != nil {
			return nil, err
		}
		p.Resource = res
		if err := p.connect(ctx, endpoint, res); err != nil {
			return nil, err
		}
	}
	otel.SetTracerProvider(p.Tracer)
	otel.SetMeterProvider(p.Meter)
	global.SetLoggerProvider(p.Logger)
	return p, nil
}

func (p *Providers) connect(ctx context.Context, endpoint string, res *resource.Resource) error {
	hasScheme := strings.Contains(endpoint, "://")
	traceExp, err := otlptracegrpc.New(ctx, traceOptions(endpoint, hasScheme)...)
	if err != nil {
		return err
	}
	metricExp, err := otlpmetricgrpc.New(ctx, metricOptions(endpoint, hasScheme)...)
	if err != nil {
		return err
	}
	logExp, err := otlploggrpc.New(ctx, logOptions(endpoint, hasScheme)...)
	if err != nil {
		return err
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(traceExp), sdktrace.WithResource(res))
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp, sdkmetric.WithInterval(metricInterval))), sdkmetric.WithResource(res))
	lp := sdklog.NewLoggerProvider(sdklog.WithProcessor(sdklog.NewBatchProcessor(logExp)), sdklog.WithResource(res))
	p.Tracer, p.Meter, p.Logger = tp, mp, lp
	p.shutdown = []func(context.Context) error{tp.Shutdown, mp.Shutdown, lp.Shutdown}
	return nil
}

func traceOptions(endpoint string, hasScheme bool) []otlptracegrpc.Option {
	if hasScheme {
		return []otlptracegrpc.Option{otlptracegrpc.WithEndpointURL(endpoint)}
	}
	return []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(endpoint), otlptracegrpc.WithInsecure()}
}

func metricOptions(endpoint string, hasScheme bool) []otlpmetricgrpc.Option {
	if hasScheme {
		return []otlpmetricgrpc.Option{otlpmetricgrpc.WithEndpointURL(endpoint)}
	}
	return []otlpmetricgrpc.Option{otlpmetricgrpc.WithEndpoint(endpoint), otlpmetricgrpc.WithInsecure()}
}

func logOptions(endpoint string, hasScheme bool) []otlploggrpc.Option {
	if hasScheme {
		return []otlploggrpc.Option{otlploggrpc.WithEndpointURL(endpoint)}
	}
	return []otlploggrpc.Option{otlploggrpc.WithEndpoint(endpoint), otlploggrpc.WithInsecure()}
}

// Transport wraps an HTTP round tripper so every call to peer becomes a child span named "peer METHOD /path".
func Transport(peer string, base http.RoundTripper, opts ...otelhttp.Option) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	opts = append([]otelhttp.Option{
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return peer + " " + r.Method + " " + r.URL.Path
		}),
		otelhttp.WithSpanOptions(trace.WithAttributes(attribute.String("peer.service", peer))),
	}, opts...)
	return otelhttp.NewTransport(base, opts...)
}

// Handler returns an slog handler that writes every record to base, with trace_id and span_id added
// when the context carries a span, and also sends it to the global OpenTelemetry logger.
func Handler(base slog.Handler) slog.Handler {
	return handler(base, global.GetLoggerProvider())
}

func handler(base slog.Handler, lp otellog.LoggerProvider) slog.Handler {
	return &tee{base: base, otel: otelslog.NewHandler(scope, otelslog.WithLoggerProvider(lp), otelslog.WithVersion(Version))}
}

type tee struct {
	base slog.Handler
	otel slog.Handler
}

func (t *tee) Enabled(ctx context.Context, level slog.Level) bool {
	return t.base.Enabled(ctx, level) || t.otel.Enabled(ctx, level)
}

func (t *tee) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	if t.base.Enabled(ctx, r.Level) {
		local := r.Clone()
		if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
			local.AddAttrs(slog.String("trace_id", sc.TraceID().String()), slog.String("span_id", sc.SpanID().String()))
		}
		errs = append(errs, t.base.Handle(ctx, local))
	}
	if t.otel.Enabled(ctx, r.Level) {
		errs = append(errs, t.otel.Handle(ctx, r))
	}
	return errors.Join(errs...)
}

func (t *tee) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &tee{base: t.base.WithAttrs(attrs), otel: t.otel.WithAttrs(attrs)}
}

func (t *tee) WithGroup(name string) slog.Handler {
	return &tee{base: t.base.WithGroup(name), otel: t.otel.WithGroup(name)}
}
