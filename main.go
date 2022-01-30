package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func newEcho() (e *echo.Echo) {
	e = echo.New()
	e.JSONSerializer = &echobinJSONSerializer{}
	return
}

func main() {
	e := newEcho()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/get", getMethodHandler)
	e.POST("/post", otherMethodHandler)
	e.PUT("/put", otherMethodHandler)
	e.PATCH("/patch", otherMethodHandler)
	e.DELETE("/delete", otherMethodHandler)

	e.Logger.Fatal(e.Start("127.0.0.1:1323"))
}
