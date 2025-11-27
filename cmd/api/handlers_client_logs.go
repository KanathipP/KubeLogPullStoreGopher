package main

import (
	"context"

	"github.com/KanathipP/KubeLogPullStoreGopher/internal/store"
)

// handleReadline stores a single log line for a client and advances last_log_read
// if the event is newer than what we have seen before.
func (app *application) handleReadline(
	ctx context.Context,
	env Envelope,
	p ReadlinePayload,
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
		"init", // initial state for a newly discovered client
	)
	if err != nil {
		return err
	}

	// Skip duplicate/out-of-order events.
	if app.shouldSkipByLastLogRead(client, env.Timestamp, env.Event) {
		return nil
	}

	// Write log entry.
	log := store.ClientLog{
		ClientID:       client.ID,
		Text:           p.Text,
		ClientOutputAt: env.Timestamp,
	}

	if err := app.store.ClientLogs.Create(ctx, log); err != nil {
		return err
	}

	// Update last_log_read marker.
	app.updateClientLastLogRead(ctx, client, env.Timestamp)

	return nil
}
