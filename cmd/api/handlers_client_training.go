package main

import (
	"context"
	"fmt"

	"github.com/KanathipP/KubeLogPullStoreGopher/internal/store"

	"github.com/google/uuid"
)

// ensureTrainingGraphForRound returns a training_graph for (client, server_round),
// creating a minimal one if it does not yet exist.
func (app *application) ensureTrainingGraphForRound(
	ctx context.Context,
	clientID uuid.UUID,
	serverRound int,
) (store.TrainingGraph, error) {
	graphs, err := app.store.TrainingGraphs.GetGraphsByClientID(ctx, clientID)
	if err != nil {
		return store.TrainingGraph{}, fmt.Errorf("GetGraphsByClientID: %w", err)
	}

	for _, g := range graphs {
		if g.ServerRound == serverRound {
			return g, nil
		}
	}

	g := store.TrainingGraph{
		ClientID:    clientID,
		ServerRound: serverRound,
	}

	if err := app.store.TrainingGraphs.Create(ctx, g); err != nil {
		return store.TrainingGraph{}, fmt.Errorf("Create training_graph: %w", err)
	}

	graphs, err = app.store.TrainingGraphs.GetGraphsByClientID(ctx, clientID)
	if err != nil {
		return store.TrainingGraph{}, fmt.Errorf("GetGraphsByClientID after create: %w", err)
	}
	for _, g2 := range graphs {
		if g2.ServerRound == serverRound {
			return g2, nil
		}
	}

	return store.TrainingGraph{}, fmt.Errorf("ensureTrainingGraphForRound: not found after create")
}

// handleCreateTrainingGraph creates a training_graph configuration for a given server round.
func (app *application) handleCreateTrainingGraph(
	ctx context.Context,
	env Envelope,
	p CreateTrainingGraphPayload,
) error {
	// Ensure training exists.
	if _, err := app.store.FLTrainings.Ensure(ctx, p.FLTrainingID); err != nil {
		return err
	}

	// Ensure client exists (state "train" for training events).
	client, err := app.store.FLTrainingClients.EnsureByFLTrainingIDAndPartitionID(
		ctx,
		p.FLTrainingID,
		p.PartitionID,
		env.NodeName,
		env.PodName,
		"train",
	)
	if err != nil {
		return err
	}

	// Skip duplicates/out-of-order events.
	if app.shouldSkipByLastLogRead(client, env.Timestamp, env.Event) {
		return nil
	}

	// Create a training graph for this client and round with full configuration.
	g := store.TrainingGraph{
		ClientID:     client.ID,
		ServerRound:  p.ServerRound,
		Optimizer:    p.Optimizer,
		LearningRate: p.LearningRate,
		NumEpochs:    p.NumEpochs,
		BatchSize:    p.BatchSize,
	}

	if err := app.store.TrainingGraphs.Create(ctx, g); err != nil {
		return err
	}

	// Update last_log_read marker.
	app.updateClientLastLogRead(ctx, client, env.Timestamp)

	return nil
}

// handleAddOneEpochTrainingGraphPoint adds a single epoch point into the training graph.
func (app *application) handleAddOneEpochTrainingGraphPoint(
	ctx context.Context,
	env Envelope,
	p AddOneEpochTrainingGraphPointPayload,
) error {
	// Ensure training exists.
	if _, err := app.store.FLTrainings.Ensure(ctx, p.FLTrainingID); err != nil {
		return err
	}

	// Ensure client exists.
	client, err := app.store.FLTrainingClients.EnsureByFLTrainingIDAndPartitionID(
		ctx,
		p.FLTrainingID,
		p.PartitionID,
		env.NodeName,
		env.PodName,
		"train",
	)
	if err != nil {
		return err
	}

	// Skip duplicates/out-of-order events.
	if app.shouldSkipByLastLogRead(client, env.Timestamp, env.Event) {
		return nil
	}

	// Ensure training graph exists for this server round.
	graph, err := app.ensureTrainingGraphForRound(ctx, client.ID, p.ServerRound)
	if err != nil {
		return err
	}

	// Insert a new graph point.
	point := store.TrainingGraphPoint{
		GraphID:          graph.ID,
		CurrentEpoch:     p.CurrentEpoch,
		TrainedBatch:     p.TrainedBatch,
		TrainLoss:        p.TrainLoss,
		ValLoss:          p.ValLoss,
		Accuracy:         p.Accuracy,
		EpochElapsedTime: p.EpochTrainingElapsed,
	}

	if err := app.store.TrainingGraphs.CreatePoint(ctx, point); err != nil {
		return err
	}

	// Update last_log_read.
	app.updateClientLastLogRead(ctx, client, env.Timestamp)

	return nil
}
