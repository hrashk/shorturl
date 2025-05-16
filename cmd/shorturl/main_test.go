package main

import (
	"os"
	"testing"

	"github.com/hrashk/shorturl/internal/app"
	"github.com/stretchr/testify/assert"
)

func Test_readConfigFromArgs(t *testing.T) {
	const listen = "localhost:9999"
	const redirect = "http://example.com:1024"
	os.Args = []string{"", "-a", listen, "-b", redirect}

	readConfigFromArgs()

	assert.Equal(t, listen, app.GetListenAddr())
	assert.Equal(t, redirect, app.GetRedirectBaseURL())
}
