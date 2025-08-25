package payprocess

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/IsaacDSC/rinhabackend2025/pkg/middleware"
	"github.com/google/uuid"
	"net/http"
	"net/url"
	"time"
)

type PayProcessor struct {
	client        *http.Client
	baseURL       *url.URL
	processorName string
}

func NewPaymentProcessor(processorName string, baseUrl *url.URL) *PayProcessor {
	return &PayProcessor{
		baseURL:       baseUrl,
		processorName: processorName,
		client:        middleware.NewLoggingHTTPClient(30 * time.Second),
	}
}

var _ Processor = (*PayProcessor)(nil)

func (d PayProcessor) Name() string {
	return fmt.Sprintf("processor.%s", d.processorName)
}

func (d PayProcessor) Health(ctx context.Context) error {
	u := *d.baseURL
	u.Path = "/payments/service-health"
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	d.setDefaultHeaders(req)

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make health check request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("health check failed: received status %d", resp.StatusCode)
	}

	var healthResp HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		return fmt.Errorf("failed to parse health response: %w", err)
	}

	if healthResp.Failing {
		return fmt.Errorf("health check indicates service is failing")
	}

	return nil
}

func (d PayProcessor) ProcessPayment(ctx context.Context, payload PaymentRequest) error {
	u := *d.baseURL
	u.Path = "/payments"
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payment request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	d.setDefaultHeaders(req)
	req.Header.Set("X-Correlation-ID", payload.CorrelationID)

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make payment request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("payment request failed: received status %d", resp.StatusCode)
	}

	return nil
}

func (d PayProcessor) setDefaultHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("%s/1.0", d.Name()))
	req.Header.Set("X-Request-Id", uuid.New().String())
}
