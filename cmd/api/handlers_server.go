package main

import (
	"context"
)

// handleModelWeights stores model weights payload for each server round (JSONB).
func (app *application) handleModelWeights(
	ctx context.Context,
	env Envelope,
	p ModelWeightsPayload,
) error {
	// Ensure training exists.
	if _, err := app.store.FLTrainings.Ensure(ctx, p.FLTrainingID); err != nil {
		return err
	}

	// Ensure training server row exists for this training.
	srv, err := app.store.TrainingServers.EnsureByFLTrainingID(
		ctx,
		p.FLTrainingID,
		env.NodeName,
		env.PodName,
	)
	if err != nil {
		return err
	}

	// Skip duplicate/out-of-order events.
	if app.shouldSkipByServerLastLogRead(srv, env.Timestamp, env.Event) {
		return nil
	}

	// Store JSONB payload as-is (includes fl_training_id, server_round, and layers).
	if err := app.store.FLModelWeights.Upsert(
		ctx,
		p.FLTrainingID,
		p.ServerRound,
		env.Payload,
	); err != nil {
		return err
	}

	app.logger.Infow("stored MODEL_WEIGHTS payload",
		"fl_training_id", p.FLTrainingID,
		"server_round", p.ServerRound,
	)

	// Update last_log_read marker for server.
	app.updateServerLastLogRead(ctx, srv, env.Timestamp)

	return nil
}

// handleCreateFLTraining is called when the server announces a new FL training
// or updates its configuration (e.g. total number of rounds).
func (app *application) handleCreateFLTraining(
	ctx context.Context,
	env Envelope,
	p CreateFLTrainingPayload,
) error {
	// Ensure training exists.
	tr, err := app.store.FLTrainings.Ensure(ctx, p.FLTrainingID)
	if err != nil {
		return err
	}

	// Ensure training server row exists.
	srv, err := app.store.TrainingServers.EnsureByFLTrainingID(
		ctx,
		p.FLTrainingID,
		env.NodeName,
		env.PodName,
	)
	if err != nil {
		return err
	}

	// Skip duplicate/out-of-order events.
	if app.shouldSkipByServerLastLogRead(srv, env.Timestamp, env.Event) {
		return nil
	}

	app.logger.Infow("create or update FL training from server",
		"fl_training_id", tr.FLTrainingID,
		"num_rounds", p.NumRounds,
	)

	// Update total number of server rounds based on server configuration.
	if err := app.store.FLTrainings.UpdateTotalServerRound(
		ctx,
		p.FLTrainingID,
		p.NumRounds,
	); err != nil {
		return err
	}

	// Update last_log_read marker for server.
	app.updateServerLastLogRead(ctx, srv, env.Timestamp)

	return nil
}
