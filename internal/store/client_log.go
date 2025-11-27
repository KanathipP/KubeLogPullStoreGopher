package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type ClientLog struct {
	ID             uuid.UUID `json:"id"`
	ClientID       uuid.UUID `json:"client_id"`
	Text           string    `json:"text"`
	ClientOutputAt time.Time `json:"client_output_at"`
	CreatedAt      time.Time `json:"created_at"`
}

type ClientLogStore struct {
	db *sql.DB
}

func NewClientLogStore(db *sql.DB) *ClientLogStore {
	return &ClientLogStore{db: db}
}

func (s *ClientLogStore) Create(ctx context.Context, log ClientLog) error {
	query := `
		INSERT INTO client_logs (client_id, text, client_output_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		log.ClientID,
		log.Text,
		log.ClientOutputAt,
	).Scan(
		&log.ID,
		&log.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *ClientLogStore) GetByClientID(ctx context.Context, clientID uuid.UUID) ([]ClientLog, error) {
	query := `
		SELECT 
			id,
			client_id,
			text,
			client_output_at,
			created_at
		FROM client_logs
		WHERE client_id = $1
		ORDER BY client_output_at ASC
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []ClientLog

	for rows.Next() {
		var l ClientLog
		err := rows.Scan(
			&l.ID,
			&l.ClientID,
			&l.Text,
			&l.ClientOutputAt,
			&l.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		logs = append(logs, l)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}
