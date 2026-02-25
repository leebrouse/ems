package observation

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Shutdown func(ctx context.Context) error

type Config struct {
	Enabled     bool
	Endpoint    string
	Protocol    string
	SampleRatio float64

	Namespace   string
	Environment string
	Version     string
}

func ConfigFromViper() Config {
	return Config{
		Enabled:     viper.GetBool("observation.enabled"),
		Endpoint:    viper.GetString("observation.otlp.endpoint"),
		Protocol:    viper.GetString("observation.otlp.protocol"),
		SampleRatio: viper.GetFloat64("observation.sampling.ratio"),
		Namespace:   viper.GetString("observation.resource.namespace"),
		Environment: viper.GetString("observation.resource.environment"),
		Version:     viper.GetString("observation.resource.version"),
	}
}

func InitFromViper(ctx context.Context, serviceName string) (Shutdown, error) {
	return Init(ctx, serviceName, ConfigFromViper())
}

func Init(ctx context.Context, serviceName string, cfg Config) (Shutdown, error) {
	if !cfg.Enabled {
		return func(context.Context) error { return nil }, nil
	}
	if serviceName == "" {
		return nil, errors.New("empty serviceName")
	}
	if cfg.Endpoint == "" {
		return nil, errors.New("empty observation.otlp.endpoint")
	}
	proto := strings.ToLower(strings.TrimSpace(cfg.Protocol))
	if proto == "" {
		proto = "grpc"
	}
	ratio := cfg.SampleRatio
	if ratio <= 0 || ratio > 1 {
		ratio = 0.1
	}
	ns := cfg.Namespace
	if ns == "" {
		ns = "ems"
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			attribute.String("service.namespace", ns),
			attribute.String("deployment.environment", cfg.Environment),
			attribute.String("service.version", cfg.Version),
		),
	)
	if err != nil {
		return nil, err
	}

	var tp *sdktrace.TracerProvider
	var traceShutdown Shutdown
	switch proto {
	case "http":
		exp, err := otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(cfg.Endpoint),
			otlptracehttp.WithInsecure(),
		)
		if err != nil {
			return nil, err
		}
		tp = newTraceProvider(exp, res, ratio)
		traceShutdown = tp.Shutdown
	case "grpc":
		exp, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(cfg.Endpoint),
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			return nil, err
		}
		tp = newTraceProvider(exp, res, ratio)
		traceShutdown = tp.Shutdown
	default:
		return nil, errors.New("unsupported observation.otlp.protocol")
	}

	mpShutdown, err := initMeterProvider(ctx, proto, cfg.Endpoint, res)
	if err != nil {
		_ = traceShutdown(ctx)
		return nil, err
	}

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mpShutdown.mp)
	b3Prop := b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader))
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
		b3Prop,
	))

	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_ = mpShutdown.shutdown(ctx)
		return traceShutdown(ctx)
	}, nil
}

func GinMiddleware(serviceName string) gin.HandlerFunc {
	return otelgin.Middleware(serviceName)
}

func GRPCServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{grpc.StatsHandler(otelgrpc.NewServerHandler())}
}

func GRPCDialOptions() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}
}

func newTraceProvider(exp sdktrace.SpanExporter, res *resource.Resource, ratio float64) *sdktrace.TracerProvider {
	return sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exp),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))),
	)
}

type meterProviderShutdown struct {
	mp       *sdkmetric.MeterProvider
	shutdown Shutdown
}

func initMeterProvider(ctx context.Context, proto, endpoint string, res *resource.Resource) (*meterProviderShutdown, error) {
	switch proto {
	case "http":
		exp, err := otlpmetrichttp.New(ctx,
			otlpmetrichttp.WithEndpoint(endpoint),
			otlpmetrichttp.WithInsecure(),
		)
		if err != nil {
			return nil, err
		}
		mp := sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp, sdkmetric.WithInterval(10*time.Second))),
		)
		return &meterProviderShutdown{mp: mp, shutdown: mp.Shutdown}, nil
	case "grpc":
		exp, err := otlpmetricgrpc.New(ctx,
			otlpmetricgrpc.WithEndpoint(endpoint),
			otlpmetricgrpc.WithInsecure(),
		)
		if err != nil {
			return nil, err
		}
		mp := sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp, sdkmetric.WithInterval(10*time.Second))),
		)
		return &meterProviderShutdown{mp: mp, shutdown: mp.Shutdown}, nil
	default:
		return nil, errors.New("unsupported observation.otlp.protocol")
	}
}
