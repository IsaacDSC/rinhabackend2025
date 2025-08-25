package paystore

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"strings"
	"time"
)

type MySQLStore struct {
	conn *sql.DB
}

func NewMySQLStore(db *sql.DB) MySQLStore {
	return MySQLStore{conn: db}
}

type PaymentModel struct {
	Kind   string
	Amount int
}

type SummaryResponse struct {
	ResponseDefault  PaymentSummary `json:"default"`
	ResponseFallback PaymentSummary `json:"fallback"`
}

type PaymentSummary struct {
	TotalRequests int     `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

func (s MySQLStore) GetPayments(ctx context.Context) (SummaryResponse, error) {
	const query = `SELECT processor_type, amount FROM transactions WHERE status = 'completed';`
	rows, err := s.conn.Query(query)
	if err != nil {
		return SummaryResponse{}, fmt.Errorf("failed to query rows: %w", err)
	}

	var (
		totalTransDefault    int
		totalRequestDefault  int
		totalTransFallback   int
		totalRequestFallback int
		data                 PaymentModel
		scanErr              error
	)

	for rows.Next() {
		if err := rows.Scan(&data.Kind, &data.Amount); err != nil {
			scanErr = fmt.Errorf("failed to scan row: %w", err)
			break
		}

		if strings.Contains(data.Kind, "default") {
			totalTransDefault += data.Amount
			totalRequestDefault++
		}

		if strings.Contains(data.Kind, "fallback") {
			totalTransFallback += data.Amount
			totalRequestFallback++
		}
	}

	if scanErr != nil {
		return SummaryResponse{}, err
	}

	return SummaryResponse{
		ResponseDefault: PaymentSummary{
			TotalRequests: totalRequestDefault,
			TotalAmount:   float64(totalTransDefault),
		},
		ResponseFallback: PaymentSummary{
			TotalRequests: totalRequestFallback,
			TotalAmount:   float64(totalTransFallback),
		},
	}, nil
}

type Transaction struct {
	ID            uuid.UUID
	CorrelationID uuid.UUID
	Amount        string
	CreatedAt     time.Time
}

func (s MySQLStore) CreateTransaction(ctx context.Context, transaction Transaction) error {
	const query = `INSERT INTO transactions (id, correlation_id, amount, created_at) VALUES (?, ?, ?, ?);`
	tx, err := s.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer tx.Rollback()

	result, err := s.conn.Exec(query, transaction.ID, transaction.CorrelationID, transaction.Amount, transaction.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	totalAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	if totalAffected != 1 {
		return fmt.Errorf("failed to insert transaction: expected 1 row affected, got %d", totalAffected)
	}

	return tx.Commit()
}

func (s MySQLStore) UpdateCompleteStatus(ctx context.Context, txID uuid.UUID, processorType string) error {
	const query = `UPDATE transactions SET status = 'completed', processor_type = ? WHERE (status = 'pending' AND id = ?);`
	tx, err := s.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	result, err := s.conn.Exec(query, processorType, txID)
	if err != nil {
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	totalAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	if totalAffected != 1 {
		return fmt.Errorf("failed to insert transaction: expected 1 row affected, got %d", totalAffected)
	}

	return tx.Commit()
}
