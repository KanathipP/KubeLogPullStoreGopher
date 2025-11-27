package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type TrainingGraph struct {
	ID           uuid.UUID `json:"id"`
	ClientID     uuid.UUID `json:"client_id"`
	ServerRound  int       `json:"server_round"`
	Optimizer    string    `json:"optimizer"`
	LearningRate float64   `json:"learning_rate"`
	NumEpochs    int       `json:"num_epochs"`
	BatchSize    int       `json:"batch_size"`
	CreatedAt    time.Time `json:"created_at"`
}

type TrainingGraphPoint struct {
	ID               uuid.UUID `json:"id"`
	GraphID          uuid.UUID `json:"graph_id"`
	CurrentEpoch     int       `json:"current_epoch"`
	TrainedBatch     int       `json:"trained_batch"`
	TrainLoss        float64   `json:"train_loss"`
	ValLoss          float64   `json:"val_loss"`
	Accuracy         float64   `json:"accuracy"`
	EpochElapsedTime float64   `json:"epoch_elapsed_time"`
	CreatedAt        time.Time `json:"created_at"`
}

type TrainingGraphStore struct {
	db *sql.DB
}

func NewTrainingGraphStore(db *sql.DB) *TrainingGraphStore {
	return &TrainingGraphStore{db: db}
}

func (s *TrainingGraphStore) Create(ctx context.Context, g TrainingGraph) error {
	query := `
		INSERT INTO training_graphs (
			client_id,
			server_round,
			optimizer,
			learning_rate,
			num_epochs,
			batch_size
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		g.ClientID,
		g.ServerRound,
		g.Optimizer,
		g.LearningRate,
		g.NumEpochs,
		g.BatchSize,
	).Scan(
		&g.ID,
		&g.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *TrainingGraphStore) CreatePoint(ctx context.Context, p TrainingGraphPoint) error {
	query := `
		INSERT INTO training_graph_points (
			graph_id,
			current_epoch,
			trained_batch,
			train_loss,
			val_loss,
			accuracy,
			epoch_elapsed_time
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		p.GraphID,
		p.CurrentEpoch,
		p.TrainedBatch,
		p.TrainLoss,
		p.ValLoss,
		p.Accuracy,
		p.EpochElapsedTime,
	).Scan(
		&p.ID,
		&p.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *TrainingGraphStore) GetGraphIDsByClientID(ctx context.Context, clientID uuid.UUID) ([]uuid.UUID, error) {
	query := `
		SELECT id
		FROM training_graphs
		WHERE client_id = $1
		ORDER BY server_round ASC
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID

	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

func (s *TrainingGraphStore) GetGraphsByClientID(ctx context.Context, clientID uuid.UUID) ([]TrainingGraph, error) {
	query := `
		SELECT 
			id,
			client_id,
			server_round,
			optimizer,
			learning_rate,
			num_epochs,
			batch_size,
			created_at
		FROM training_graphs
		WHERE client_id = $1
		ORDER BY server_round ASC
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var graphs []TrainingGraph

	for rows.Next() {
		var g TrainingGraph
		err := rows.Scan(
			&g.ID,
			&g.ClientID,
			&g.ServerRound,
			&g.Optimizer,
			&g.LearningRate,
			&g.NumEpochs,
			&g.BatchSize,
			&g.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		graphs = append(graphs, g)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return graphs, nil
}

func (s *TrainingGraphStore) GetPointsByGraphID(ctx context.Context, graphID uuid.UUID) ([]TrainingGraphPoint, error) {
	query := `
		SELECT
			id,
			graph_id,
			current_epoch,
			trained_batch,
			train_loss,
			val_loss,
			accuracy,
			epoch_elapsed_time,
			created_at
		FROM training_graph_points
		WHERE graph_id = $1
		ORDER BY current_epoch ASC, created_at ASC
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, graphID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []TrainingGraphPoint

	for rows.Next() {
		var p TrainingGraphPoint
		err := rows.Scan(
			&p.ID,
			&p.GraphID,
			&p.CurrentEpoch,
			&p.TrainedBatch,
			&p.TrainLoss,
			&p.ValLoss,
			&p.Accuracy,
			&p.EpochElapsedTime,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		points = append(points, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return points, nil
}
