package main

import (
	"context"
	"fmt"
	"log"

	"dagger.io/dagger"
	"dagger.io/dagger/telemetry"
	"go.opentelemetry.io/otel/attribute"
)

func main() {
	// Initialize telemetry
	otelCtx := telemetry.Init(context.Background(), telemetry.Config{Detect: true})

	// Create a tracer using telemetry.Tracer
	tracer := telemetry.Tracer(otelCtx, "dagger-otel-example")

	// Start a root span
	ctx, span := tracer.Start(otelCtx, "main-process")
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
