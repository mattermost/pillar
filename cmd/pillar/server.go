package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	logrus "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cloud "github.com/mattermost/mattermost-cloud/model"

	"github.com/mattermost/pillar/api"
	"github.com/mattermost/pillar/utils"
)

const defaultLocalServerAPI = "http://localhost:8078"

var instanceID string

func init() {
	viper.SetEnvPrefix("PILLAR")
	viper.AutomaticEnv()

	instanceID = utils.NewID()

	// Core Settings
	serverCmd.PersistentFlags().String("listen", ":8078", "The interface and port on which to listen on the API.")
	serverCmd.PersistentFlags().Bool("debug", false, "Whether to output debug logs.")

	// Dev Settings
	serverCmd.PersistentFlags().Bool("dev", false, "Set to run in dev mode.")

	// Provisioner Settings
	serverCmd.PersistentFlags().String("cloud-url", viper.GetString("CLOUD_URL"), "Endpoint where the Cloud Provisioning Server can be reached (include the scheme and port number) | ENV: PILLAR_CLOUD_URL")
}

// Config holds the configuration for pillar.
type Config struct {
	DevMode   bool
	DebugLogs bool
	CloudURL  string
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the customer web server.",
	RunE: func(command *cobra.Command, args []string) error {
		command.SilenceUsage = true

		debug, _ := command.Flags().GetBool("debug")
		if debug {
			logger.SetLevel(logrus.DebugLevel)
		}

		logger := logger.WithField("instance", instanceID)

		var config Config

		config.CloudURL, _ = command.Flags().GetString("cloud-url")

		dev, _ := command.Flags().GetBool("dev")
		if dev {
			logger.Debug("Using dev configuration")
			config.DevMode = true
		}

		// Require a Provisioner connection
		if config.CloudURL == "" {
			return errors.New("a hostname and port number where a cloud provisioner endpoint can be found are required")
		}

		wd, err := os.Getwd()
		if err != nil {
			wd = "error getting working directory"
			logger.WithError(err).Error("Unable to get current working directory")
		}

		listen, _ := command.Flags().GetString("listen")

		logger.WithFields(logrus.Fields{
			"workingdirectory": wd,
			"debug":            debug,
			"backend":          listen,
			"dev":              dev,
		}).Info("Starting Pillar")

		publicRouter := mux.NewRouter()

		api.Register(publicRouter, &api.Context{
			Logger:      logger,
			CloudClient: cloud.NewClient(config.CloudURL),
		})

		startServer := func(router *mux.Router, listen string) *http.Server {
			srv := &http.Server{
				Addr:           listen,
				Handler:        router,
				ReadTimeout:    180 * time.Second,
				WriteTimeout:   180 * time.Second,
				IdleTimeout:    time.Second * 180,
				MaxHeaderBytes: 1 << 20,
				ErrorLog:       log.New(&logrusWriter{logger}, "", 0),
			}

			go func() {
				logger.WithField("addr", srv.Addr).Info("Listening")
				err := srv.ListenAndServe()
				if err != nil && err != http.ErrServerClosed {
					logger.WithError(err).Error("Failed to listen and serve")
				}
			}()
			return srv
		}

		publicServer := startServer(publicRouter, listen)

		c := make(chan os.Signal, 1)
		// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
		// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
		signal.Notify(c, os.Interrupt)

		// Block until we receive our signal.
		<-c
		logger.Info("Shutting down")

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		publicServer.Shutdown(ctx)

		return nil
	},
}
