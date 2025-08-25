package evstate

import (
	"context"
	"fmt"
	"github.com/IsaacDSC/workqueue"
	"github.com/IsaacDSC/workqueue/SDK"
)

type Event struct {
	CurrentState string
	Triggers     []string
	Producer     *SDK.Producer
}

func (e Event) Publisher(ctx context.Context, payload any) error {
	for _, trigger := range e.Triggers {
		payload := workqueue.NewInputBuilder().
			WithEvent(trigger).
			WithData(payload).
			Build()

		if err := e.Producer.Publish(ctx, payload); err != nil {
			return fmt.Errorf("failed to publish event: %v", err)
		}
	}

	return nil
}
