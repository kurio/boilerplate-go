package main

import (
	"net/http"
	"os"

	// for profiling purpose
	_ "net/http/pprof"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	handler "github.com/kurio/boilerplate-go/internal/http"
)

const address = ":7723"

var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "Start the HTTP server.",
	Run: func(cmd *cobra.Command, args []string) {
		e := echo.New()

		/******
		Prometheus
		******/
		promAddress := os.Getenv("PROMETHEUS_ADDRESS")
		// TODO: change serviceName
		p := handler.NewPrometheus("catalina", handler.URLSkipper)
		if promAddress != "" {
			p.SetListenAddress(promAddress)
		}
		p.Use(e)

		/******
		Statsd
		******/
		statsdURL := os.Getenv("STATSD_URL")
		if statsdURL != "" {
			statsdClient, err := statsd.New(statsdURL)
			if err == nil {
				e.Use(handler.ResponseTimeMiddleware(statsdClient))
			} else {
				log.Errorf("error initializing statsd client: %+v", err)
			}
		} else {
			log.Warning("STATSD_URL is not set")
		}

		e.Use(handler.ErrorMiddleware())

		// TODO: handler.AddSomeHandler(e, ...)

		e.GET("/ping", func(c echo.Context) error {
			return c.String(http.StatusOK, "pong")
		})

		errCh := make(chan error)
		go func(ch chan error) {
			log.Info("Starting HTTP server at ", address)
			errCh <- e.Start(address)
		}(errCh)

		go func(ch chan error) {
			errCh <- http.ListenAndServe(":6060", nil)
		}(errCh)

		for {
			log.Fatal(<-errCh)
		}
	},
}

func init() {
	rootCmd.AddCommand(httpCmd)
}
