package payprocess

import "context"

type Processor interface {
	Health(ctx context.Context) error
	ProcessPayment(ctx context.Context, payload PaymentRequest) error
	Name() string
}
