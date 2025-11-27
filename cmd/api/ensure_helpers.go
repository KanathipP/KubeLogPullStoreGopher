package main

import (
	"context"
	"fmt"

	"github.com/KanathipP/KubeLogPullStoreGopher/internal/store"
)

// ensureTraining guarantees that a training row exists for the given ID.
func (app *application) ensureTraining(ctx context.Context, flTrainingID string) (store.FLTraining, error) {
	t, err := app.store.FLTrainings.Ensure(ctx, flTrainingID)
	if err != nil {
		return store.FLTraining{}, fmt.Errorf("ensureTraining(%s): %w", flTrainingID, err)
	}
	return t, nil
}

// ensureTrainingAndClient ensures both FLTraining and FLTrainingClient exist.
// If you do not use this helper anywhere, you can safely remove this file.
func (app *application) ensureTrainingAndClient(
	ctx context.Context,
	flTrainingID string,
	partitionID int,
	nodeName string,
	podName string,
) (store.FLTraining, store.FLTrainingClient, error) {
	t, err := app.ensureTraining(ctx, flTrainingID)
	if err != nil {
		return store.FLTraining{}, store.FLTrainingClient{}, err
	}

	c, err := app.store.FLTrainingClients.EnsureByFLTrainingIDAndPartitionID(
		ctx,
		flTrainingID,
		partitionID,
		nodeName,
		podName,
		"unknown", // initial state, expected to be updated later
	)
	if err != nil {
		return store.FLTraining{}, store.FLTrainingClient{}, err
	}

	return t, c, nil
}
