package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/swaggo/echo-swagger"

	_ "github.com/masakichi/echobin/docs"
)

func newEcho() (e *echo.Echo) {
	e = echo.New()
	e.JSONSerializer = &echobinJSONSerializer{}
	return
}

// @title        echobin API
// @version      0.1
// @description  A simple HTTP Request & Response Service.

// @contact.name   Yuanji
// @contact.url    https://gimo.me
// @contact.email  self@gimo.me

// @license.name  MIT License
// @license.url   https://github.com/masakichi/echobin/blob/main/LICENSE

// @tag.name         HTTP methods
// @tag.description  Testing different HTTP verbs
// @tag.name         Request inspection
// @tag.description  Inspect the request data
// @tag.name         Status codes
// @tag.description  Generates responses with given status code
func main() {
	e := newEcho()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Swagger docs
	e.GET("/*", echoSwagger.WrapHandler)

	// HTTP methods
	e.GET("/get", getMethodHandler)
	e.POST("/post", otherMethodHandler)
	e.PUT("/put", otherMethodHandler)
	e.PATCH("/patch", otherMethodHandler)
	e.DELETE("/delete", otherMethodHandler)

	// Request inspection
	e.GET("/ip", requestIPHandler)

	// Status Codes
	e.Any("/status/:codes", statusCodesHandler)

	e.Logger.Fatal(e.Start("127.0.0.1:1323"))
}
