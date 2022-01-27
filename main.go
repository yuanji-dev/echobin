package main

import (
	"net/http"

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

type responseRef struct {
	URL     string            `json:"url"`
	Args    interface{}       `json:"args"`
	Form    interface{}       `json:"form"`
	Data    interface{}       `json:"data"`
	Origin  string            `json:"origin"`
	Headers map[string]string `json:"headers"`
	Files   interface{}       `json:"files"`
	JSON    interface{}       `json:"json"`
	Method  string            `json:"method"`
}

type getResponse struct {
	Args    map[string]interface{} `json:"args"`
	Headers map[string]string      `json:"headers"`
	Origin  string                 `json:"origin"`
	URL     string                 `json:"url"`
}

func getURL(c echo.Context) string {
	r := c.Request()
	return c.Scheme() + "://" + r.Host + r.URL.RequestURI()
}

func getOrigin(c echo.Context) string {
	return c.RealIP()
}

func getHeaders(c echo.Context) map[string]string {
	headers := map[string]string{}
	for k, v := range c.Request().Header {
		headers[k] = v[0]
	}
	return headers
}

func getArgs(c echo.Context) map[string]interface{} {
	args := map[string]interface{}{}
	for k, v := range c.QueryParams() {
		if len(v) == 1 {
			args[k] = v[0]
		} else {
			args[k] = v
		}
	}
	return args
}

func getHandler(c echo.Context) error {
	return c.JSONPretty(http.StatusOK, &getResponse{
		URL:     getURL(c),
		Args:    getArgs(c),
		Headers: getHeaders(c),
		Origin:  getOrigin(c),
	}, "  ")
}
