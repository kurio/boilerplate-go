package http

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	goboilerplate "github.com/kurio/boilerplate-go"
)

// ResponseTimeMiddleware is used to send response_time metrics to statsd.
func ResponseTimeMiddleware(statsdClient *statsd.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			startTime := time.Now()

			err := next(c)
			if err != nil {
				return err
			}

			path := c.Request().URL.Path

			metricName := getMetricName(c.Request().Method + " " + path)
			if metricName == "" {
				return nil
			}

			responseTime := time.Since(startTime)

			tags := []string{
				"service:boilerplate-go",
				fmt.Sprintf("operation:%s", metricName),
			}

			if err := statsdClient.Gauge("http.request.duration.seconds", responseTime.Seconds(), tags, float64(1.0)); err != nil {
				log.Warningf("error statsdClient.Gauge: %+v", err)
			}

			if err := statsdClient.Incr("http.request.total", tags, float64(1)); err != nil {
				log.Warningf("error statsdClient.Incr: %+v", err)
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

			headers := []string{}
			for k, v := range c.Request().Header {
				headers = append(headers, fmt.Sprintf(`%s:%s`, k, strings.Join(v, ",")))
			}
			lg := log.WithFields(log.Fields{
				"headers": strings.Join(headers, " | "),
				"method":  c.Request().Method,
				"uri":     c.Request().RequestURI,
			})

			if e, ok := err.(*echo.HTTPError); ok {
				lg.Errorln(e.Message)
				return echo.NewHTTPError(e.Code, e.Message)
			}

			msg := err.Error()

			newErr, ok := err.(stackTracer)
			if ok {
				st := newErr.StackTrace()
				msg = fmt.Sprintf("%s\n%+v", err.Error(), st[0:3])
			}

			err = errors.Cause(err)

			if _, ok := err.(goboilerplate.ConstraintError); ok {
				lg.Errorln(msg)
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}

			switch err {
			case context.DeadlineExceeded, context.Canceled:
				lg.Errorln(msg)
				return echo.NewHTTPError(http.StatusRequestTimeout, err.Error())
			case goboilerplate.ErrNotFound:
				lg.Errorln(msg)
				return echo.NewHTTPError(http.StatusNotFound, err.Error())
			}

			lg.Errorln(msg)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
}

func getMetricName(pathURL string) string {
	pathLower := strings.ToLower(pathURL)

	m := map[*regexp.Regexp]string{
		// TODO: mapping of endpoints to metric name.
		// regexp.MustCompile(`^get /something$`):             "fetch_something",
		// regexp.MustCompile(`^get /something/([\w-])+$`):    "get_something",
		// regexp.MustCompile(`^post /something$`):            "create_something",
		// regexp.MustCompile(`^delete /something/([\w-])+$`): "delete_something",
	}

	for re, metricName := range m {
		if re.MatchString(pathLower) {
			return metricName
		}
	}
	return ""
}
