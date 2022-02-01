package main

import (
	_ "embed"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

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
