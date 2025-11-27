package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrConflict          = errors.New("resource already exists")
	QueryTimeoutDuration = 5 * time.Second
)

type Storage struct {
	FLTrainings interface {
		GetAll(context.Context) ([]FLTraining, error)
		GetByFLTrainingID(context.Context, string) (FLTraining, error)
		Ensure(context.Context, string) (FLTraining, error)
		Create(context.Context, FLTraining) error
		UpdateCurrentServerRound(ctx context.Context, flTrainingID string, serverRound int) error
		UpdateTotalServerRound(ctx context.Context, flTrainingID string, totalServerRound int) error
	}

	FLTrainingClients interface {
		GetAll(context.Context) ([]FLTrainingClient, error)
		Create(context.Context, FLTrainingClient) error
		UpdateState(ctx context.Context, flTrainingID string, partitionID int, state string) error
		GetByFLTrainingIDAndPartitionID(ctx context.Context, flTrainingID string, partitionID int) (FLTrainingClient, error)
		EnsureByFLTrainingIDAndPartitionID(
			ctx context.Context,
			flTrainingID string,
			partitionID int,
			nodeName string,
			podName string,
			state string,
		) (FLTrainingClient, error)
		UpdateLastLogRead(ctx context.Context, flTrainingID string, partitionID int, t time.Time) error
	}

	TrainingServers interface {
		EnsureByFLTrainingID(ctx context.Context, flTrainingID, nodeName, podName string) (FLTrainingServer, error)
		UpdateLastLogRead(ctx context.Context, flTrainingID string, t time.Time) error
	}

	ClientLogs interface {
		Create(context.Context, ClientLog) error
		GetByClientID(context.Context, uuid.UUID) ([]ClientLog, error)
	}

	TrainingGraphs interface {
		Create(context.Context, TrainingGraph) error
		CreatePoint(context.Context, TrainingGraphPoint) error
		GetGraphIDsByClientID(context.Context, uuid.UUID) ([]uuid.UUID, error)
		GetGraphsByClientID(context.Context, uuid.UUID) ([]TrainingGraph, error)
		GetPointsByGraphID(context.Context, uuid.UUID) ([]TrainingGraphPoint, error)
	}

	TestingGraphs interface {
		Create(context.Context, TestingGraph) error
		CreatePoint(context.Context, TestingGraphPoint) error
		GetGraphIDsByClientID(context.Context, uuid.UUID) ([]uuid.UUID, error)
		GetGraphsByClientID(context.Context, uuid.UUID) ([]TestingGraph, error)
		GetPointsByGraphID(context.Context, uuid.UUID) ([]TestingGraphPoint, error)
	}

	FLModelWeights interface {
		Upsert(ctx context.Context, flTrainingID string, serverRound int, payload []byte) error
	}
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		FLTrainings:       NewFLTrainingStore(db),
		FLTrainingClients: NewFLTrainingClientStore(db),
		TrainingServers:   NewFLTrainingServerStore(db),
		ClientLogs:        NewClientLogStore(db),
		TrainingGraphs:    NewTrainingGraphStore(db),
		TestingGraphs:     NewTestingGraphStore(db),
		FLModelWeights:    NewFLModelWeightsStore(db),
	}
}
