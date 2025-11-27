package main

import (
	"context"
	"fmt"

	"github.com/KanathipP/KubeLogPullStoreGopher/internal/store"

	"github.com/google/uuid"
)

// ensureTestingGraph returns the single testing_graph for a client, creating it if needed.
func (app *application) ensureTestingGraph(
	ctx context.Context,
	clientID uuid.UUID,
) (store.TestingGraph, error) {
	graphs, err := app.store.TestingGraphs.GetGraphsByClientID(ctx, clientID)
	if err != nil {
		return store.TestingGraph{}, fmt.Errorf("GetGraphsByClientID: %w", err)
	}

	if len(graphs) > 0 {
		return graphs[0], nil
	}

	g := store.TestingGraph{
		ClientID: clientID,
	}

	if err := app.store.TestingGraphs.Create(ctx, g); err != nil {
		return store.TestingGraph{}, fmt.Errorf("Create testing_graph: %w", err)
	}

	graphs, err = app.store.TestingGraphs.GetGraphsByClientID(ctx, clientID)
	if err != nil {
		return store.TestingGraph{}, fmt.Errorf("GetGraphsByClientID after create: %w", err)
	}
	if len(graphs) == 0 {
		return store.TestingGraph{}, fmt.Errorf("ensureTestingGraph: not found after create")
	}

	return graphs[0], nil
}

// handleCreateTestingGraph ensures the testing graph exists for a client.
// The graph itself does not carry parameters; it is essentially a container for points.
func (app *application) handleCreateTestingGraph(
	ctx context.Context,
	env Envelope,
	p CreateTestingGraphPayload,
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
		"test",
	)
	if err != nil {
		return err
	}

	// Skip duplicates/out-of-order events.
	if app.shouldSkipByLastLogRead(client, env.Timestamp, env.Event) {
		return nil
	}

	// Ensure a single testing graph for this client.
	if _, err := app.ensureTestingGraph(ctx, client.ID); err != nil {
		return err
	}

	// Update last_log_read marker.
	app.updateClientLastLogRead(ctx, client, env.Timestamp)

	return nil
}

// handleAddOneServerRoundTestingGraphPoint inserts a new testing_graph_point for a client.
func (app *application) handleAddOneServerRoundTestingGraphPoint(
	ctx context.Context,
	env Envelope,
	p AddOneServerRoundTestingGraphPointPayload,
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
		"test",
	)
	if err != nil {
		return err
	}

	// Skip duplicates/out-of-order events.
	if app.shouldSkipByLastLogRead(client, env.Timestamp, env.Event) {
		return nil
	}

	// Ensure testing graph exists.
	graph, err := app.ensureTestingGraph(ctx, client.ID)
	if err != nil {
		return err
	}

	// Insert a new point for this server_round.
	point := store.TestingGraphPoint{
		GraphID:     graph.ID,
		ServerRound: p.ServerRound,
		Criterion:   p.Criterion,
		BatchSize:   p.BatchSize,
		TestLoss:    p.TestLoss,
		Accuracy:    p.Accuracy,
	}

	if err := app.store.TestingGraphs.CreatePoint(ctx, point); err != nil {
		return err
	}

	// Update last_log_read marker.
	app.updateClientLastLogRead(ctx, client, env.Timestamp)

	return nil
}
