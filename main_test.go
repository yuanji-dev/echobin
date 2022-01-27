package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestGetHandler(t *testing.T) {
	e := echo.New()
	e.JSONSerializer = &echobinJSONSerializer{}
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
		if assert.NoError(t, getHandler(c)) {
			assert.Equal(t, http.StatusOK, res.Code)
			assert.Equal(t, tt.expectedJSON+"\n", res.Body.String())
		}
	}
}
