package store

import (
	"context"
	"database/sql"
)

type FLModelWeightsStore struct {
	db *sql.DB
}

func NewFLModelWeightsStore(db *sql.DB) *FLModelWeightsStore {
	return &FLModelWeightsStore{db: db}
}

// Upsert inserts a new row for (fl_training_id, server_round) or updates payload if it already exists.
func (s *FLModelWeightsStore) Upsert(
	ctx context.Context,
	flTrainingID string,
	serverRound int,
	payload []byte,
) error {
	query := `
		INSERT INTO fl_model_weights (
			fl_training_id,
			server_round,
			payload
		)
		VALUES ($1, $2, $3)
		ON CONFLICT (fl_training_id, server_round)
		DO UPDATE SET payload = EXCLUDED.payload
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, flTrainingID, serverRound, payload)
	return err
}
