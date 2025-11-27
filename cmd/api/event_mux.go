package main

import (
	"context"
	"encoding/json"
	"fmt"
)

// eventMux routes events based on the "component" label.
func (app *application) eventMux(env Envelope) error {
	switch env.Component {
	case "clientapp":
		return app.clientEventMux(env)
	case "serverapp":
		return app.serverEventMux(env)
	default:
		app.logger.Debugw("unknown component, skipping event",
			"component", env.Component,
			"event", env.Event,
			"pod", env.PodName,
			"node", env.NodeName,
		)
		return nil
	}
}

// clientEventMux routes client-side events.
func (app *application) clientEventMux(env Envelope) error {
	ctx := context.Background()

	app.logger.Infow("handling CLIENT event",
		"event", env.Event,
		"pod", env.PodName,
		"node", env.NodeName,
		"ts", env.Timestamp,
	)

	switch env.Event {
	case "READLINE":
		var p ReadlinePayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("READLINE unmarshal: %w", err)
		}
		return app.handleReadline(ctx, env, p)

	case "SETSTATE":
		var p SetStatePayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("SETSTATE unmarshal: %w", err)
		}
		return app.handleSetState(ctx, env, p)

	case "CREATE_TRAINING_GRAPH":
		var p CreateTrainingGraphPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("CREATE_TRAINING_GRAPH unmarshal: %w", err)
		}
		return app.handleCreateTrainingGraph(ctx, env, p)

	case "ADD_ONE_EPOCH_TRAINING_GRAPH_POINT":
		var p AddOneEpochTrainingGraphPointPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("ADD_ONE_EPOCH_TRAINING_GRAPH_POINT unmarshal: %w", err)
		}
		return app.handleAddOneEpochTrainingGraphPoint(ctx, env, p)

	case "CREATE_TESTING_GRAPH":
		var p CreateTestingGraphPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("CREATE_TESTING_GRAPH unmarshal: %w", err)
		}
		return app.handleCreateTestingGraph(ctx, env, p)

	case "ADD_ONE_SERVER_ROUND_TESTING_GRAPH_POINT":
		var p AddOneServerRoundTestingGraphPointPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("ADD_ONE_SERVER_ROUND_TESTING_GRAPH_POINT unmarshal: %w", err)
		}
		return app.handleAddOneServerRoundTestingGraphPoint(ctx, env, p)

	case "SET_CURRENT_SERVER_ROUND":
		var p SetCurrentServerRoundPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("SET_CURRENT_SERVER_ROUND unmarshal: %w", err)
		}
		return app.handleSetCurrentServerRound(ctx, env, p)

	default:
		app.logger.Warnw("unknown client event type",
			"event", env.Event,
			"payload", string(env.Payload),
			"pod", env.PodName,
			"node", env.NodeName,
		)
		return nil
	}
}

// serverEventMux routes server-side events.
func (app *application) serverEventMux(env Envelope) error {
	ctx := context.Background()

	app.logger.Infow("handling SERVER event",
		"event", env.Event,
		"pod", env.PodName,
		"node", env.NodeName,
		"ts", env.Timestamp,
	)

	switch env.Event {
	case "CREATE_FL_TRAINING":
		var p CreateFLTrainingPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("CREATE_FL_TRAINING unmarshal: %w", err)
		}
		return app.handleCreateFLTraining(ctx, env, p)

	case "MODEL_WEIGHTS":
		var p ModelWeightsPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("MODEL_WEIGHTS unmarshal: %w", err)
		}
		return app.handleModelWeights(ctx, env, p)

	default:
		app.logger.Warnw("unknown server event type",
			"event", env.Event,
			"payload", string(env.Payload),
			"pod", env.PodName,
			"node", env.NodeName,
		)
		return nil
	}
}
