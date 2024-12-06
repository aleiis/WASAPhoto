package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/XSAM/otelsql"
	"github.com/aleiis/WASAPhoto/service/config"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func connectMySQLDB(dsn string) (*sql.DB, error) {
	return otelsql.Open("mysql", dsn, otelsql.WithAttributes(
		semconv.DBSystemMySQL,
	))
}

func newExporter(ctx context.Context) (sdktrace.SpanExporter, error) {
	cfg, _ := config.GetConfig()
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.OTLP.HTTPTraceExporterEndpoint),
		otlptracehttp.WithInsecure(), // IMPORTANT!!: DO NOT USE IN PRODUCTION
	)
	if err != nil {
		return nil, err
	}
	return exporter, nil
}

func newTraceProvider(ctx context.Context) (*sdktrace.TracerProvider, error) {

	exp, err := newExporter(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create the exporter: %w", err)
	}

	// Ensure default SDK resources and the required service name are set.
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("webapi"),
		),
	)

	if err != nil {
		return nil, err
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	), nil
}
