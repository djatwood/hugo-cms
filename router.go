package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func server() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.CORS())

	e.GET("/", listSites)
	e.GET("/:site", getSite)
	e.GET("/:site/:section", getSection)
	e.GET("/:site/:section/*", getFile)

	return e
}
