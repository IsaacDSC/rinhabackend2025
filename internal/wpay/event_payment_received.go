package wpay

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IsaacDSC/rinhabackend2025/internal/payprocess"
	"github.com/IsaacDSC/rinhabackend2025/internal/paystate"
	"github.com/IsaacDSC/rinhabackend2025/internal/paystore"
	"github.com/IsaacDSC/rinhabackend2025/pkg/handle"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type PaymentStore interface {
	CreateTransaction(ctx context.Context, transaction paystore.Transaction) error
	UpdateCompleteStatus(ctx context.Context, txID uuid.UUID, processorType string) error
}

type StateProcessor interface {
	Get(ctx context.Context) paystate.Processor
}

type PayReceivedEventState interface {
	Publisher(ctx context.Context, payload any) error
}

func EventPaymentReceived(stateProcessor StateProcessor, eventState PayReceivedEventState) handle.HandleHTTP {
	return handle.HandleHTTP{
		Path: "POST /worker-queue/payment/received",
		Handle: func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			w.Header().Set("Content-Type", "application/json")

			//var body workqueue.Payload
			var transaction paystore.Transaction
			if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			processor := stateProcessor.Get(ctx)
			if err := processor.ProcessPayment(ctx, payprocess.PaymentRequest{
				CorrelationID: transaction.CorrelationID.String(),
				Amount:        transaction.Amount, // Convert cents to dollars
				RequestTime:   time.Now().String(),
			}); err != nil {
				http.Error(w, fmt.Sprintf("error on processing payment: %v", err), http.StatusInternalServerError)
				return
			}

			if err := eventState.Publisher(ctx, PaymentProcessed{TxID: transaction.ID, TypeProcessor: processor.Name()}); err != nil {
				http.Error(w, fmt.Sprintf("failed to publish transaction: %v", err), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
		},
	}
}
