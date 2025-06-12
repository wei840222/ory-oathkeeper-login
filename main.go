package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ipfans/fxlogger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/fx"

	"github.com/wei840222/ory-oathkeeper-login/config"
	"github.com/wei840222/ory-oathkeeper-login/handler"
)

var (
	flagReplacer = strings.NewReplacer(".", "-")
)

var rootCmd = &cobra.Command{
	Use:   "ory-oathkeeper-login",
	Short: "Login Callback Server for Ory Oathkeeper.",
	Long:  `Login Callback Server for Ory Oathkeeper for handling login callback to third party web apps.`,
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		if err := config.InitViper(); err != nil {
			return err
		}

		viper.BindPFlag(config.KeyLogLevel, cmd.Flags().Lookup(flagReplacer.Replace(config.KeyLogLevel)))
		viper.BindPFlag(config.KeyLogFormat, cmd.Flags().Lookup(flagReplacer.Replace(config.KeyLogFormat)))
		viper.BindPFlag(config.KeyLogColor, cmd.Flags().Lookup(flagReplacer.Replace(config.KeyLogColor)))

		config.InitZerolog()

		b, err := json.Marshal(viper.AllSettings())
		if err != nil {
			return err
		}
		log.Debug().Str("config", string(b)).Msg("config loaded")

		return nil
	},
	Run: func(*cobra.Command, []string) {
		app := fx.New(
			fx.Provide(
				NewCache,
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
	rootCmd.PersistentFlags().String(flagReplacer.Replace(config.KeyLogLevel), "info", "Log level")
	rootCmd.PersistentFlags().String(flagReplacer.Replace(config.KeyLogFormat), "console", "Log format")
	rootCmd.PersistentFlags().Bool(flagReplacer.Replace(config.KeyLogColor), true, "Log color")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
