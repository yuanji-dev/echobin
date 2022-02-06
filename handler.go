package main

import (
	"bytes"
	_ "embed"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

const maxByteCount = 100 << 10
const maxDelay = 10 // seconds

// @Summary  The request's query parameters.
// @Tags     HTTP methods
// @Produce  json
// @Success  200  {object}  getMethodResponse
// @Router   /get [get]
func getMethodHandler(c echo.Context) error {
	return c.JSONPretty(http.StatusOK, &getMethodResponse{
		Args:    getArgs(c),
		Headers: getHeaders(c),
		Origin:  getOrigin(c),
		URL:     getURL(c),
	}, "  ")
}

// @Summary  The request's query parameters.
// @Tags     HTTP methods
// @Accept   json
// @Accept   mpfd
// @Accept   x-www-form-urlencoded
// @Produce  json
// @Success  200  {object}  otherMethodResponse
// @Router   /post [post]
// @Router   /put [put]
// @Router   /patch [patch]
// @Router   /delete [delete]
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

// @Summary  Returns the requester's IP Address.
// @Tags     Request inspection
// @Produce  json
// @Success  200  {object}  requestIPResponse  "The Requester’s IP Address."
// @Router   /ip [get]
func requestIPHandler(c echo.Context) error {
	return c.JSONPretty(http.StatusOK, &requestIPResponse{
		Origin: getOrigin(c),
	}, "  ")
}

// @Summary  Return the incoming request's HTTP headers.
// @Tags     Request inspection
// @Produce  json
// @Success  200  {object}  requestHeadersResponse  "The request’s headers."
// @Router   /headers [get]
func requestHeadersHandler(c echo.Context) error {
	return c.JSONPretty(http.StatusOK, &requestHeadersResponse{
		Headers: getHeaders(c),
	}, "  ")
}

// @Summary  Return the incoming requests's User-Agent header.
// @Tags     Request inspection
// @Produce  json
// @Success  200  {object}  requestUserAgentResponse  "The request’s User-Agent header."
// @Router   /user-agent [get]
func requestUserAgentHandler(c echo.Context) error {
	return c.JSONPretty(http.StatusOK, &requestUserAgentResponse{
		UserAgent: getUserAgent(c),
	}, "  ")
}

type weightedCode struct {
	weight float64
	code   int
}

func chooseStatusCode(weightedCodes []weightedCode) int {
	var code int
	var total float64
	var cumWeights []float64
	for _, wc := range weightedCodes {
		total += wc.weight
		cumWeights = append(cumWeights, total)
	}
	rand.Seed(time.Now().UnixNano())
	x := rand.Float64() * total
	for i, cumWeight := range cumWeights {
		if cumWeight > x {
			code = weightedCodes[i].code
			break
		}
	}
	return code
}

// @Summary   Return status code or random status code if more than one are given
// @Tags      Status codes
// @Produce   plain
// @Param     codes  path  string  true  "codes"
// @Response  100    "Informational responses"
// @Response  200    "Success"
// @Response  300    "Redirection"
// @Response  400    "Client Errors"
// @Response  500    "Server Errors"
// @Router    /status/{codes} [delete]
// @Router    /status/{codes} [get]
// @Router    /status/{codes} [patch]
// @Router    /status/{codes} [post]
// @Router    /status/{codes} [put]
func statusCodesHandler(c echo.Context) error {
	codes, _ := url.PathUnescape(c.Param("codes"))
	if !strings.Contains(codes, ",") {
		code, err := strconv.Atoi(codes)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid status code")
		}
		return c.NoContent(code)
	}

	var weightedCodes []weightedCode
	var _code, _weight string
	for _, choice := range strings.Split(codes, ",") {
		if !strings.Contains(choice, ":") {
			_code = choice
			_weight = "1"
		} else {
			s := strings.SplitN(choice, ":", 2)
			_code, _weight = s[0], s[1]
		}
		code, err := strconv.Atoi(_code)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid status code")
		}
		weight, err := strconv.ParseFloat(_weight, 64)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid status code")
		}
		weightedCodes = append(weightedCodes, weightedCode{weight, code})
	}
	return c.NoContent(chooseStatusCode(weightedCodes))
}

//go:embed static/moby.html
var sampleHTML []byte

// @Summary   Returns a simple HTML document.
// @Tags      Response formats
// @Produce   html
// @Response  200  "An HTML page."
// @Router    /html [get]
func serveHTMLHandler(c echo.Context) error {
	return c.HTMLBlob(http.StatusOK, sampleHTML)
}

//go:embed static/sample.xml
var sampleXML []byte

// @Summary   Returns a simple XML document.
// @Tags      Response formats
// @Produce   xml
// @Response  200  "An XML document."
// @Router    /xml [get]
func serveXMLHandler(c echo.Context) error {
	return c.XMLBlob(http.StatusOK, sampleXML)
}

//go:embed static/sample.json
var sampleJSON []byte

// @Summary   Returns a simple JSON document.
// @Tags      Response formats
// @Produce   json
// @Response  200  "An JSON document."
// @Router    /json [get]
func serveJSONHandler(c echo.Context) error {
	return c.JSONBlob(http.StatusOK, sampleJSON)
}

const ROBOTS_TXT = `User-agent: *
Disallow: /deny
`

// @Summary   Returns some robots.txt rules.
// @Tags      Response formats
// @Produce   plain
// @Response  200  "Robots file"
// @Router    /robots.txt [get]
func serveRobotsTXTHandler(c echo.Context) error {
	return c.String(http.StatusOK, ROBOTS_TXT)
}

//go:embed static/deny.txt
var denyTXT string

// @Summary   Returns page denied by robots.txt rules.
// @Tags      Response formats
// @Produce   plain
// @Response  200  "Denied message"
// @Router    /deny [get]
func serveDenyHandler(c echo.Context) error {
	return c.String(http.StatusOK, denyTXT)
}

//go:embed static/sample-utf8.html
var sampleUTF8HTML []byte

// @Summary   Returns a UTF-8 encoded body.
// @Tags      Response formats
// @Produce   html
// @Response  200  "Encoded UTF-8 content."
// @Router    /encoding/utf8 [get]
func serveUTF8HTMLHandler(c echo.Context) error {
	return c.Blob(http.StatusOK, echo.MIMETextHTMLCharsetUTF8, sampleUTF8HTML)
}

// @Summary   Returns GZip-encoded data.
// @Tags      Response formats
// @Produce   json
// @Response  200  "GZip-encoded data."
// @Router    /gzip [get]
func serveGzipHandler(c echo.Context) error {
	res := gzippedResponse{}
	res.Origin = getOrigin(c)
	res.Headers = getHeaders(c)
	res.Method = c.Request().Method
	res.Gzipped = true
	return c.JSONPretty(http.StatusOK, &res, "  ")
}

// @Summary   Returns Deflate-encoded data.
// @Tags      Response formats
// @Produce   json
// @Response  200  "Defalte-encoded data."
// @Router    /deflate [get]
func serveDeflateHandler(c echo.Context) error {
	res := deflatedResponse{}
	res.Origin = getOrigin(c)
	res.Headers = getHeaders(c)
	res.Method = c.Request().Method
	res.Deflated = true
	return c.JSONPretty(http.StatusOK, &res, "  ")
}

// @Summary   Returns Brotli-encoded data.
// @Tags      Response formats
// @Produce   json
// @Response  200  "Brotli-encoded data."
// @Router    /brotli [get]
// @Deprecated
func serveBrotliHandler(c echo.Context) error {
	return nil
}

// @Summary   Returns n random bytes generated with given seed
// @Tags      Dynamic data
// @Produce   octet-stream
// @Param     n    path  int  true  "number of bytes"
// @Response  200  "Bytes."
// @Router    /bytes/{n} [get]
func generateBytesHandler(c echo.Context) error {
	// TODO: support seed querystring
	// https://github.com/postmanlabs/httpbin/blob/f8ec666b4d1b654e4ff6aedd356f510dcac09f83/httpbin/core.py#L1442
	n := c.Param("n")
	intN, err := strconv.Atoi(n)
	if err != nil || intN < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid number of bytes")
	}
	if intN > maxByteCount {
		intN = maxByteCount
	}
	bytes := make([]byte, intN)
	if _, err := rand.Read(bytes); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "generate random bytes failed")
	}
	return c.Blob(http.StatusOK, echo.MIMEOctetStream, bytes)
}

// @Summary   Returns a delayed response (max of 10 seconds).
// @Tags      Dynamic data
// @Produce   json
// @Param     delay  path  int  true  "delay"
// @Response  200    "A delayed response."
// @Router    /delay/{delay} [delete]
// @Router    /delay/{delay} [get]
// @Router    /delay/{delay} [patch]
// @Router    /delay/{delay} [post]
// @Router    /delay/{delay} [put]
func delayHandler(c echo.Context) error {
	delay := c.Param("delay")
	intDelay, err := strconv.Atoi(delay) // TODO: support float type delay
	if err != nil || intDelay < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid number of delay")
	}
	if intDelay > maxDelay {
		intDelay = maxDelay
	}
	time.Sleep(time.Duration(intDelay) * time.Second)
	data := ""
	files := getFiles(c)
	form := getForm(c)
	if len(files) == 0 && len(form) == 0 {
		data = getData(c)
	}
	return c.JSONPretty(http.StatusOK, &delayResponse{
		Args:    getArgs(c),
		Data:    data,
		Files:   files,
		Form:    form,
		Headers: getHeaders(c),
		Origin:  getOrigin(c),
		URL:     getURL(c),
	}, "  ")
}

type dripParams struct {
	// The amount of time (in seconds) over which to drip each byte
	Duration float64 `query:"duration" default:"2"`
	// The number of bytes to respond with
	Numbytes int `query:"numbytes" default:"10"`
	// The response code that will be returned
	Code int `query:"code" default:"200"`
	// The amount of time (in seconds) to delay before responding
	Delay float64 `query:"delay" default:"2"`
}

// @Summary   Drips data over a duration after an optional initial delay.
// @Tags      Dynamic data
// @Produce   octet-stream
// @Param     dripParams  query  dripParams  true  "dripParams"
// @Response  200         "A dripped response."
// @Router    /drip [get]
func dripHandler(c echo.Context) error {
	dp := &dripParams{
		Duration: 2,
		Numbytes: 10,
		Code:     200,
		Delay:    2,
	}
	if err := c.Bind(dp); err != nil {
		return err
	}

	if dp.Delay < 0 {
		dp.Delay = 0
	} else if dp.Delay > 10 {
		dp.Delay = 10
	}

	if dp.Duration < 0.1 {
		dp.Duration = 0.1 // Minimum duration = 100 Millisecond
	} else if dp.Duration > 60 {
		dp.Duration = 60
	}

	if dp.Numbytes < 0 {
		dp.Numbytes = 0
	} else if dp.Numbytes > 10<<20 {
		dp.Numbytes = 10 << 20 // Millisecond
	}

	time.Sleep(time.Duration(dp.Delay*1000) * time.Millisecond)

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEOctetStream)
	c.Response().Header().Set(echo.HeaderContentLength, strconv.Itoa(dp.Numbytes))
	c.Response().WriteHeader(dp.Code) // TODO: validate status code?

	remainBytes := dp.Numbytes
	times := int(dp.Duration / 0.1)
	chunkLength := dp.Numbytes
	if times > 1 {
		chunkLength = dp.Numbytes/times + 1
	}
	if chunkLength == 1 {
		pause := int(dp.Duration*1000) / remainBytes
		for remainBytes > 0 {
			if _, err := c.Response().Write([]byte{'*'}); err != nil {
				return err
			}
			c.Response().Flush()
			time.Sleep(time.Duration(pause) * time.Millisecond)
			remainBytes--
		}
	} else {
		for remainBytes > 0 {
			var length int
			if remainBytes > chunkLength {
				length = chunkLength
			} else {
				length = remainBytes
			}
			if _, err := c.Response().Write(bytes.Repeat([]byte{'*'}, length)); err != nil {
				return err
			}
			c.Response().Flush()
			time.Sleep(100 * time.Millisecond)
			remainBytes -= length
		}
	}
	return nil
}
