package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRequestRange(t *testing.T) {
	cases := []struct {
		rawRange string
		length   int
		start    int
		end      int
	}{
		{"unknownUnit=0-500", 10000, 0, 9999},
		{"bytes=", 10000, 0, 9999},
		{"bytes=0-500", 10000, 0, 500},
		{"bytes=500-", 10000, 500, 9999},
		{"bytes=-500", 10000, 9500, 9999},
	}
	for _, v := range cases {
		start, end := getRequestRange(v.rawRange, v.length)
		assert.Equal(t, v.start, start)
		assert.Equal(t, v.end, end)
	}
}
