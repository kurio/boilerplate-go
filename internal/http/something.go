package http

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	goboilerplate "github.com/kurio/boilerplate-go"
)

func AddSomeHandler(e *echo.Echo, cacher goboilerplate.Cacher) {
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
		articleID := c.Param("id")

		var err error
		articleStr, err := cacher.Get(c.Request().Context(), articleID)
		if err == nil {
			var article map[string]interface{}
			err := json.Unmarshal([]byte(articleStr), &article)
			if err == nil {
				return c.JSON(http.StatusOK, article)
			}
		}

		if errors.Cause(err) != goboilerplate.ErrNotFound {
			logrus.Warnf("Error getting article '%s' from cache: %+v", articleID, err)
		}

		time.Sleep(time.Duration(rand.Intn(200)+100) * time.Millisecond) // simulate getting data from storage
		article := map[string]interface{}{"id": articleID}

		articleBytes, err := json.Marshal(article)
		if err != nil {
			logrus.Warnf("Error marshalling article to JSON: %+v", err)
			return c.JSON(http.StatusOK, article)
		}

		err = cacher.Set(c.Request().Context(), articleID, string(articleBytes), goboilerplate.DurationShort)
		if err != nil {
			logrus.Warnf("Error setting article '%s' to cache: %+v", articleID, err)
		}

		return c.JSON(http.StatusOK, article)
	}).Name = "getArticle"
}
