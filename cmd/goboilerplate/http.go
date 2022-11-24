package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof" // for profiling purpose
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

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

func init() {
	rootCMD.AddCommand(httpCMD)
}

func initHttpApp() {
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
	// promAddress := os.Getenv("PROMETHEUS_ADDRESS")
	// TODO: change serviceName
	p := handler.NewPrometheus(app, handler.URLSkipper)
	p.Use(e)

	e.Use(
		handler.ResponseTimeMiddleware(statsdClient),
		handler.TimeoutMiddleware(config.HTTP.Server.Timeout),
		handler.ErrorMiddleware(),
	)

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	}).Name = "ping"

	e.GET("/_version", func(context echo.Context) error {
		return context.String(http.StatusOK, gitCommit)
	}).Name = "version"

	e.GET("/articles", func(c echo.Context) error {
		return c.JSON(http.StatusOK, make([]interface{}, 0))
	}).Name = "fetchArticles"

	e.GET("/articles/:id", func(c echo.Context) error {
		return c.String(http.StatusOK, c.Param("id"))
	}).Name = "getArticle"

	e.GET("/something/:duration", func(c echo.Context) error {
		sleepTime, err := strconv.ParseInt(c.Param("duration"), 10, 8)
		if err != nil {
			logrus.Errorf("error parsing duration: %+v", err)
			sleepTime = 1
		}
		logrus.Debugf("sleep for %d ms", sleepTime)
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)
		logrus.Debugf("returning...")
		return c.String(http.StatusOK, "ok")
	})

	// TODO: handler.AddSomeHandler(e, ...)
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
		logrus.Info("gracefully shutting down HTTP server...")
		if err := e.Shutdown(ctx); err != nil {
			logrus.Fatalf("error shutting down server: %+v", err)
		}
	}()

	wg.Wait()
	logrus.Info("gracefully shut down")
}
