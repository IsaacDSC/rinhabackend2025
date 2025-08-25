package middleware

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"
)

type LoggingRoundTripper struct {
	transport http.RoundTripper
}

func NewLoggingRoundTripper(transport http.RoundTripper) *LoggingRoundTripper {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &LoggingRoundTripper{
		transport: transport,
	}
}

func (lrt *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	var requestBody string
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err == nil {
			requestBody = string(bodyBytes)
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}

	resp, err := lrt.transport.RoundTrip(req)
	if err != nil {
		duration := time.Since(start)
		log.Printf(
			"[HTTP_CLIENT] %s %s | Error: %v | Duration: %v | Request Body: %s",
			req.Method,
			req.URL.String(),
			err,
			duration,
			requestBody,
		)
		return nil, err
	}

	var responseBody string
	if resp.Body != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err == nil {
			responseBody = string(bodyBytes)
			resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}

	duration := time.Since(start)
	if resp.StatusCode >= 400 {
		log.Printf(
			"[HTTP_CLIENT] %s %s | Status: %d | Duration: %v | Request Body: %s | Response Body: %s",
			req.Method,
			req.URL.String(),
			resp.StatusCode,
			duration,
			requestBody,
			responseBody,
		)
	}

	return resp, nil
}

func NewLoggingHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: NewLoggingRoundTripper(nil),
	}
}
