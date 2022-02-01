package main

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestGetHandler(t *testing.T) {
	e := newEcho()
	tests := []struct {
		target       string
		expectedJSON string
	}{
		{"/get?q=1", `{
  "args": {
    "q": "1"
  },
  "headers": {
    "User-Agent": "fake-agent"
  },
  "origin": "192.0.2.1",
  "url": "http://example.com/get?q=1"
}`},
		{"/get?q=1&q=2", `{
  "args": {
    "q": [
      "1",
      "2"
    ]
  },
  "headers": {
    "User-Agent": "fake-agent"
  },
  "origin": "192.0.2.1",
  "url": "http://example.com/get?q=1&q=2"
}`},
	}
	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, tt.target, nil)
		req.Header.Set("user-agent", "fake-agent")
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		if assert.NoError(t, getMethodHandler(c)) {
			assert.Equal(t, http.StatusOK, res.Code)
			assert.Equal(t, tt.expectedJSON+"\n", res.Body.String())
		}
	}
}

func TestOtherHandlerWithJSON(t *testing.T) {
	e := newEcho()

	reqJSON := `{"name":"Bob"}`
	req := httptest.NewRequest(http.MethodPost, "/post?q=1&q=2", strings.NewReader(reqJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("user-agent", "fake-agent")
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	expected := `{
  "args": {
    "q": [
      "1",
      "2"
    ]
  },
  "data": "{\"name\":\"Bob\"}",
  "files": {},
  "form": {},
  "headers": {
    "Content-Type": "application/json",
    "User-Agent": "fake-agent"
  },
  "json": {
    "name": "Bob"
  },
  "origin": "192.0.2.1",
  "url": "http://example.com/post?q=1&q=2"
}`
	if assert.NoError(t, otherMethodHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, expected+"\n", res.Body.String())
	}
}

func TestOtherHandlerWithForm(t *testing.T) {
	e := newEcho()

	f := make(url.Values)
	f.Set("name", "Bob")
	req := httptest.NewRequest(http.MethodPost, "/post?q=1&q=2", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.Header.Set("user-agent", "fake-agent")
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	expected := `{
  "args": {
    "q": [
      "1",
      "2"
    ]
  },
  "data": "",
  "files": {},
  "form": {
    "name": "Bob"
  },
  "headers": {
    "Content-Type": "application/x-www-form-urlencoded",
    "User-Agent": "fake-agent"
  },
  "json": null,
  "origin": "192.0.2.1",
  "url": "http://example.com/post?q=1&q=2"
}`
	if assert.NoError(t, otherMethodHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, expected+"\n", res.Body.String())
	}
}

func TestOtherHandlerWithFiles(t *testing.T) {
	e := newEcho()
	pr, pw := io.Pipe()
	w := multipart.NewWriter(pw)
	go func() {
		defer pw.Close()
		defer w.Close()

		file, err := os.Open("handler.go")
		assert.NoError(t, err)

		formFile, err := w.CreateFormFile("file", "handler.go")
		assert.NoError(t, err)

		_, err = io.Copy(formFile, file)
		assert.NoError(t, err)

		assert.NoError(t, w.WriteField("name", "Bob"))
	}()

	req := httptest.NewRequest(http.MethodPost, "/post", pr)
	req.Header.Set(echo.HeaderContentType, w.FormDataContentType())
	req.Header.Set("user-agent", "fake-agent")
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	if assert.NoError(t, otherMethodHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		// TODO: Better to check structure and format
		assert.Contains(t, res.Body.String(), "multipart/form-data")
		assert.Contains(t, res.Body.String(), "Bob")
	}
}

func TestStatusCodesHandler(t *testing.T) {
	e := newEcho()

	validCases := []struct {
		codes    string
		expected []int
	}{
		{"200", []int{200}},
		{"200,500", []int{200, 500}},
		{"200%2C500", []int{200, 500}},
		{"200:0.3,500:0.7", []int{200, 500}},
		{"200:0.3,500:0.7,301", []int{200, 500, 301}},
	}
	for _, v := range validCases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		c.SetPath("/status/:codes")
		c.SetParamNames("codes")
		c.SetParamValues(v.codes)
		if assert.NoError(t, statusCodesHandler(c)) {
			assert.Contains(t, v.expected, res.Code)
		}
	}
}

func TestRequestIPHandler(t *testing.T) {
	e := newEcho()

	expected := `{
  "origin": "192.0.2.1"
}`
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	if assert.NoError(t, requestIPHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, expected+"\n", res.Body.String())
	}
}

func TestRequestHeadersHandler(t *testing.T) {
	e := newEcho()

	expected := fmt.Sprintf(`{
  "headers": {
    "%s": "%s"
  }
}`, echo.HeaderContentType, echo.MIMEApplicationJSON)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	if assert.NoError(t, requestHeadersHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, expected+"\n", res.Body.String())
	}
}

func TestRequestUserAgentHandler(t *testing.T) {
	e := newEcho()

	userAgent := "fake-agent"
	expected := fmt.Sprintf(`{
  "user-agent": "%s"
}`, userAgent)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("user-agent", userAgent)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	if assert.NoError(t, requestUserAgentHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, expected+"\n", res.Body.String())
	}
}

func TestServeHTMLHandler(t *testing.T) {
	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	if assert.NoError(t, serveHTMLHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Contains(t, res.Body.String(), "<html>")
	}
}
