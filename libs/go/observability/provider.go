package observability

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Config holds observability bootstrap configuration.
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	LogLevel       string
	OTLPEndpoint   string
	EnableTracing  bool
	EnableMetrics  bool
}

// Provider bundles logger, tracer, and metrics providers for graceful shutdown.
type Provider struct {
	Logger         *slog.Logger
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
	shutdownFuncs  []func(context.Context) error
}

// NewProvider initializes structured logging and OpenTelemetry providers.
func NewProvider(ctx context.Context, cfg Config) (*Provider, error) {
	logger := newLogger(cfg.LogLevel, cfg.ServiceName, cfg.Environment)

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create otel resource: %w", err)
	}

	provider := &Provider{Logger: logger}

	if cfg.EnableTracing {
		tp, shutdown, err := newTracerProvider(ctx, res, cfg.OTLPEndpoint)
		if err != nil {
			return nil, err
		}
		provider.TracerProvider = tp
		provider.shutdownFuncs = append(provider.shutdownFuncs, shutdown)
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
	}

	if cfg.EnableMetrics {
		mp, shutdown, err := newMeterProvider(res)
		if err != nil {
			return nil, err
		}
		provider.MeterProvider = mp
		provider.shutdownFuncs = append(provider.shutdownFuncs, shutdown)
		otel.SetMeterProvider(mp)
	}

	return provider, nil
}

// Shutdown flushes telemetry exporters.
func (p *Provider) Shutdown(ctx context.Context) error {
	var errs []error
	for _, fn := range p.shutdownFuncs {
		if err := fn(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("observability shutdown: %v", errs)
	}
	return nil
}

func newLogger(level, service, env string) *slog.Logger {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey {
				return slog.Attr{Key: "timestamp", Value: attr.Value}
			}
			return attr
		},
	})
	return slog.New(handler).With(
		slog.String("service", service),
		slog.String("environment", env),
	)
}

func newTracerProvider(ctx context.Context, res *resource.Resource, endpoint string) (*sdktrace.TracerProvider, func(context.Context) error, error) {
	opts := []otlptracehttp.Option{}
	if endpoint != "" {
		opts = append(opts, otlptracehttp.WithEndpoint(strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://")))
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("create otlp trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	shutdown := func(shutdownCtx context.Context) error {
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		return tp.Shutdown(shutdownCtx)
	}

	return tp, shutdown, nil
}

func newMeterProvider(res *resource.Resource) (*sdkmetric.MeterProvider, func(context.Context) error, error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, nil, fmt.Errorf("create prometheus exporter: %w", err)
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(exporter),
	)

	shutdown := func(shutdownCtx context.Context) error {
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		return mp.Shutdown(shutdownCtx)
	}

	return mp, shutdown, nil
}
