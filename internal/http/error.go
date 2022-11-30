package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	goboilerplate "github.com/kurio/boilerplate-go"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ErrorHandler always return JSON, even on debug mode
func ErrorHandler(err error, c echo.Context) {
	if err == nil {
		return
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

	code := http.StatusInternalServerError
	message := originalError.Error()
	logError := true

	switch e := originalError.(type) {
	case goboilerplate.ConstraintError:
		code = http.StatusBadRequest
		message = e.Error()
		logError = false
	case *echo.HTTPError:
		if e.Internal != nil {
			if herr, ok := e.Internal.(*echo.HTTPError); ok {
				e = herr
			}
		}

		code = e.Code
		if m, ok := e.Message.(string); ok {
			message = m
		} else {
			message = e.Error()
		}

		if code >= 500 {
			logError = true
		}
	}

	switch originalError {
	case context.DeadlineExceeded, context.Canceled:
		code = http.StatusRequestTimeout
		logError = false
	case goboilerplate.ErrNotFound:
		code = http.StatusNotFound
		logError = false
	}

	if logError {
		log.Errorf("%+v", err)
	}

	// Send response
	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead { // Issue #608
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, map[string]interface{}{"message": message})
		}
		if err != nil {
			c.Echo().Logger.Error(err)
		}
	}
}
