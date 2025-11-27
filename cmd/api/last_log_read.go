package main

import (
	"context"
	"time"

	"github.com/KanathipP/KubeLogPullStoreGopher/internal/store"
)

// shouldSkipByLastLogRead decides whether to skip a client-side event
// based on the last seen timestamp for that client.
func (app *application) shouldSkipByLastLogRead(
	client store.FLTrainingClient,
	ts time.Time,
	event string,
) bool {
	if ts.IsZero() || client.LastLogRead.IsZero() {
		return false
	}

	// If the timestamp is not strictly newer than last_log_read, skip it.
	if !ts.After(client.LastLogRead) {
		app.logger.Debugw("skip client event older than last_log_read",
			"event", event,
			"client_id", client.ID,
			"fl_training_id", client.FLTrainingID,
			"partition_id", client.PartitionID,
			"ts", ts,
			"last_log_read", client.LastLogRead,
		)
		return true
	}

	return false
}

// updateClientLastLogRead updates last_log_read for a client if the timestamp is newer.
func (app *application) updateClientLastLogRead(
	ctx context.Context,
	client store.FLTrainingClient,
	ts time.Time,
) {
	if ts.IsZero() {
		return
	}

	if !ts.After(client.LastLogRead) {
		return
	}

	if err := app.store.FLTrainingClients.UpdateLastLogRead(
		ctx,
		client.FLTrainingID,
		client.PartitionID,
		ts,
	); err != nil {
		app.logger.Warnw("failed to update client last_log_read",
			"error", err,
			"client_id", client.ID,
			"fl_training_id", client.FLTrainingID,
			"partition_id", client.PartitionID,
		)
	}
}

// shouldSkipByServerLastLogRead decides whether to skip a server-side event
// based on the last seen timestamp for that training server.
func (app *application) shouldSkipByServerLastLogRead(
	srv store.FLTrainingServer,
	ts time.Time,
	event string,
) bool {
	if ts.IsZero() || srv.LastLogRead.IsZero() {
		return false
	}

	if !ts.After(srv.LastLogRead) {
		app.logger.Debugw("skip server event older than last_log_read",
			"event", event,
			"fl_training_id", srv.FLTrainingID,
			"ts", ts,
			"last_log_read", srv.LastLogRead,
		)
		return true
	}

	return false
}

// updateServerLastLogRead updates last_log_read for a training server if the timestamp is newer.
func (app *application) updateServerLastLogRead(
	ctx context.Context,
	srv store.FLTrainingServer,
	ts time.Time,
) {
	if ts.IsZero() {
		return
	}

	if !ts.After(srv.LastLogRead) {
		return
	}

	if err := app.store.TrainingServers.UpdateLastLogRead(
		ctx,
		srv.FLTrainingID,
		ts,
	); err != nil {
		app.logger.Warnw("failed to update server last_log_read",
			"error", err,
			"fl_training_id", srv.FLTrainingID,
		)
	}
}
