package wpay

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IsaacDSC/rinhabackend2025/internal/paystore"
	"github.com/IsaacDSC/rinhabackend2025/pkg/handle"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"time"
)

type CreateTransactionRequest struct {
	Amount        float64   `json:"amount"`
	CorrelationID uuid.UUID `json:"correlationId"`
}

func (ctr CreateTransactionRequest) ToTransaction() paystore.Transaction {
	conv := func(amount float64) string {
		amountInCents := fmt.Sprintf("%.2f", amount)
		return strings.Replace(amountInCents, ".", "", -1)
	}
	return paystore.Transaction{
		ID:            uuid.New(),
		CorrelationID: ctr.CorrelationID,
		Amount:        conv(ctr.Amount), // Assuming amount is in cents
		CreatedAt:     time.Now(),
	}
}

type PayCmdEventState interface {
	Publisher(ctx context.Context, payload any) error
}

func CmdPaymentProcessor(store PaymentStore, event PayCmdEventState) handle.HandleHTTP {
	return handle.HandleHTTP{
		Path: "POST /payments",
		Handle: func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			w.Header().Set("Content-Type", "application/json")

			var body CreateTransactionRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			transaction := body.ToTransaction()
			if err := store.CreateTransaction(ctx, transaction); err != nil {
				http.Error(w, fmt.Sprintf("failed to create transaction: %v", err), http.StatusInternalServerError)
				return
			}

			if err := event.Publisher(ctx, transaction); err != nil {
				http.Error(w, "Failed to publish event", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
		},
	}
}
