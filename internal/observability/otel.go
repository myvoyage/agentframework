package observability

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// InitTracer initializes OpenTelemetry tracer with OTLP exporter
func InitTracer(serviceName string) error {
	// 获取 OTLP 端点配置
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:4318" // 默认端点
	}

	// 创建 OTLP 导出器
	var clientOpt []otlptracehttp.Option
	if strings.HasPrefix(endpoint, "http://") {
		clientOpt = append(clientOpt, otlptracehttp.WithInsecure())
	}

	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(strings.TrimPrefix(endpoint, "http://")),
		otlptracehttp.WithTimeout(30*time.Second),
	)

	exporter, err := otlptrace.New(context.Background(), client)
	if err != nil {
		return errors.New("failed to create OTLP exporter")
	}

	// 配置资源
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String("1.0.0"),
			semconv.DeploymentEnvironmentKey.String(getEnvOrDefault("OTEL_ENVIRONMENT", "development")),
		),
	)
	if err != nil {
		return errors.New("failed to create resource")
	}

	// 创建 trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// 设置全局 trace provider
	otel.SetTracerProvider(tp)

	// 设置传播器
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// 注册关闭函数
	registerShutdownHook(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			println("Error shutting down trace provider:", err.Error())
		}
	})

	return nil
}

// StartSpan starts a new span with the given name
func StartSpan(ctx context.Context, name string) (context.Context, func()) {
	tracer := otel.Tracer("agent-framework")
	ctx, span := tracer.Start(ctx, name)
	return ctx, func() {
		span.End()
	}
}

// getEnvOrDefault gets an environment variable with a default value
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// shutdownHooks holds functions to be called when the program exits
var shutdownHooks []func()

// registerShutdownHook registers a function to be called on program exit
func registerShutdownHook(hook func()) {
	shutdownHooks = append(shutdownHooks, hook)
}

// Shutdown calls all registered shutdown hooks
func Shutdown() {
	for _, hook := range shutdownHooks {
		hook()
	}
}
