package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

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

// ResponseTimeMiddleware is used to send response_time metrics to statsd.
func ResponseTimeMiddleware(statsdClient *statsd.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			startTime := time.Now()

			err := next(c)
			if err != nil {
				return err
			}

			endpoint := "UNKNOWN"
			for _, r := range c.Echo().Routes() {
				if r.Method == c.Request().Method && r.Path == c.Path() {
					endpoint = r.Name
				}
			}

			responseTime := time.Since(startTime)

			tags := []string{
				// TODO: update
				"service:goboilerplate",
				fmt.Sprintf("operation:%s", endpoint),
			}

			if err := statsdClient.Gauge("http.request.duration.seconds", responseTime.Seconds(), tags, float64(1.0)); err != nil {
				logrus.Warningf("error statsdClient.Gauge: %+v", err)
			}

			if err := statsdClient.Incr("http.request.total", tags, float64(1)); err != nil {
				logrus.Warningf("error statsdClient.Incr: %+v", err)
			}

			return nil
		}
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
