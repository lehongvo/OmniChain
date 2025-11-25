package middleware

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
)

// TracingMiddleware creates OpenTelemetry tracing middleware
func TracingMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract trace context from headers
		ctx := otel.GetTextMapPropagator().Extract(c.Context(), propagation.HeaderCarrier(c.GetReqHeaders()))

		// Start span
		tracer := otel.Tracer("api-gateway")
		ctx, span := tracer.Start(ctx, c.Route().Path)
		defer span.End()

		// Set span attributes
		span.SetAttributes(
			attribute.String("http.method", c.Method()),
			attribute.String("http.path", c.Path()),
			attribute.String("http.route", c.Route().Path),
		)

		// Store span in context
		c.Locals("span", span)
		c.Locals("trace_id", span.SpanContext().TraceID().String())

		// Inject trace context into response headers
		headers := make(propagation.HeaderCarrier)
		c.Response().Header.VisitAll(func(key, value []byte) {
			headers[string(key)] = []string{string(value)}
		})
		otel.GetTextMapPropagator().Inject(ctx, headers)
		// Set headers back
		for key, values := range headers {
			for _, value := range values {
				c.Response().Header.Set(key, value)
			}
		}

		// Process request
		err := c.Next()

		// Set status code
		span.SetAttributes(attribute.Int("http.status_code", c.Response().StatusCode()))

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		return err
	}
}
