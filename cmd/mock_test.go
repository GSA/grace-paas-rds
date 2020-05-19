package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
)

// nolint: gochecknoglobals
var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server

	// req is mock provisioning requested
	mockReq *req
)

func setup() {
	var err error
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	mockReq, err = newReq()
	if err != nil {
		panic(fmt.Sprintf("couldn't initialize request: %v", err))
	}

	mockURL, err := url.Parse(server.URL)
	if err != nil {
		panic(fmt.Sprintf("couldn't parse test server URL: %s", server.URL))
	}

	mockReq.circleClient.BaseURL = mockURL
	mockReq.githubURL = mockURL.Hostname()
	mockReq.snowClient.Instance = mockURL.Hostname()
}

func teardown() {
	defer server.Close()
}
