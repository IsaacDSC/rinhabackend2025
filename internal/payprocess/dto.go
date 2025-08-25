package payprocess

type PaymentRequest struct {
	CorrelationID string `json:"correlationId"`
	Amount        string `json:"amount"`
	RequestTime   string `json:"requestAt"`
}

type PaymentResponse struct {
	Message string `json:"message"`
}

type HealthResponse struct {
	Failing         bool `json:"failing"`
	MinResponseTime int  `json:"minResponseTime"`
}
