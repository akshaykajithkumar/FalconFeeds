package shared

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

func InitTracer() *trace.TracerProvider {
	// Jaeger exporter setup with the correct collector endpoint configuration
	exp, err := jaeger.New(
		jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://jaeger:14268/api/traces")),
	)
	if err != nil {
		log.Fatalf("failed to create Jaeger exporter: %v", err)
	}

	// Create and configure the TracerProvider
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(attribute.String("service.name", "feed-collector")),
	)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(res),
	)

	// Set the global tracer provider
	otel.SetTracerProvider(tp)

	return tp
}
