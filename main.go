package main

import (
	"context"
	"fmt"
	"log"

	"dagger.io/dagger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
)

// InitializeTracer initializes the OpenTelemetry tracer with a simple stdout exporter
func InitializeTracer() *trace.TracerProvider {
	// Create a stdout trace exporter
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
	}

	// Create a trace provider with the stdout exporter
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	return tp
}

func main() {
	// Initialize the OpenTelemetry tracer
	tracerProvider := InitializeTracer()
	defer func() {
		// Gracefully shut down the tracer provider
		if err := tracerProvider.Shutdown(context.Background()); err != nil {
			log.Fatalf("failed to shutdown TracerProvider: %v", err)
		}
	}()

	// Create a new tracer
	tracer := otel.Tracer("dagger-otel-example")

	// Start a root span
	ctx, span := tracer.Start(context.Background(), "main-process")
	defer span.End()

	// Create a Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(log.Default().Writer()))
	if err != nil {
		log.Fatalf("failed to connect to Dagger: %v", err)
	}
	defer client.Close()

	// Emit a custom span for Dagger pipeline execution
	ctx, stepSpan := tracer.Start(ctx, "dagger-pipeline")
	stepSpan.SetAttributes(attribute.String("step", "Alpine Container build with Dagger!"))

	// Start a Container build process with Dagger
	container := client.Container().
		From("alpine:latest").
		WithExec([]string{"echo", "Hello from Dagger!"})

	// Execute the container process
	_, err = container.Sync(ctx)
	if err != nil {
		log.Fatalf("failed to execute container: %v", err)
	}

	// Print a success message
	fmt.Println("Executed container successfully!")

	// End the Dagger span
	stepSpan.End()

	// Log the completion span
	span.AddEvent("Execution completed successfully!")
}
