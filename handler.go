package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func getMethodHandler(c echo.Context) error {
	return c.JSONPretty(http.StatusOK, &getMethodResponse{
		Args:    getArgs(c),
		Headers: getHeaders(c),
		Origin:  getOrigin(c),
		URL:     getURL(c),
	}, "  ")
}

func otherMethodHandler(c echo.Context) error {
	data := ""
	files := getFiles(c)
	form := getForm(c)
	if len(files) == 0 && len(form) == 0 {
		data = getData(c)
	}
	res := otherMethodResponse{}
	res.Args = getArgs(c)
	res.Data = data
	res.Files = files
	res.Form = form
	res.Headers = getHeaders(c)
	res.JSON = getJSON(c)
	res.Origin = getOrigin(c)
	res.URL = getURL(c)
	return c.JSONPretty(http.StatusOK, &res, "  ")
}
