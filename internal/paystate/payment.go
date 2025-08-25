package paystate

import (
	"context"
	"github.com/IsaacDSC/rinhabackend2025/internal/payprocess"
	"github.com/redis/go-redis/v9"
)

type Processor interface {
	Health(ctx context.Context) error
	ProcessPayment(ctx context.Context, payload payprocess.PaymentRequest) error
	Name() string
}

type PayState interface {
	Set(ctx context.Context, processor Processor) error
	Get(ctx context.Context) Processor
}

type State struct {
	cache  *redis.Client
	mapper map[string]Processor
}

var _ PayState = (*State)(nil)

func NewState(cache *redis.Client, defaultProcessor Processor, fallbackProcessor Processor) PayState {
	return State{
		cache: cache,
		mapper: map[string]Processor{
			"default":  defaultProcessor,
			"fallback": fallbackProcessor,
		},
	}
}

const key = "rinhabackend2025.processor.registry"

func (s State) Set(ctx context.Context, processor Processor) error {
	return s.cache.Set(ctx, key, processor.Name(), 0).Err()
}

func (s State) Get(ctx context.Context) Processor {
	processorStr := s.cache.Get(ctx, key).String()

	if processor, exists := s.mapper[processorStr]; exists {
		return processor
	}

	return s.mapper["default"]
}
