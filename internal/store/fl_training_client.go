package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type FLTrainingClient struct {
	ID           uuid.UUID `json:"id"`
	FLTrainingID string    `json:"fl_training_id"`
	PartitionID  int       `json:"partition_id"`
	NodeName     string    `json:"node_name"`
	PodName      string    `json:"pod_name"`
	State        string    `json:"state"`
	LastLogRead  time.Time `json:"last_log_read"`
	CreatedAt    time.Time `json:"created_at"`
}

type FLTrainingClientStore struct {
	db *sql.DB
}

func NewFLTrainingClientStore(db *sql.DB) *FLTrainingClientStore {
	return &FLTrainingClientStore{db: db}
}

func (s *FLTrainingClientStore) UpdateState(
	ctx context.Context,
	flTrainingID string,
	partitionID int,
	state string,
) error {
	query := `
		UPDATE training_clients
		SET state = $3
		WHERE fl_training_id = $1 AND partition_id = $2
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, flTrainingID, partitionID, state)
	return err
}

func (s *FLTrainingClientStore) GetAll(ctx context.Context) ([]FLTrainingClient, error) {
	query := `
		SELECT 
			id,
			fl_training_id,
			partition_id,
			node_name,
			pod_name,
			state,
			created_at,
			last_log_read
		FROM training_clients
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []FLTrainingClient

	for rows.Next() {
		var c FLTrainingClient
		err := rows.Scan(
			&c.ID,
			&c.FLTrainingID,
			&c.PartitionID,
			&c.NodeName,
			&c.PodName,
			&c.State,
			&c.CreatedAt,
			&c.LastLogRead,
		)
		if err != nil {
			return nil, err
		}

		clients = append(clients, c)
	}

	return clients, nil
}

func (s *FLTrainingClientStore) Create(ctx context.Context, c FLTrainingClient) error {
	query := `
		INSERT INTO training_clients (
			fl_training_id,
			partition_id,
			node_name,
			pod_name,
			state,
			last_log_read
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		c.FLTrainingID,
		c.PartitionID,
		c.NodeName,
		c.PodName,
		c.State,
		c.LastLogRead,
	).Scan(
		&c.ID,
		&c.CreatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *FLTrainingClientStore) GetByFLTrainingIDAndPartitionID(
	ctx context.Context,
	flTrainingID string,
	partitionID int,
) (FLTrainingClient, error) {
	query := `
		SELECT
			id,
			fl_training_id,
			partition_id,
			node_name,
			pod_name,
			state,
			created_at,
			last_log_read
		FROM training_clients
		WHERE fl_training_id = $1 AND partition_id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var c FLTrainingClient
	err := s.db.QueryRowContext(ctx, query, flTrainingID, partitionID).Scan(
		&c.ID,
		&c.FLTrainingID,
		&c.PartitionID,
		&c.NodeName,
		&c.PodName,
		&c.State,
		&c.CreatedAt,
		&c.LastLogRead,
	)
	if err != nil {
		return FLTrainingClient{}, err
	}

	return c, nil
}

func (s *FLTrainingClientStore) EnsureByFLTrainingIDAndPartitionID(
	ctx context.Context,
	flTrainingID string,
	partitionID int,
	nodeName string,
	podName string,
	state string,
) (FLTrainingClient, error) {
	c, err := s.GetByFLTrainingIDAndPartitionID(ctx, flTrainingID, partitionID)
	if err == nil {
		return c, nil
	}
	if err != sql.ErrNoRows {
		return FLTrainingClient{}, err
	}

	newClient := FLTrainingClient{
		FLTrainingID: flTrainingID,
		PartitionID:  partitionID,
		NodeName:     nodeName,
		PodName:      podName,
		State:        state,
		LastLogRead:  time.Time{},
	}

	if err := s.Create(ctx, newClient); err != nil {
		return FLTrainingClient{}, err
	}

	return s.GetByFLTrainingIDAndPartitionID(ctx, flTrainingID, partitionID)
}

func (s *FLTrainingClientStore) UpdateLastLogRead(
	ctx context.Context,
	flTrainingID string,
	partitionID int,
	t time.Time,
) error {
	query := `
		UPDATE training_clients
		SET last_log_read = $3
		WHERE fl_training_id = $1 AND partition_id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, flTrainingID, partitionID, t)
	return err
}
