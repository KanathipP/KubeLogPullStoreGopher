package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type TestingGraph struct {
	ID        uuid.UUID `json:"id"`
	ClientID  uuid.UUID `json:"client_id"`
	CreatedAt time.Time `json:"created_at"`
}

type TestingGraphPoint struct {
	ID          uuid.UUID `json:"id"`
	GraphID     uuid.UUID `json:"graph_id"`
	ServerRound int       `json:"server_round"`
	Criterion   string    `json:"criterion"`
	BatchSize   int       `json:"batch_size"`
	TestLoss    float64   `json:"test_loss"`
	Accuracy    float64   `json:"accuracy"`
	CreatedAt   time.Time `json:"created_at"`
}

type TestingGraphStore struct {
	db *sql.DB
}

func NewTestingGraphStore(db *sql.DB) *TestingGraphStore {
	return &TestingGraphStore{db: db}
}

func (s *TestingGraphStore) Create(ctx context.Context, g TestingGraph) error {
	query := `
		INSERT INTO testing_graphs (
			client_id
		)
		VALUES ($1)
		RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		g.ClientID,
	).Scan(
		&g.ID,
		&g.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *TestingGraphStore) CreatePoint(ctx context.Context, p TestingGraphPoint) error {
	query := `
		INSERT INTO testing_graph_points (
			graph_id,
			server_round,
			criterion,
			batch_size,
			test_loss,
			accuracy
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		p.GraphID,
		p.ServerRound,
		p.Criterion,
		p.BatchSize,
		p.TestLoss,
		p.Accuracy,
	).Scan(
		&p.ID,
		&p.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *TestingGraphStore) GetGraphIDsByClientID(ctx context.Context, clientID uuid.UUID) ([]uuid.UUID, error) {
	query := `
		SELECT id
		FROM testing_graphs
		WHERE client_id = $1
		ORDER BY created_at ASC
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

func (s *TestingGraphStore) GetGraphsByClientID(ctx context.Context, clientID uuid.UUID) ([]TestingGraph, error) {
	query := `
		SELECT 
			id,
			client_id,
			created_at
		FROM testing_graphs
		WHERE client_id = $1
		ORDER BY created_at ASC
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var graphs []TestingGraph

	for rows.Next() {
		var g TestingGraph
		err := rows.Scan(
			&g.ID,
			&g.ClientID,
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

func (s *TestingGraphStore) GetPointsByGraphID(ctx context.Context, graphID uuid.UUID) ([]TestingGraphPoint, error) {
	query := `
		SELECT
			id,
			graph_id,
			server_round,
			criterion,
			batch_size,
			test_loss,
			accuracy,
			created_at
		FROM testing_graph_points
		WHERE graph_id = $1
		ORDER BY server_round ASC, created_at ASC
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, graphID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []TestingGraphPoint

	for rows.Next() {
		var p TestingGraphPoint
		err := rows.Scan(
			&p.ID,
			&p.GraphID,
			&p.ServerRound,
			&p.Criterion,
			&p.BatchSize,
			&p.TestLoss,
			&p.Accuracy,
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
