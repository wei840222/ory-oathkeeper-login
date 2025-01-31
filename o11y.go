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
	"go.uber.org/fx"

	"github.com/wei840222/login-server/config"
)

func RunO11yHTTPServer(lc fx.Lifecycle) {
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", viper.GetString(config.ConfigKeyO11yHost), viper.GetInt(config.ConfigKeyO11yPort)),
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
