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
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	h := middleware.Gzip()(serveGzipHandler)
	h = forceEncode(h, "gzip")
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, "gzip", res.Header().Get(echo.HeaderContentEncoding))
	}
}

func TestServeDeflateHandler(t *testing.T) {
	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	h := middleware.Deflate()(serveGzipHandler)
	h = forceEncode(h, "deflate")
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, "deflate", res.Header().Get(echo.HeaderContentEncoding))
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
