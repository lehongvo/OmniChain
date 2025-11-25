package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/onichange/pos-system/pkg/logger"
)

// TracerProvider wraps OpenTelemetry tracer provider
type TracerProvider struct {
	tp        *tracesdk.TracerProvider
	tracer    trace.Tracer
	logger    *logger.Logger
}

// NewTracerProvider creates a new OpenTelemetry tracer provider with Jaeger exporter
func NewTracerProvider(serviceName, jaegerURL string, log *logger.Logger) (*TracerProvider, error) {
	// Create Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerURL)))
	if err != nil {
		return nil, fmt.Errorf("failed to create Jaeger exporter: %w", err)
	}

	// Create resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider with sampling (1% in production)
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(res),
		tracesdk.WithSampler(tracesdk.TraceIDRatioBased(0.01)), // 1% sampling
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer := tp.Tracer(serviceName)

	log.Infof("OpenTelemetry tracer initialized for service: %s", serviceName)

	return &TracerProvider{
		tp:     tp,
		tracer: tracer,
		logger: log,
	}, nil
}

// GetTracer returns the tracer
func (t *TracerProvider) GetTracer() trace.Tracer {
	return t.tracer
}

// Shutdown shuts down the tracer provider
func (t *TracerProvider) Shutdown(ctx context.Context) error {
	return t.tp.Shutdown(ctx)
}

// StartSpan starts a new span
func StartSpan(ctx context.Context, tracer trace.Tracer, name string) (context.Context, trace.Span) {
	return tracer.Start(ctx, name)
}

