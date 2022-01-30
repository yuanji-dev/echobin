package main

import (
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
// @Produce   json
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
