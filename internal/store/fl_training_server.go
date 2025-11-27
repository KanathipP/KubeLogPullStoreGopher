package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type FLTrainingServer struct {
	ID           uuid.UUID `json:"id"`
	FLTrainingID string    `json:"fl_training_id"`
	NodeName     string    `json:"node_name"`
	PodName      string    `json:"pod_name"`
	LastLogRead  time.Time `json:"last_log_read"`
	CreatedAt    time.Time `json:"created_at"`
}

type FLTrainingServerStore struct {
	db *sql.DB
}

func NewFLTrainingServerStore(db *sql.DB) *FLTrainingServerStore {
	return &FLTrainingServerStore{db: db}
}

func (s *FLTrainingServerStore) GetByFLTrainingID(
	ctx context.Context,
	flTrainingID string,
) (FLTrainingServer, error) {
	query := `
		SELECT
			id,
			fl_training_id,
			node_name,
			pod_name,
			created_at,
			last_log_read
		FROM training_servers
		WHERE fl_training_id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var srv FLTrainingServer
	err := s.db.QueryRowContext(ctx, query, flTrainingID).Scan(
		&srv.ID,
		&srv.FLTrainingID,
		&srv.NodeName,
		&srv.PodName,
		&srv.CreatedAt,
		&srv.LastLogRead,
	)
	if err != nil {
		return FLTrainingServer{}, err
	}

	return srv, nil
}

func (s *FLTrainingServerStore) Create(ctx context.Context, srv FLTrainingServer) error {
	query := `
		INSERT INTO training_servers (
			fl_training_id,
			node_name,
			pod_name,
			last_log_read
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		srv.FLTrainingID,
		srv.NodeName,
		srv.PodName,
		srv.LastLogRead,
	).Scan(
		&srv.ID,
		&srv.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

// EnsureByFLTrainingID returns the server row for a training or creates a new one with the provided node/pod.
func (s *FLTrainingServerStore) EnsureByFLTrainingID(
	ctx context.Context,
	flTrainingID string,
	nodeName string,
	podName string,
) (FLTrainingServer, error) {
	srv, err := s.GetByFLTrainingID(ctx, flTrainingID)
	if err == nil {
		return srv, nil
	}
	if err != sql.ErrNoRows {
		return FLTrainingServer{}, err
	}

	newSrv := FLTrainingServer{
		FLTrainingID: flTrainingID,
		NodeName:     nodeName,
		PodName:      podName,
		LastLogRead:  time.Time{},
	}

	if err := s.Create(ctx, newSrv); err != nil {
		return FLTrainingServer{}, err
	}

	return s.GetByFLTrainingID(ctx, flTrainingID)
}

func (s *FLTrainingServerStore) UpdateLastLogRead(
	ctx context.Context,
	flTrainingID string,
	t time.Time,
) error {
	query := `
		UPDATE training_servers
		SET last_log_read = $2
		WHERE fl_training_id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, flTrainingID, t)
	return err
}
