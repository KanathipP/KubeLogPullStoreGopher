package main

import (
	"bufio"
	"context"
	"encoding/json"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (app *application) kubeLogPuller(
	ctx context.Context,
	out chan<- Envelope,
) {
	// let's set it CronJob every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	lastRead := make(map[string]time.Time)

	for {
		select {
		case <-ctx.Done():
			app.logger.Info("log puller context canceled")
			return

		case <-ticker.C:
			namespace := app.config.podFilter.namespace
			labelSelector := app.config.podFilter.labelSelector

			pods, err := app.kube.Pods(namespace).List(ctx, metav1.ListOptions{
				LabelSelector: labelSelector,
			})
			if err != nil {
				app.logger.Errorw("failed to list pod logs", "namespace", namespace, "labelSelector", labelSelector)
			}

			for _, p := range pods.Items {
				pod := p
				podName := pod.Name

				since := lastRead[podName]

				app.logger.Infow("polling pod logs",
					"pod", podName,
					"node", pod.Spec.NodeName,
					"since", since,
				)

				newLast, err := app.kubeLogPull(ctx, pod, namespace, since, out)
				if err != nil && ctx.Err() == nil {
					app.logger.Infow("error while streaming pod logs (will retry next tick)",
						"pod", podName,
						"error", err,
					)
				}

				// update lastRead
				if !newLast.IsZero() && newLast.After(lastRead[podName]) {
					lastRead[podName] = newLast
				}
			}
		}
	}
}

func (app *application) kubeLogPull(
	ctx context.Context,
	pod corev1.Pod,
	namespace string,
	since time.Time, out chan<- Envelope,
) (time.Time, error) {
	// get pod name and node name from corev1.Pod
	podName := pod.Name
	nodeName := pod.Spec.NodeName

	// set log options
	opts := corev1.PodLogOptions{
		Timestamps: true,
	}

	// first pod pulling `since` will always be zero so it will send all the logs
	// if it's not zero, we will only pull the log from `since`
	if !since.IsZero() {
		t := metav1.NewTime(since)
		opts.SinceTime = &t
	}

	// setup connection
	podLogsConnection := app.kube.Pods(namespace).GetLogs(podName, &opts)

	logStream, err := podLogsConnection.Stream(ctx)
	if err != nil {
		app.logger.Errorw("failed to open pod log stream", "pod", podName, "error", err)
		return since, err
	}
	defer logStream.Close()

	scanner := bufio.NewScanner(logStream)

	lastTS := since

	for scanner.Scan() {
		// prevents canceled when reading
		select {
		case <-ctx.Done():
			return lastTS, ctx.Err()
		default:
		}

		line := scanner.Text()

		app.logger.Debugw("logging raw pod line",
			"pod", podName,
			"line", line,
		)

		ts, msg, timestampParseErr := parseK8sTimestampLine(line)
		if timestampParseErr != nil {
			continue
		} else {
			// prevents duplicate
			if !since.IsZero() && (ts.Before(since) || ts.Equal(since)) {
				continue
			}

			// set new lastTS (because the new one is the last one)
			if ts.After(lastTS) {
				lastTS = ts
			}
		}

		msg = strings.TrimSpace(msg)
		if msg == "" {
			continue
		}

		// JSON format will always starts with `{`
		if !strings.HasPrefix(msg, "{") {
			continue
		}

		var env Envelope
		if err := json.Unmarshal([]byte(msg), &env); err != nil {
			app.logger.Debugw("failed to unmarshal json",
				"pod", podName,
				"payload", msg,
				"err", err,
			)
			continue
		}

		// add kubernetes metadata in envelope
		env.PodName = podName
		env.NodeName = nodeName

		if pod.Labels != nil {
			if c, ok := pod.Labels["component"]; ok {
				env.Component = c
			}
		}
		if timestampParseErr == nil {
			env.Timestamp = ts
		}

		app.logger.Infow("parsed event",
			"env", env)

		out <- env
	}

	if err := scanner.Err(); err != nil && ctx.Err() == nil {
		app.logger.Errorw("read log error", "pod", podName, "err", err)
		return lastTS, err
	}

	return lastTS, nil
}
