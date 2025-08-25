package wpay

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IsaacDSC/rinhabackend2025/pkg/handle"
	"github.com/google/uuid"
	"net/http"
)

type PayEventPaymentProcessed interface {
	Publisher(ctx context.Context, payload any) error
}

type PaymentProcessed struct {
	TxID          uuid.UUID
	TypeProcessor string
}

func EventPaymentProcessed(store PaymentStore, eventState PayEventPaymentProcessed) handle.HandleHTTP {
	return handle.HandleHTTP{
		Path: "POST /worker-queue/payment/processed",
		Handle: func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			w.Header().Set("Content-Type", "application/json")

			//var body workqueue.Payload
			var payload PaymentProcessed
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			if err := store.UpdateCompleteStatus(ctx, payload.TxID, payload.TypeProcessor); err != nil {
				http.Error(w, fmt.Sprintf("failed to update transaction: %v", err), http.StatusInternalServerError)
				return
			}

			if err := eventState.Publisher(ctx, PaymentProcessed{TxID: payload.TxID}); err != nil {
				http.Error(w, fmt.Sprintf("failed to produce event: %v", err), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
		},
	}
}
