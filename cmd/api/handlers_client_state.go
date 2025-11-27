package main

import (
	"context"
)

// handleSetState updates the state field of a client for a given training/partition.
func (app *application) handleSetState(
	ctx context.Context,
	env Envelope,
	p SetStatePayload,
) error {
	// Ensure training exists.
	if _, err := app.store.FLTrainings.Ensure(ctx, p.FLTrainingID); err != nil {
		return err
	}

	// Ensure client exists (new client will start in the state from this event).
	client, err := app.store.FLTrainingClients.EnsureByFLTrainingIDAndPartitionID(
		ctx,
		p.FLTrainingID,
		p.PartitionID,
		env.NodeName,
		env.PodName,
		p.State,
	)
	if err != nil {
		return err
	}

	// Skip duplicate/out-of-order events.
	if app.shouldSkipByLastLogRead(client, env.Timestamp, env.Event) {
		return nil
	}

	// Update state in the DB.
	if err := app.store.FLTrainingClients.UpdateState(
		ctx,
		p.FLTrainingID,
		p.PartitionID,
		p.State,
	); err != nil {
		return err
	}

	app.logger.Infow("client state updated",
		"fl_training_id", p.FLTrainingID,
		"partition_id", p.PartitionID,
		"state", p.State,
	)

	// Update last_log_read marker.
	app.updateClientLastLogRead(ctx, client, env.Timestamp)

	return nil
}

// handleSetCurrentServerRound updates the current_server_round for a training.
func (app *application) handleSetCurrentServerRound(
	ctx context.Context,
	env Envelope,
	p SetCurrentServerRoundPayload,
) error {
	// Ensure training exists.
	tr, err := app.store.FLTrainings.Ensure(ctx, p.FLTrainingID)
	if err != nil {
		return err
	}

	// Ensure client exists (state "init" for this type of event by default).
	client, err := app.store.FLTrainingClients.EnsureByFLTrainingIDAndPartitionID(
		ctx,
		p.FLTrainingID,
		p.PartitionID,
		env.NodeName,
		env.PodName,
		"init",
	)
	if err != nil {
		return err
	}

	// Skip duplicate/out-of-order events.
	if app.shouldSkipByLastLogRead(client, env.Timestamp, env.Event) {
		return nil
	}

	// Update current_server_round with monotonic semantics (never decrease).
	if err := app.store.FLTrainings.UpdateCurrentServerRound(
		ctx,
		p.FLTrainingID,
		p.ServerRound,
	); err != nil {
		return err
	}

	app.logger.Infow("current server round updated",
		"fl_training_id", tr.FLTrainingID,
		"server_round", p.ServerRound,
	)

	// Update last_log_read marker for the client that emitted this event.
	app.updateClientLastLogRead(ctx, client, env.Timestamp)

	return nil
}
