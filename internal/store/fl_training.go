package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type FLTraining struct {
	ID                 uuid.UUID `json:"id"`
	FLTrainingID       string    `json:"fl_training_id"`
	CurrentServerRound int       `json:"current_server_round"`
	TotalServerRound   int       `json:"total_server_round"`
	CreatedAt          time.Time `json:"created_at"`
}

type FLTrainingStore struct {
	db *sql.DB
}

func NewFLTrainingStore(db *sql.DB) *FLTrainingStore {
	return &FLTrainingStore{db: db}
}

func (s *FLTrainingStore) GetAll(ctx context.Context) ([]FLTraining, error) {
	query := `
		SELECT
			id,
			fl_training_id,
			current_server_round,
			total_server_round,
			created_at
		FROM fl_trainings
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flTrainings []FLTraining

	for rows.Next() {
		var f FLTraining
		err := rows.Scan(
			&f.ID,
			&f.FLTrainingID,
			&f.CurrentServerRound,
			&f.TotalServerRound,
			&f.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		flTrainings = append(flTrainings, f)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return flTrainings, nil
}

func (s *FLTrainingStore) Create(ctx context.Context, flTraining FLTraining) error {
	query := `
		INSERT INTO fl_trainings (fl_training_id)
		VALUES ($1)
		RETURNING id, fl_training_id, current_server_round, total_server_round, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		flTraining.FLTrainingID,
	).Scan(
		&flTraining.ID,
		&flTraining.FLTrainingID,
		&flTraining.CurrentServerRound,
		&flTraining.TotalServerRound,
		&flTraining.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *FLTrainingStore) GetByFLTrainingID(ctx context.Context, flTrainingID string) (FLTraining, error) {
	query := `
		SELECT
			id,
			fl_training_id,
			current_server_round,
			total_server_round,
			created_at
		FROM fl_trainings
		WHERE fl_training_id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var f FLTraining
	err := s.db.QueryRowContext(ctx, query, flTrainingID).Scan(
		&f.ID,
		&f.FLTrainingID,
		&f.CurrentServerRound,
		&f.TotalServerRound,
		&f.CreatedAt,
	)
	if err != nil {
		return FLTraining{}, err
	}

	return f, nil
}

// Ensure returns the FLTraining row for the given ID, creating it if it does not exist.
func (s *FLTrainingStore) Ensure(ctx context.Context, flTrainingID string) (FLTraining, error) {
	f, err := s.GetByFLTrainingID(ctx, flTrainingID)
	if err == nil {
		return f, nil
	}
	if err != sql.ErrNoRows {
		return FLTraining{}, err
	}

	if err := s.Create(ctx, FLTraining{FLTrainingID: flTrainingID}); err != nil {
		return FLTraining{}, err
	}

	return s.GetByFLTrainingID(ctx, flTrainingID)
}

func (s *FLTrainingStore) UpdateCurrentServerRound(
	ctx context.Context,
	flTrainingID string,
	serverRound int,
) error {
	query := `
		UPDATE fl_trainings
		SET current_server_round = GREATEST(current_server_round, $2)
		WHERE fl_training_id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, flTrainingID, serverRound)
	return err
}

func (s *FLTrainingStore) UpdateTotalServerRound(
	ctx context.Context,
	flTrainingID string,
	totalServerRound int,
) error {
	query := `
		UPDATE fl_trainings
		SET total_server_round = $2
		WHERE fl_training_id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, flTrainingID, totalServerRound)
	return err
}
