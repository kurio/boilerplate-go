package http

import (
	"net/http"

	"github.com/labstack/echo"
)

func NewServer() *echo.Echo {
	e := echo.New()

	e.Use(errorMiddleware())

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	return e
}
