package main

import "github.com/labstack/echo/v4"

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
