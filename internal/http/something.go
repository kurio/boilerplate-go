package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func AddSomeHandler(e *echo.Echo) {
	g := e.Group("/something")

	g.GET("/:duration", func(c echo.Context) error {
		sleepTime, err := strconv.ParseInt(c.Param("duration"), 10, 64)
		if err != nil {
			logrus.Errorf("Error parsing duration: %+v", err)
			sleepTime = 1
		}
		logrus.Debugf("Sleep for %d ms", sleepTime)
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)

		logrus.Debugf("Returning...")
		return c.String(http.StatusOK, "ok")
	}).Name = "getSomething"

	e.GET("/articles", func(c echo.Context) error {
		return c.JSON(http.StatusOK, make([]interface{}, 0))
	}).Name = "fetchArticles"

	e.GET("/articles/:id", func(c echo.Context) error {
		return c.String(http.StatusOK, c.Param("id"))
	}).Name = "getArticle"

}
