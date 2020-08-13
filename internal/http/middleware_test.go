package http_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	goboilerplate "github.com/kurio/boilerplate-go"
	handler "github.com/kurio/boilerplate-go/internal/http"
)

func TestErrorMiddleware(t *testing.T) {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetReportCaller(true)

	mw := handler.ErrorMiddleware()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	t.Run("with not found object", func(t *testing.T) {
		h := func(c echo.Context) error {
			return goboilerplate.ErrNotFound
		}

		buf := new(bytes.Buffer)
		log.SetOutput(buf)

		err := mw(h)(c).(*echo.HTTPError)

		require.Error(t, err)
		require.Equal(t, http.StatusNotFound, err.Code)
		require.Contains(t, buf.String(), goboilerplate.ErrNotFound.Error())
	})

	t.Run("with constraint error", func(t *testing.T) {
		h := func(c echo.Context) error {
			return goboilerplate.ConstraintErrorf("this is a constraint error")
		}

		buf := new(bytes.Buffer)
		log.SetOutput(buf)

		err := mw(h)(c).(*echo.HTTPError)
		require.Error(t, err)
		require.Equal(t, http.StatusBadRequest, err.Code)
		require.Contains(t, buf.String(), "this is a constraint error")
	})

	t.Run("with unknown error", func(t *testing.T) {
		h := func(c echo.Context) error {
			return errors.New("unexpected error")
		}

		buf := new(bytes.Buffer)
		log.SetOutput(buf)

		err := mw(h)(c).(*echo.HTTPError)
		require.Error(t, err)
		require.Equal(t, http.StatusInternalServerError, err.Code)
		require.Contains(t, buf.String(), "unexpected error")
	})

	t.Run("with context timeout", func(t *testing.T) {
		h := func(c echo.Context) error {
			ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Millisecond)
			defer cancel()

			slowOperation := func(ctx context.Context) error {
				select {
				case <-time.After(100 * time.Millisecond):
					return nil
				case <-ctx.Done():
					err := ctx.Err()
					return err
				}
			}

			err := slowOperation(ctx)

			return err
		}

		buf := new(bytes.Buffer)
		log.SetOutput(buf)

		err := mw(h)(c).(*echo.HTTPError)
		require.Error(t, err)
		require.Equal(t, http.StatusRequestTimeout, err.Code)
		require.Contains(t, buf.String(), context.DeadlineExceeded.Error())
	})
}
