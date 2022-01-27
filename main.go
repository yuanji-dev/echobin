package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.JSONSerializer = &echobinJSONSerializer{}

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/get", getHandler)

	e.Logger.Fatal(e.Start("127.0.0.1:1323"))
}
