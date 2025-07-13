package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ipfans/fxlogger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/fx"

	"github.com/wei840222/ory-oathkeeper-login/config"
	"github.com/wei840222/ory-oathkeeper-login/handler"
)

var rootCmd = &cobra.Command{
	Use:   "ory-oathkeeper-login",
	Short: "Login Callback Server for Ory Oathkeeper.",
	Long:  `Login Callback Server for Ory Oathkeeper for handling login callback to third party web apps.`,
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		logger := log.With().Str("logger", "cobra").Logger()

		if err := config.InitViper(); err != nil {
			return err
		}

		config.InitCobraPFlag(cmd)
		config.InitZerolog()

		logger.Info().Any("config", viper.AllSettings()).Msg("config loaded")

		return nil
	},
	Run: func(*cobra.Command, []string) {
		app := fx.New(
			fx.Provide(
				NewCache,
				NewMeterProvider,
				NewTracerProvider,
				NewGinEngine,
			),
			fx.Invoke(
				handler.RegisterLoginHandler,
				handler.RegisterSessionHandler,
				RunO11yHTTPServer,
			),
			fx.WithLogger(fxlogger.WithZerolog(log.Logger)),
		)

		app.Run()
	},
}

func main() {
	rootCmd.PersistentFlags().String(config.FlagReplacer.Replace(config.KeyLogLevel), "info", "Log level")
	rootCmd.PersistentFlags().String(config.FlagReplacer.Replace(config.KeyLogFormat), "console", "Log format")
	rootCmd.PersistentFlags().Bool(config.FlagReplacer.Replace(config.KeyLogColor), true, "Log color")

	rootCmd.PersistentFlags().String(config.FlagReplacer.Replace(config.KeyO11yHost), "0.0.0.0", "Observability server host")
	rootCmd.PersistentFlags().Int(config.FlagReplacer.Replace(config.KeyO11yPort), 9090, "Observability server port")

	rootCmd.PersistentFlags().String(config.FlagReplacer.Replace(config.KeyGinMode), "debug", "Gin mode")

	rootCmd.PersistentFlags().String(config.FlagReplacer.Replace(config.KeyHTTPHost), "0.0.0.0", "HTTP server host")
	rootCmd.PersistentFlags().Int(config.FlagReplacer.Replace(config.KeyHTTPPort), 8080, "HTTP server port")

	rootCmd.PersistentFlags().Duration(config.FlagReplacer.Replace(config.KeyCacheTTL), 15*time.Minute, "Cache TTL")
	rootCmd.PersistentFlags().String(config.FlagReplacer.Replace(config.KeyCacheRedisHost), "", "Cache Redis host")
	rootCmd.PersistentFlags().Int(config.FlagReplacer.Replace(config.KeyCacheRedisPort), 6379, "Cache Redis port")
	rootCmd.PersistentFlags().String(config.FlagReplacer.Replace(config.KeyCacheRedisPassword), "", "Cache Redis password")
	rootCmd.PersistentFlags().Int(config.FlagReplacer.Replace(config.KeyCacheRedisDB), 0, "Cache Redis database")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
