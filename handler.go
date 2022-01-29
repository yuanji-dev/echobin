package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func getHandler(c echo.Context) error {
	return c.JSONPretty(http.StatusOK, &getResponse{
		Args:    getArgs(c),
		Headers: getHeaders(c),
		Origin:  getOrigin(c),
		URL:     getURL(c),
	}, "  ")
}

func postHandler(c echo.Context) error {
	data := ""
	files := getFiles(c)
	form := getForm(c)
	if len(files) == 0 && len(form) == 0 {
		data = getData(c)
	}
	return c.JSONPretty(http.StatusOK, &postResponse{
		Args:    getArgs(c),
		Data:    data,
		Files:   files,
		Form:    form,
		Headers: getHeaders(c),
		JSON:    getJSON(c),
		Origin:  getOrigin(c),
		URL:     getURL(c),
	}, "  ")
}
