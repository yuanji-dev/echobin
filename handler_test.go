package main

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
		assert.Contains(t, res.Body.String(), "<html")
		assert.Equal(t, echo.MIMETextHTMLCharsetUTF8, res.Header().Get(echo.HeaderContentType))
	}
}

func TestServeXMLHandler(t *testing.T) {
	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	if assert.NoError(t, serveXMLHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Contains(t, res.Body.String(), "<?xml")
		assert.Equal(t, echo.MIMEApplicationXMLCharsetUTF8, res.Header().Get(echo.HeaderContentType))
	}
}

func TestServeJSONHandler(t *testing.T) {
	e := newEcho()

	jsonFile, _ := os.Open("static/sample.json")
	defer jsonFile.Close()
	expectedJSON, _ := io.ReadAll(jsonFile)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	if assert.NoError(t, serveJSONHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, expectedJSON, res.Body.Bytes())
		assert.Equal(t, echo.MIMEApplicationJSONCharsetUTF8, res.Header().Get(echo.HeaderContentType))
	}
}

func TestServeRobotsTXTHandler(t *testing.T) {
	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	if assert.NoError(t, serveRobotsTXTHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, ROBOTS_TXT, res.Body.String())
		assert.Equal(t, echo.MIMETextPlainCharsetUTF8, res.Header().Get(echo.HeaderContentType))
	}
}

func TestServeDenyHandler(t *testing.T) {
	e := newEcho()

	denyFile, _ := os.Open("static/deny.txt")
	defer denyFile.Close()
	expectedTXT, _ := io.ReadAll(denyFile)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	if assert.NoError(t, serveDenyHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, expectedTXT, res.Body.Bytes())
		assert.Equal(t, echo.MIMETextPlainCharsetUTF8, res.Header().Get(echo.HeaderContentType))
	}
}

func TestServeUTF8HTMLHandler(t *testing.T) {
	e := newEcho()

	txtFile, _ := os.Open("static/sample-utf8.html")
	defer txtFile.Close()
	expectedTXT, _ := io.ReadAll(txtFile)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	if assert.NoError(t, serveUTF8HTMLHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, expectedTXT, res.Body.Bytes())
		assert.Equal(t, echo.MIMETextHTMLCharsetUTF8, res.Header().Get(echo.HeaderContentType))
	}
}

func TestServeGzipHandler(t *testing.T) {
	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, "gzip")
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	h := middleware.Gzip()(serveGzipHandler)
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, "gzip", res.Header().Get(echo.HeaderContentEncoding))
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	res = httptest.NewRecorder()
	c = e.NewContext(req, res)
	h = middleware.Gzip()(serveGzipHandler)
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Empty(t, res.Header().Get(echo.HeaderContentEncoding))
		assert.Contains(t, res.Body.String(), `"gzipped": false`)
	}
}

func TestServeDeflateHandler(t *testing.T) {
	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, "deflate")
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	h := middleware.Deflate()(serveDeflateHandler)
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, "deflate", res.Header().Get(echo.HeaderContentEncoding))
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	res = httptest.NewRecorder()
	c = e.NewContext(req, res)
	h = middleware.Deflate()(serveDeflateHandler)
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Empty(t, res.Header().Get(echo.HeaderContentEncoding))
		assert.Contains(t, res.Body.String(), `"deflated": false`)
	}
}

func TestServeBrotliHandler(t *testing.T) {
	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, "br")
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	if assert.NoError(t, serveBrotliHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, "br", res.Header().Get(echo.HeaderContentEncoding))
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	res = httptest.NewRecorder()
	c = e.NewContext(req, res)
	if assert.NoError(t, serveBrotliHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Contains(t, res.Body.String(), `"brotli": false`)
	}
}

func TestBase64Handler(t *testing.T) {
	e := newEcho()

	cases := []struct {
		input  string
		output string
	}{
		{"RUNIT0JJTiBpcyBhd2Vzb21l", "ECHOBIN is awesome"},
		// Support both encodings: StdEncoding & URLEncoding
		// See also: https://gobyexample.com/base64-encoding
		{"YWJjMTIzIT8kKiYoKSctPUB+", "abc123!?$*&()'-=@~"},
		{"YWJjMTIzIT8kKiYoKSctPUB-", "abc123!?$*&()'-=@~"},
		// Test urlencoded string and whitespace
		{"YWJjMTIzIT8kKiYoKSctPUB%2B%20%20", "abc123!?$*&()'-=@~"},
		{"  RUNIT0JJTiBpcyBhd2Vzb21l  ", "ECHOBIN is awesome"},
	}
	for _, v := range cases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		c.SetPath("/base64/:value")
		c.SetParamNames("value")
		c.SetParamValues(v.input)
		if assert.NoError(t, base64Handler(c)) {
			assert.Equal(t, http.StatusOK, res.Code)
			assert.Equal(t, v.output, res.Body.String())
		}

	}
}

func TestGenerateBytesHandler(t *testing.T) {
	e := newEcho()

	cases := []struct {
		n          string
		statusCode int
	}{
		{"0", http.StatusOK},
		{"10", http.StatusOK},
		{"102401", http.StatusOK},
		{"-1", http.StatusBadRequest},
		{"e", http.StatusBadRequest},
	}
	for _, v := range cases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		c.SetPath("/bytes/:n")
		c.SetParamNames("n")
		c.SetParamValues(v.n)

		// if handler return an error, it won't write to res automatically
		// see also: https://github.com/labstack/echo/issues/593#issuecomment-230926351
		err := generateBytesHandler(c)
		if err != nil {
			he, _ := err.(*echo.HTTPError)
			assert.Equal(t, v.statusCode, he.Code)
		} else {
			assert.Equal(t, v.statusCode, res.Code)
			assert.Equal(t, echo.MIMEOctetStream, res.Header().Get(echo.HeaderContentType))
			// TODO: seems Content-Length header is not set in testing
			//assert.Equal(t, "100", res.Header().Get(echo.HeaderContentLength))
		}
	}
}

// TODO: add test for /delay/:delay endpoint
// func TestDelayHandler(t *testing.T)

// TODO: add test for /drip endpoint
// func TestDripHandler(t *testing.T)

func TestLinksHandler(t *testing.T) {
	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	c.SetPath("/links/:n/:offset")
	c.SetParamNames("n", "offset")
	c.SetParamValues("100", "50")

	if assert.NoError(t, linksHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Contains(t, res.Body.String(), "/links/100/99")
	}
}

func TestStreamHandler(t *testing.T) {
	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	c.SetPath("/stream/:n")
	n := 20
	c.SetParamNames("n")
	c.SetParamValues(fmt.Sprintf("%d", n))

	if assert.NoError(t, streamHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		actual := 0
		// https://pkg.go.dev/encoding/json#Decoder
		dec := json.NewDecoder(res.Body)
		for ; ; actual++ {
			var sr streamResponse
			if err := dec.Decode(&sr); err == io.EOF {
				break
			} else if err != nil {
				assert.Error(t, err)
			}
		}
		assert.Equal(t, n, actual)
	}
}

func TestUUIDHandler(t *testing.T) {
	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)

	if assert.NoError(t, UUIDHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		var ur UUIDResponse
		assert.Empty(t, ur)
		if err := json.Unmarshal(res.Body.Bytes(), &ur); err != nil {
			assert.Error(t, err)
		}
		assert.NotEmpty(t, ur)
	}
}

func TestImageHandler(t *testing.T) {
	e := newEcho()

	cases := []string{
		"image/webp",
		"image/svg+xml",
		"image/jpeg",
		"image/png",
		"image/*",
	}
	for _, v := range cases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderAccept, v)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		if assert.NoError(t, imageHandler(c)) {
			assert.Equal(t, http.StatusOK, res.Code)
			if v == "image/*" {
				assert.Equal(t, "image/png", res.Header().Get(echo.HeaderContentType))
			} else {
				assert.Equal(t, v, res.Header().Get(echo.HeaderContentType))
			}
		}
	}

	// case for 406 status code
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	err := imageHandler(c)
	assert.Error(t, err)
	he, _ := err.(*echo.HTTPError)
	assert.Equal(t, http.StatusNotAcceptable, he.Code)
}

func TestGetCookiesHandler(t *testing.T) {
	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	cookie := &http.Cookie{
		Name:  "hello",
		Value: "world",
		Path:  "/",
	}
	expectedJSON := `{
  "cookies": {
    "hello": "world"
  }
}`
	req.AddCookie(cookie)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	if assert.NoError(t, getCookiesHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.JSONEq(t, expectedJSON, res.Body.String())
	}
}

func TestSetCookiesInQueryHandler(t *testing.T) {
	e := newEcho()

	q := make(url.Values)
	q.Set("hello", "world")
	q.Set("konnichiwa", "sekai")
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	if assert.NoError(t, setCookiesInQueryHandler(c)) {
		assert.Equal(t, http.StatusFound, res.Code)
		assert.Len(t, res.Result().Cookies(), 2)
	}
}

func TestSetCookiesInPathHandler(t *testing.T) {
	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	c.SetPath("/cookies/set/:name/:value")
	c.SetParamNames("name", "value")
	c.SetParamValues("hello", "world")
	if assert.NoError(t, setCookiesInPathHandler(c)) {
		assert.Equal(t, http.StatusFound, res.Code)
		assert.Equal(t, "hello", res.Result().Cookies()[0].Name)
		assert.Equal(t, "world", res.Result().Cookies()[0].Value)
	}
}

func TestDeleteCookiesHandler(t *testing.T) {
	e := newEcho()

	q := make(url.Values)
	q.Set("hello", "")
	q.Set("konnichiwa", "")
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.AddCookie(&http.Cookie{
		Name:  "hello",
		Value: "world",
	})
	req.AddCookie(&http.Cookie{
		Name:  "konnichiwa",
		Value: "sekai",
	})
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	if assert.NoError(t, deleteCookiesHandler(c)) {
		assert.Equal(t, http.StatusFound, res.Code)
		assert.Contains(t, res.Header().Values(echo.HeaderSetCookie), "hello=world; Path=/; Max-Age=0")
		assert.Contains(t, res.Header().Values(echo.HeaderSetCookie), "konnichiwa=sekai; Path=/; Max-Age=0")
	}
}

func TestGetRedirectToHandler(t *testing.T) {
	e := newEcho()

	cases := []struct {
		url          string
		statusCode   string
		expectedCode int
	}{
		{"http://example.com", "303", 303},
		{"http://example.com", "400", 302},
		{"http://example.com", "", 302},
		{"/", "301", 301},
	}
	for _, v := range cases {
		q := make(url.Values)
		q.Set("url", v.url)
		q.Set("status_code", v.statusCode)

		req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		if assert.NoError(t, getRedirectToHandler(c)) {
			assert.Equal(t, v.expectedCode, res.Code)
			assert.Equal(t, v.url, res.Header().Get(echo.HeaderLocation))
		}
	}
}

func TestOtherRedirectToHandler(t *testing.T) {
	e := newEcho()

	cases := []struct {
		url          string
		statusCode   string
		expectedCode int
	}{
		{"http://example.com", "303", 303},
		{"http://example.com", "400", 302},
		{"http://example.com", "", 302},
		{"/", "301", 301},
	}
	for _, v := range cases {
		f := make(url.Values)
		f.Set("url", v.url)
		f.Set("status_code", v.statusCode)

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		if assert.NoError(t, otherRedirectToHandler(c)) {
			assert.Equal(t, v.expectedCode, res.Code)
			assert.Equal(t, v.url, res.Header().Get(echo.HeaderLocation))
		}
	}
}

func TestStaticRedirectHandler(t *testing.T) {
	e := newEcho()

	cases := []struct {
		n        string
		handler  echo.HandlerFunc
		location string
	}{
		{"100", absoluteRedirectHandler, "http://example.com/absolute-redirect/99"},
		{"1", absoluteRedirectHandler, "http://example.com/get"},
		{"100", relativeRedirectHandler, "/relative-redirect/99"},
		{"1", relativeRedirectHandler, "/get"},
	}
	for _, v := range cases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		c.SetParamNames("n")
		c.SetParamValues(v.n)
		if assert.NoError(t, v.handler(c)) {
			assert.Equal(t, http.StatusFound, res.Code)
			assert.Equal(t, v.location, res.Header().Get(echo.HeaderLocation))
		}
	}
}

func TestDynamicRedirectHandler(t *testing.T) {
	e := newEcho()

	cases := []struct {
		n        string
		absolute bool
		location string
	}{
		{"100", true, "http://example.com/absolute-redirect/99"},
		{"1", true, "http://example.com/get"},
		{"100", false, "/relative-redirect/99"},
		{"1", false, "/get"},
	}
	for _, v := range cases {
		target := "/"
		if v.absolute {
			q := make(url.Values)
			q.Set("absolute", "true")
			target = "/?" + q.Encode()
		}
		req := httptest.NewRequest(http.MethodGet, target, nil)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		c.SetParamNames("n")
		c.SetParamValues(v.n)
		if assert.NoError(t, redirectHandler(c)) {
			assert.Equal(t, http.StatusFound, res.Code)
			assert.Equal(t, v.location, res.Header().Get(echo.HeaderLocation))
		}
	}
}

func TestAnythingHandler(t *testing.T) {
	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	if assert.NoError(t, anythingHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
	}
}

func TestCacheHandler(t *testing.T) {
	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	if assert.NoError(t, cacheHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, echo.MIMEApplicationJSONCharsetUTF8, res.Header().Get(echo.HeaderContentType))
		assert.Contains(t, res.Header(), echo.HeaderLastModified)
		assert.Contains(t, res.Header(), "Etag")
	}

	cachedCases := []struct {
		header string
		value  string
	}{
		{echo.HeaderIfModifiedSince, time.Now().UTC().Format(http.TimeFormat)},
		{"If-None-Match", "772867218dd444f6b15f1d9eb67f74bd"},
	}
	for _, v := range cachedCases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(v.header, v.value)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		if assert.NoError(t, cacheHandler(c)) {
			assert.Equal(t, http.StatusNotModified, res.Code)
		}
	}
}

func TestCacheDurationHandler(t *testing.T) {
	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	c.SetParamNames("value")
	c.SetParamValues("100")
	if assert.NoError(t, cacheDurationHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, echo.MIMEApplicationJSONCharsetUTF8, res.Header().Get(echo.HeaderContentType))
		assert.Equal(t, "public, max-age=100", res.Header().Get("Cache-Control"))
	}
}

func TestEtagHandler(t *testing.T) {
	e := newEcho()

	// No headers
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	c.SetParamNames("etag")
	c.SetParamValues("abcdef")
	if assert.NoError(t, etagHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, "abcdef", res.Header().Get("ETag"))
	}

	// Test If-None-Match header
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("If-None-Match", "abcdef") // Match, return 304
	res = httptest.NewRecorder()
	c = e.NewContext(req, res)
	c.SetParamNames("etag")
	c.SetParamValues("abcdef")
	if assert.NoError(t, etagHandler(c)) {
		assert.Equal(t, http.StatusNotModified, res.Code)
		assert.Equal(t, "abcdef", res.Header().Get("ETag"))
	}
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("If-None-Match", "fedcba") // Not match, return normal response
	res = httptest.NewRecorder()
	c = e.NewContext(req, res)
	c.SetParamNames("etag")
	c.SetParamValues("abcdef")
	if assert.NoError(t, etagHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, "abcdef", res.Header().Get("ETag"))
	}

	// Test If-Match header
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("If-Match", "abcdef") // Match, return normal response
	res = httptest.NewRecorder()
	c = e.NewContext(req, res)
	c.SetParamNames("etag")
	c.SetParamValues("abcdef")
	if assert.NoError(t, etagHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, "abcdef", res.Header().Get("ETag"))
	}
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("If-Match", "fedcba") // Not match, return 412
	res = httptest.NewRecorder()
	c = e.NewContext(req, res)
	c.SetParamNames("etag")
	c.SetParamValues("abcdef")
	if assert.NoError(t, etagHandler(c)) {
		assert.Equal(t, http.StatusPreconditionFailed, res.Code)
	}
}

func TestResponseHeadersHandler(t *testing.T) {
	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/?a=1&a=2&b=3", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	if assert.NoError(t, responseHeadersHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, echo.MIMEApplicationJSONCharsetUTF8, res.Header().Get(echo.HeaderContentType))
		assert.Equal(t, []string{"1", "2"}, res.Header()["a"])
		assert.Equal(t, []string{"3"}, res.Header()["b"])
		// TODO: Content-Length in body and header should be equal,
		// but atm idk why Content-Length wasn't set in header.
		// var i interface{}
		// json.Unmarshal(res.Body.Bytes(), &i)
		// data := i.(map[string]interface{})
		// assert.Equal(t, data["Content-Length"], res.Result().ContentLength)
	}
}

func TestBasicAuthHandler(t *testing.T) {
	e := newEcho()

	// Test Unauthorized
	req := httptest.NewRequest(http.MethodGet, "/basic-auth/a/b", nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Contains(t, res.Header().Get(echo.HeaderWWWAuthenticate), "basic")

	// Test Authorized
	req = httptest.NewRequest(http.MethodGet, "/basic-auth/a/b", nil)
	req.SetBasicAuth("a", "b")
	res = httptest.NewRecorder()
	e.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, echo.MIMEApplicationJSONCharsetUTF8, res.Header().Get(echo.HeaderContentType))

	// Test Bad Credentials
	req = httptest.NewRequest(http.MethodGet, "/basic-auth/a/b", nil)
	req.SetBasicAuth("a", "a")
	res = httptest.NewRecorder()
	e.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Contains(t, res.Header().Get(echo.HeaderWWWAuthenticate), "basic")
}

func TestBearerHandler(t *testing.T) {
	e := newEcho()

	// Test Unauthorized
	req := httptest.NewRequest(http.MethodGet, "/bearer", nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Contains(t, res.Header().Get(echo.HeaderWWWAuthenticate), "Bearer")

	// Test Authorized
	req = httptest.NewRequest(http.MethodGet, "/bearer", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer abc")
	res = httptest.NewRecorder()
	e.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, echo.MIMEApplicationJSONCharsetUTF8, res.Header().Get(echo.HeaderContentType))

	// Test Bad Credentials
	req = httptest.NewRequest(http.MethodGet, "/bearer", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer ")
	res = httptest.NewRecorder()
	e.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Contains(t, res.Header().Get(echo.HeaderWWWAuthenticate), "Bearer")
}

func TestStreamBytesHandler(t *testing.T) {
	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/stream-bytes/100?seed=321", nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, echo.MIMEOctetStream, res.Header().Get(echo.HeaderContentType))

	req1 := httptest.NewRequest(http.MethodGet, "/stream-bytes/100?seed=123", nil)
	res1 := httptest.NewRecorder()
	e.ServeHTTP(res1, req1)
	assert.Equal(t, http.StatusOK, res1.Code)
	assert.Equal(t, echo.MIMEOctetStream, res1.Header().Get(echo.HeaderContentType))

	req2 := httptest.NewRequest(http.MethodGet, "/stream-bytes/100?seed=123", nil)
	res2 := httptest.NewRecorder()
	e.ServeHTTP(res2, req2)
	assert.Equal(t, http.StatusOK, res2.Code)
	assert.Equal(t, echo.MIMEOctetStream, res2.Header().Get(echo.HeaderContentType))

	assert.Equal(t, res1.Body.Bytes(), res2.Body.Bytes())
	assert.NotEqual(t, res1.Body.Bytes(), res.Body.Bytes())
}

func TestRangeHandler(t *testing.T) {
	e := newEcho()

	// numBytes is over maxByteCount
	req := httptest.NewRequest(http.MethodGet, "/range/102401", nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	assert.Equal(t, http.StatusNotFound, res.Code)

	badHeaders := []string{
		"bytes=100-1",     // start is greater than end
		"bytes=100-10000", // end is greater than maximum
	}
	for _, bh := range badHeaders {
		req := httptest.NewRequest(http.MethodGet, "/range/10000", nil)
		req.Header.Set("Range", bh)
		res := httptest.NewRecorder()
		e.ServeHTTP(res, req)
		assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, res.Code)
	}

	reqFull := httptest.NewRequest(http.MethodGet, "/range/10000", nil)
	reqFull.Header.Set("Range", "bytes=0-9999")
	resFull := httptest.NewRecorder()
	e.ServeHTTP(resFull, reqFull)
	assert.Equal(t, http.StatusOK, resFull.Code)
	assert.Equal(t, echo.MIMEOctetStream, resFull.Header().Get(echo.HeaderContentType))
	assert.EqualValues(t, 10000, resFull.Result().ContentLength)
	assert.Equal(t, "bytes 0-9999/10000", resFull.Result().Header.Get("Content-Range"))
	assert.Len(t, resFull.Body.Bytes(), 10000)

	partialCases := []struct {
		start int
		end   int
		chunk int
	}{
		{500, 1999, 1},
		{500, 1999, 1000},
		{500, 1999, 10000},
	}
	for _, v := range partialCases {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/range/10000?chunk_size=%d", v.chunk), nil)
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", v.start, v.end))
		res := httptest.NewRecorder()
		e.ServeHTTP(res, req)
		assert.Equal(t, http.StatusPartialContent, res.Code)
		assert.Equal(t, echo.MIMEOctetStream, res.Header().Get(echo.HeaderContentType))
		assert.EqualValues(t, v.end-v.start+1, res.Result().ContentLength)
		assert.Equal(t, fmt.Sprintf("bytes %d-%d/10000", v.start, v.end), res.Result().Header.Get("Content-Range"))
		assert.Len(t, res.Body.Bytes(), v.end-v.start+1)
	}
}

func TestFormHandler(t *testing.T) {
	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	if assert.NoError(t, formHandler(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Contains(t, res.Body.String(), "Pizza")
		assert.Equal(t, echo.MIMETextHTMLCharsetUTF8, res.Header().Get(echo.HeaderContentType))
	}
}
