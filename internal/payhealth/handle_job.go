package payhealth

import (
	"context"
	"github.com/IsaacDSC/rinhabackend2025/internal/paystate"
	"log"
	"time"
)

type PayState interface {
	Set(ctx context.Context, processor paystate.Processor) error
	Get(ctx context.Context) paystate.Processor
}

func StartJob(ctx context.Context, state PayState, dp paystate.Processor, fp paystate.Processor) {
	log.Println("StartJob Pay Health")
	for {
		if err := dp.Health(ctx); err != nil {
			if err := state.Set(ctx, fp); err != nil {
				log.Printf("failed to set processor state: %v", err)
			}

			if err := fp.Health(ctx); err != nil {
				if err := state.Set(ctx, dp); err != nil {
					log.Printf("failed to set processor state: %v", err)
				}
			}
		}

		time.Sleep(time.Second * 5)
	}
}
