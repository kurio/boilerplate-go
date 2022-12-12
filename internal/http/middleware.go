package http

import (
	"context"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
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

// URLSkipper skip unwanted URL from scrapped by Prometheus
func URLSkipper(c echo.Context) bool {
	switch c.Path() {
	case "/ping", "/_version", "/metrics":
		return true
	}
	return strings.HasPrefix(c.Path(), "/debug")
}
