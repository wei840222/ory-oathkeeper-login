package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"

	_ "github.com/grafana/pyroscope-go/godeltaprof/http/pprof"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"

	"github.com/wei840222/ory-oathkeeper-login/config"
)

func NewTracerProvider(lc fx.Lifecycle) (trace.TracerProvider, error) {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(config.AppName),
		),
	)
	if err != nil {
		return nil, err
	}

	exp, err := otlptracegrpc.New(context.Background())
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	otel.SetTracerProvider(tp)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return tp.Shutdown(ctx)
		},
	})

	return tp, nil
}

func RunO11yHTTPServer(lc fx.Lifecycle) {
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", viper.GetString(config.KeyO11yHost), viper.GetInt(config.KeyO11yPort)),
		Handler: mux,
	}

	var isShuttingDown bool
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		if !isShuttingDown {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Service is shutting down"))
		}
	})
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/debug/pprof/", http.DefaultServeMux)

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			go srv.Serve(ln)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			isShuttingDown = true
			return srv.Shutdown(ctx)
		},
	})

}
