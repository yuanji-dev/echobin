package main

import (
	// "io/ioutil"

	"bytes"
	"encoding/json"
	"io/ioutil"

	"github.com/labstack/echo/v4"
)

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

func getData(c echo.Context) string {
	reqBody := []byte{}
	if c.Request().Body != nil {
		reqBody, _ = ioutil.ReadAll(c.Request().Body)
	}
	// https://github.com/labstack/echo/blob/8da8e161380fd926d4341721f0328f1e94d6d0a2/middleware/body_dump.go#L73
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))
	return string(reqBody)
}

func getForm(c echo.Context) map[string]interface{} {
	form := map[string]interface{}{}
	for k, v := range c.Request().PostForm {
		if len(v) == 1 {
			form[k] = v[0]
		} else {
			form[k] = v
		}
	}
	return form
}

func getJSON(c echo.Context) map[string]interface{} {
	var i map[string]interface{}
	json.Unmarshal([]byte(getData(c)), &i)
	return i
}
