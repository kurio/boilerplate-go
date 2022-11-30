package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	goboilerplate "github.com/kurio/boilerplate-go"
	handler "github.com/kurio/boilerplate-go/internal/http"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func getErrorResponse(t *testing.T, err error) string {
	t.Helper()

	if err == nil {
		return ""
	}

	b, err := json.Marshal(map[string]interface{}{
		"message": err.Error(),
	})
	require.NoError(t, err)

	return string(b)
}

func slowOperation(ctx context.Context, dur time.Duration) error {
	select {
	case <-time.After(dur):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestErrorHandling(t *testing.T) {
	tests := map[string]struct {
		handler             func(c echo.Context) error
		expectedStatus      int
		expectedBody        string
		expectedLogIncludes []string
	}{
		"context deadline exceeded": {
			handler: func(c echo.Context) error {
				ctx, cancel := context.WithTimeout(c.Request().Context(), 1*time.Millisecond)
				defer cancel()

				return slowOperation(ctx, 10*time.Millisecond)
			},
			expectedStatus: http.StatusRequestTimeout,
		},
		"context canceled": {
			handler: func(c echo.Context) error {
				ctx, cancel := context.WithTimeout(c.Request().Context(), 200*time.Millisecond)
				cancel()

				return slowOperation(ctx, 10*time.Millisecond)
			},
			expectedStatus: http.StatusRequestTimeout,
		},
		"bad request": {
			handler: func(c echo.Context) error {
				return goboilerplate.ConstraintErrorf("invalid int: %s", "a")
			},
			expectedStatus:      http.StatusBadRequest,
			expectedBody:        getErrorResponse(t, goboilerplate.ConstraintErrorf("invalid int: %s", "a")),
			expectedLogIncludes: []string{},
		},
		"not found": {
			handler: func(c echo.Context) error {
				return goboilerplate.ErrNotFound
			},
			expectedStatus:      http.StatusNotFound,
			expectedBody:        getErrorResponse(t, goboilerplate.ErrNotFound),
			expectedLogIncludes: []string{},
		},
		"wrapped errors": {
			handler: func(c echo.Context) error {
				err := errors.New("unexpected error")
				err = errors.Wrap(err, "first layer")
				err = errors.Wrap(err, "second layer")
				err = errors.Wrap(err, "third layer")
				return err
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   getErrorResponse(t, errors.New("unexpected error")),
			expectedLogIncludes: []string{
				"unexpected error",
				"first layer",
				"second layer",
				"third layer",
			},
		},
		"handler returns HTTPError": {
			handler: func(c echo.Context) error {
				return echo.NewHTTPError(http.StatusBadGateway, "bad gateway error")
			},
			expectedStatus:      http.StatusBadGateway,
			expectedBody:        `{"message":"bad gateway error"}`,
			expectedLogIncludes: []string{"bad gateway error"},
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			e := echo.New()
			e.HTTPErrorHandler = handler.ErrorHandler
			e.GET("/my-endpoint", test.handler)

			buf := new(bytes.Buffer)
			logrus.SetOutput(buf)

			req := httptest.NewRequest(echo.GET, "/my-endpoint", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			require.Equal(t, test.expectedStatus, rec.Code)
			if test.expectedBody != "" {
				require.Equal(t, test.expectedBody, strings.Trim(rec.Body.String(), "\n"))
			}

			logMessage := buf.String()
			if len(test.expectedLogIncludes) == 0 {
				require.Empty(t, logMessage)
			} else {
				for _, s := range test.expectedLogIncludes {
					require.Contains(t, logMessage, s)
				}
			}
		})
	}
}

func TestErrorHandling_HEAD(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = handler.ErrorHandler
	e.HEAD("/my-endpoint", func(c echo.Context) error {
		return errors.New("unexpected error")
	})

	buf := new(bytes.Buffer)
	logrus.SetOutput(buf)

	req := httptest.NewRequest(echo.HEAD, "/my-endpoint", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Empty(t, rec.Body)
	require.Contains(t, buf.String(), "unexpected error")
}
