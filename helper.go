package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/url"

	"github.com/labstack/echo/v4"
)

func getURL(c echo.Context) string {
	r := c.Request()
	fullURL := c.Scheme() + "://" + r.Host + r.URL.Path
	query, err := url.QueryUnescape(r.URL.RawQuery)
	if err == nil && query != "" {
		fullURL += "?" + query
	}
	return fullURL
}

func getOrigin(c echo.Context) string {
	return c.RealIP()
}

func getUserAgent(c echo.Context) string {
	return c.Request().UserAgent()
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
	c.Request().ParseForm()
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

func getFiles(c echo.Context) map[string]interface{} {
	files := map[string]interface{}{}
	multipartForm, _ := c.MultipartForm()
	if multipartForm == nil {
		return map[string]interface{}{}
	}
	for k, v := range multipartForm.File {
		var contents []string
		for _, fh := range v {
			f, _ := fh.Open()
			defer f.Close()
			content, _ := ioutil.ReadAll(f)
			contents = append(contents, string(content))
		}
		if len(contents) == 1 {
			files[k] = contents[0]
		} else {
			// Seems original httpbin doesn't support upload multiple files
			// sharing same field name, but implemented here.
			files[k] = contents
		}
	}
	return files
}

func getCookies(c echo.Context) map[string]string {
	cookies := map[string]string{}
	for _, c := range c.Cookies() {
		cookies[c.Name] = c.Value
	}
	return cookies
}
