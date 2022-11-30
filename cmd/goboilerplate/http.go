package main

import (
	"context"
	"net/http"
	_ "net/http/pprof" // for profiling purpose
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

	handler "github.com/kurio/boilerplate-go/internal/http"
)

var (
	httpCMD = &cobra.Command{
		Use:   "http",
		Short: "Start the HTTP server.",
		Run:   runHttp,
	}

	e *echo.Echo
)

func initHttpApp() {
	initConfig()
	initOtel()

	initMysqlDB()
	initMongoClient()
	initRedisClient()
	initHttpClient()

	// expiryConf := goboilerplate.ExpiryConf{
	// 	goboilerplate.DurationShort: config.Redis.ShortExpirationTime,
	// 	goboilerplate.DurationLong: config.Redis.LongExpirationTime,
	// }
	// cacher := redis.NewRedisCacher(redisClient, expiryConf, app)

	// initService()

	e = echo.New()
	if config.Debug {
		e.Debug = true
	} else {
		e.HideBanner = true
		e.HidePort = true
	}

	e.Server.ReadTimeout = config.HTTP.Server.ReadTimeout
	e.Server.WriteTimeout = config.HTTP.Server.WriteTimeout

	e.HTTPErrorHandler = handler.ErrorHandler

	/*****
	Statsd
	******/
	var statsdClient *statsd.Client
	var err error
	if config.StatsdURL != "" {
		statsdClient, err = statsd.New(config.StatsdURL)
		if err != nil {
			statsdClient = nil
			logrus.Errorf("error initializing statsd client: %+v", err)
		}
	}

	/*********
	Prometheus
	**********/
	p := handler.NewPrometheus(app, handler.URLSkipper)

	responseTimeMiddleware := handler.NewResponseTimeMiddleware(statsdClient, handler.URLSkipper)

	/*********
	Middleware
	**********/
	/* uncomment if needed, set to debug, or set skipper
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:       true,
		LogRoutePath: true,
		LogStatus:    true,
		LogLatency:   true,
		LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
			logrus.WithFields(logrus.Fields{
				"URI":        values.URI,
				"status":     values.Status,
				"latency_ms": fmt.Sprintf("%.3f", values.Latency.Seconds()*1000),
				"route":      values.RoutePath,
			}).Info("request info")

			return nil
		},
	}))
	*/

	e.Use(otelecho.Middleware(app))
	p.Use(e)
	responseTimeMiddleware.Use(e)
	e.Use(handler.TimeoutMiddleware(config.HTTP.Server.Timeout))

	// Basic handlers...
	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	}).Name = "ping"

	e.GET("/_version", func(context echo.Context) error {
		return context.String(http.StatusOK, gitCommit)
	}).Name = "version"

	handler.AddSomeHandler(e)
}

func runHttp(cmd *cobra.Command, args []string) {
	initHttpApp()

	if config.Debug {
		logrus.Warn("adding /debug for profiling")
		e.GET("/debug/*", echo.WrapHandler(http.DefaultServeMux)).Name = "debug"
	}

	const address = ":7723"
	go func() {
		logrus.Infof("starting HTTP server at %s", address)
		if err := e.Start(address); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("error http server: %+v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	logrus.Debug("waiting on signal...")
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		logrus.Debug("closing mysql client...")
		if err := mysqlDB.Close(); err != nil {
			logrus.Errorf("error closing mysql client: %+v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if tracerProvider == nil {
			return
		}
		logrus.Debug("shutting down tracer provider...")
		if err := tracerProvider.Shutdown(ctx); err != nil {
			logrus.Errorf("error shutting down tracer provider: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if meterProvider == nil {
			return
		}
		logrus.Debug("shutting down meter provider...")
		if err := meterProvider.Shutdown(ctx); err != nil {
			logrus.Errorf("error shutting down meter provider")
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		logrus.Info("gracefully shutting down HTTP server...")
		if err := e.Shutdown(ctx); err != nil {
			logrus.Fatalf("error shutting down server: %+v", err)
		}
	}()

	wg.Wait()
	logrus.Info("gracefully shut down")
}
