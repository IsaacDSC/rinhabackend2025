package wpay

import (
	"github.com/IsaacDSC/rinhabackend2025/pkg/evstate"
	"github.com/IsaacDSC/workqueue/SDK"
)

const (
	eventPaymentReceived  = "payment.received"
	eventPaymentProcessed = "payment.processed"
)

const (
	cmdPaymentProcessor = "payment.processor"
)

func NewCmdPaymentProcessor(producer *SDK.Producer) evstate.Event {
	return evstate.Event{
		CurrentState: cmdPaymentProcessor,
		Triggers:     []string{eventPaymentReceived},
		Producer:     producer,
	}
}

func NewEventPaymentReceived(producer *SDK.Producer) evstate.Event {
	return evstate.Event{
		CurrentState: eventPaymentReceived,
		Triggers:     []string{eventPaymentProcessed},
		Producer:     producer,
	}
}

func NewEventPaymentProcessed(producer *SDK.Producer) evstate.Event {
	return evstate.Event{
		CurrentState: eventPaymentProcessed,
		Triggers:     []string{},
		Producer:     producer,
	}
}
