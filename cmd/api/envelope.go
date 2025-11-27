package main

import (
	"encoding/json"
	"time"
)

// Envelope represents a single structured log event coming from a pod log line.
// The raw message from Kubernetes is a JSON object that is unmarshaled into this struct.
type Envelope struct {
	Event     string          `json:"event"`
	Payload   json.RawMessage `json:"payload"`
	PodName   string          `json:"-"`
	NodeName  string          `json:"-"`
	Component string          `json:"-"` // e.g. "clientapp" or "serverapp"
	Timestamp time.Time       `json:"-"` // timestamp parsed from Kubernetes log line
}

// Client-side events (component = "clientapp")

type ReadlinePayload struct {
	FLTrainingID string `json:"fl_training_id"`
	PartitionID  int    `json:"partition_id"`
	Text         string `json:"text"`
}

type SetStatePayload struct {
	FLTrainingID string `json:"fl_training_id"`
	PartitionID  int    `json:"partition_id"`
	State        string `json:"state"`
}

type SetCurrentServerRoundPayload struct {
	FLTrainingID string `json:"fl_training_id"`
	PartitionID  int    `json:"partition_id"` // may or may not be used
	ServerRound  int    `json:"server_round"`
}

type CreateTrainingGraphPayload struct {
	FLTrainingID string  `json:"fl_training_id"`
	PartitionID  int     `json:"partition_id"`
	ServerRound  int     `json:"server_round"`
	Optimizer    string  `json:"optimizer"`
	LearningRate float64 `json:"learning_rate"`
	NumEpochs    int     `json:"num_epochs"`
	BatchSize    int     `json:"batch_size"`
}

type AddOneEpochTrainingGraphPointPayload struct {
	FLTrainingID         string  `json:"fl_training_id"`
	PartitionID          int     `json:"partition_id"`
	ServerRound          int     `json:"server_round"`
	TrainedBatch         int     `json:"trained_batch"`
	CurrentEpoch         int     `json:"current_epoch"`
	TrainLoss            float64 `json:"train_loss"`
	ValLoss              float64 `json:"val_loss"`
	Accuracy             float64 `json:"accuracy"`
	EpochTrainingElapsed float64 `json:"epoch_training_elapsed_time"`
}

type CreateTestingGraphPayload struct {
	FLTrainingID string `json:"fl_training_id"`
	PartitionID  int    `json:"partition_id"`
}

type AddOneServerRoundTestingGraphPointPayload struct {
	FLTrainingID string  `json:"fl_training_id"`
	PartitionID  int     `json:"partition_id"`
	ServerRound  int     `json:"server_round"`
	Criterion    string  `json:"criterion"`
	BatchSize    int     `json:"batch_size"`
	TestLoss     float64 `json:"test_loss"`
	Accuracy     float64 `json:"accuracy"`
}

// Server-side events (component = "serverapp")

type CreateFLTrainingPayload struct {
	FLTrainingID string `json:"fl_training_id"`
	NumRounds    int    `json:"num_rounds"`
}

type ModelWeightsPayload struct {
	FLTrainingID string          `json:"fl_training_id"`
	ServerRound  int             `json:"server_round"`
	Layers       json.RawMessage `json:"layers"` // stored as JSONB; caller does not inspect it directly
}
