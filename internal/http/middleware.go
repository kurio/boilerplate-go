package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/unit"

	goboilerplate "github.com/kurio/boilerplate-go"
)

// TimeoutMiddleware is a middleware that set maximum HTTP response time before considered timeout.
func TimeoutMiddleware(httpProcessTimeout time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx, cancel := context.WithTimeout(c.Request().Context(), httpProcessTimeout)
			defer cancel()

			// recompose request with a new context
			httpReq := c.Request().WithContext(ctx)
			c.SetRequest(httpReq)

			return next(c)
		}
	}
}

type ResponseTimeMiddleware struct {
	statsdClient          *statsd.Client
	responseTimeHistogram syncfloat64.Histogram
	requestCounter        syncfloat64.Counter

	Skipper middleware.Skipper
}

func NewResponseTimeMiddleware(statsdClient *statsd.Client, skipper middleware.Skipper) *ResponseTimeMiddleware {
	meter := global.Meter("")
	responseTimeHistogram, err := meter.SyncFloat64().Histogram(
		"echo_response_time",
		instrument.WithDescription("response time comes from echo middleware"),
		instrument.WithUnit(unit.Milliseconds),
	)
	if err != nil {
		logrus.Fatalf("error creating response time histogram instrument: %+v", err)
	}
	requestCounter, err := meter.SyncFloat64().Counter(
		"echo_request_counter",
		instrument.WithDescription("request counter comes from echo middleware"),
	)
	if err != nil {
		logrus.Fatalf("error creating request counter instrument: %+v", err)
	}

	if skipper == nil {
		skipper = URLSkipper
	}

	return &ResponseTimeMiddleware{
		statsdClient:          statsdClient,
		responseTimeHistogram: responseTimeHistogram,
		requestCounter:        requestCounter,
		Skipper:               skipper,
	}
}

func (m *ResponseTimeMiddleware) Use(e *echo.Echo) {
	e.Use(m.HandlerFunc)
}

func (m *ResponseTimeMiddleware) HandlerFunc(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Path() == "/metrics" {
			return next(c)
		}
		if m.Skipper(c) {
			return next(c)
		}

		startTime := time.Now()

		err := next(c)
		if err != nil {
			return err
		}

		operation := "UNKNOWN"
		for _, r := range c.Echo().Routes() {
			if r.Method == c.Request().Method && r.Path == c.Path() {
				operation = r.Name
			}
		}

		responseTime := time.Since(startTime)

		tags := []string{
			// TODO: update
			"service:goboilerplate",
			fmt.Sprintf("operation:%s", operation),
		}

		if m.statsdClient != nil {
			if err := m.statsdClient.Gauge("http.request.duration.seconds", responseTime.Seconds(), tags, float64(1.0)); err != nil {
				logrus.Warningf("error statsdClient.Gauge: %+v", err)
			}

			if err := m.statsdClient.Incr("http.request.total", tags, float64(1)); err != nil {
				logrus.Warningf("error statsdClient.Incr: %+v", err)
			}
		}

		m.responseTimeHistogram.Record(
			c.Request().Context(),
			float64(responseTime.Milliseconds()),
			attribute.Key("operation").String(operation),
		)

		m.requestCounter.Add(
			c.Request().Context(),
			1,
			attribute.Key("operation").String(operation),
		)

		return nil
	}
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// ErrorMiddleware is a function to generate http status code.
func ErrorMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil {
				return nil
			}

			originalError := errors.Cause(err)

			headers := []string{}
			for k, v := range c.Request().Header {
				headers = append(headers, fmt.Sprintf(`%s:%s`, k, strings.Join(v, ",")))
			}

			log := logrus.WithFields(logrus.Fields{
				"headers": strings.Join(headers, " | "),
				"method":  c.Request().Method,
				"uri":     c.Request().RequestURI,
			})

			if e, ok := originalError.(*echo.HTTPError); ok {
				if e.Code >= 500 {
					log.Error(e.Message)
				}

				return e
			}

			msg := originalError.Error()

			if newErr, ok := err.(stackTracer); ok {
				st := newErr.StackTrace()
				msg = fmt.Sprintf("%s\n%+v", err.Error(), st[0:3])
			}

			if _, ok := originalError.(goboilerplate.ConstraintError); ok {
				return echo.NewHTTPError(http.StatusBadRequest, originalError.Error())
			}

			switch originalError {
			case context.DeadlineExceeded, context.Canceled:
				return echo.NewHTTPError(http.StatusRequestTimeout, originalError.Error())
			case goboilerplate.ErrNotFound:
				return echo.NewHTTPError(http.StatusNotFound, originalError.Error())
			}

			log.Error(msg)

			return echo.NewHTTPError(http.StatusInternalServerError, originalError.Error())
		}
	}
}

// URLSkipper skip unwanted URL from scrapped by Prometheus
func URLSkipper(c echo.Context) bool {
	switch c.Path() {
	case "/ping", "/_version", "/metrics":
		return true
	}
	return strings.HasPrefix(c.Path(), "/debug")
}
