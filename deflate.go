package main

import (
	"bufio"
	"compress/flate"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type (
	// DeflateConfig defines the config for Deflate middleware.
	DeflateConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper

		// Deflate compression level.
		// Optional. Default value -1.
		Level int `yaml:"level"`
	}

	deflateResponseWriter struct {
		io.Writer
		http.ResponseWriter
		wroteBody bool
	}
)

const (
	deflateScheme = "deflate"
)

var (
	// DefaultDeflateConfig is the default deflate middleware config.
	DefaultDeflateConfig = DeflateConfig{
		Skipper: middleware.DefaultSkipper,
		Level:   -1,
	}
)

// Deflate returns a middleware which compresses HTTP response using deflate compression
// scheme.
func Deflate() echo.MiddlewareFunc {
	return DeflateWithConfig(DefaultDeflateConfig)
}

// DeflateWithConfig return Deflate middleware with config.
func DeflateWithConfig(config DeflateConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultDeflateConfig.Skipper
	}
	if config.Level == 0 {
		config.Level = DefaultDeflateConfig.Level
	}

	pool := deflateCompressPool(config)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			res := c.Response()
			res.Header().Add(echo.HeaderVary, echo.HeaderAcceptEncoding)
			if strings.Contains(c.Request().Header.Get(echo.HeaderAcceptEncoding), deflateScheme) {
				res.Header().Set(echo.HeaderContentEncoding, deflateScheme) // Issue #806
				i := pool.Get()
				w, ok := i.(*flate.Writer)
				if !ok {
					return echo.NewHTTPError(http.StatusInternalServerError, i.(error).Error())
				}
				rw := res.Writer
				w.Reset(rw)
				grw := &deflateResponseWriter{Writer: w, ResponseWriter: rw}
				defer func() {
					if !grw.wroteBody {
						if res.Header().Get(echo.HeaderContentEncoding) == deflateScheme {
							res.Header().Del(echo.HeaderContentEncoding)
						}
						// We have to reset response to it's pristine state when
						// nothing is written to body or error is returned.
						// See issue #424, #407.
						res.Writer = rw
						w.Reset(ioutil.Discard)
					}
					w.Close()
					pool.Put(w)
				}()
				res.Writer = grw
			}
			return next(c)
		}
	}
}

func (w *deflateResponseWriter) WriteHeader(code int) {
	w.Header().Del(echo.HeaderContentLength) // Issue #444
	w.ResponseWriter.WriteHeader(code)
}

func (w *deflateResponseWriter) Write(b []byte) (int, error) {
	if w.Header().Get(echo.HeaderContentType) == "" {
		w.Header().Set(echo.HeaderContentType, http.DetectContentType(b))
	}
	w.wroteBody = true
	return w.Writer.Write(b)
}

func (w *deflateResponseWriter) Flush() {
	w.Writer.(*flate.Writer).Flush()
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *deflateResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *deflateResponseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := w.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}

func deflateCompressPool(config DeflateConfig) sync.Pool {
	return sync.Pool{
		New: func() interface{} {
			w, err := flate.NewWriter(ioutil.Discard, config.Level)
			if err != nil {
				return err
			}
			return w
		},
	}
}
