package main

import (
	"encoding/json"

	"github.com/labstack/echo/v4"
)

type echobinJSONSerializer struct {
	echo.DefaultJSONSerializer
}

// Serialize converts an interface into a json and writes it to the response.
// You can optionally use the indent parameter to produce pretty JSONs.
func (d echobinJSONSerializer) Serialize(c echo.Context, i interface{}, indent string) error {
	enc := json.NewEncoder(c.Response())
	// https://github.com/golang/go/issues/28453
	enc.SetEscapeHTML(false)
	if indent != "" {
		enc.SetIndent("", indent)
	}
	return enc.Encode(i)
}
