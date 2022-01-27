package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func getHandler(c echo.Context) error {
	return c.JSONPretty(http.StatusOK, &getResponse{
		URL:     getURL(c),
		Args:    getArgs(c),
		Headers: getHeaders(c),
		Origin:  getOrigin(c),
	}, "  ")
}
