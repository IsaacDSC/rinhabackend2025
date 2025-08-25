package rpay

import (
	"context"
	"encoding/json"
	"github.com/IsaacDSC/rinhabackend2025/internal/paystore"
	"github.com/IsaacDSC/rinhabackend2025/pkg/handle"
	"net/http"
)

type PaymentStore interface {
	GetPayments(ctx context.Context) (paystore.SummaryResponse, error)
}

func GetHandleHTTP(store PaymentStore) handle.HandleHTTP {
	return handle.HandleHTTP{
		Path: "GET /payments-summary",
		Handle: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			body, err := store.GetPayments(r.Context())
			if err != nil {
				http.Error(w, "Failed to get payments summary", http.StatusInternalServerError)
				return
			}

			if err := json.NewEncoder(w).Encode(body); err != nil {
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				return
			}
		},
	}
}
