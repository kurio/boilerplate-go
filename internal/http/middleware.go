package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	goboilerplate "github.com/kurio/boilerplate-go"

	"github.com/labstack/echo"
	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func errorMiddleware() echo.MiddlewareFunc {
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

			var msg string

			newErr, ok := err.(stackTracer)
			if ok {
				st := newErr.StackTrace()
				msg = fmt.Sprintf("%s\n%+v", err.Error(), st[0:3])
			} else {
				msg = err.Error()
			}

			err = errors.Cause(err)

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
