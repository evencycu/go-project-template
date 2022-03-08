package trace

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"

	m800schema "gitlab.com/cake/golibs/intercom/schemas/v1.0.0"
	oteljaeger "go.opentelemetry.io/otel/exporters/jaeger"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

func InitTracer(componentName, localNamespace string, sampleRate float64, opts ...tracesdk.TracerProviderOption) (*tracesdk.TracerProvider, error) {
	// Please setup with environment variable
	exporter, err := oteljaeger.New(oteljaeger.WithCollectorEndpoint())
	if err != nil {
		return nil, err
	}

	allOpts := []tracesdk.TracerProviderOption{
		// Sampler rate with builtin sampler
		// Reference: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/sdk.md#built-in-samplers
		// TODO: check why environment variable doesn't work
		tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.TraceIDRatioBased(sampleRate))),
		tracesdk.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(componentName),
				m800schema.M800NamespaceKey.String(localNamespace),
			),
		),
	}
	allOpts = append(allOpts, opts...)
	allOpts = append(allOpts, tracesdk.WithBatcher(exporter))

	tp := tracesdk.NewTracerProvider(allOpts...)

	// TODO: Add logger
	// otel.SetLogger()

	// TODO: Add error handler
	// otel.SetErrorHandler()

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp, nil
}
