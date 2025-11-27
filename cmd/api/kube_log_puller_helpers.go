package main

import (
	"fmt"
	"strings"
	"time"
)

// parseK8sTimestampLine splits a Kubernetes log line into timestamp and message.
// Expected format: "<RFC3339Nano timestamp> <json or text>".
func parseK8sTimestampLine(line string) (time.Time, string, error) {
	idx := strings.IndexByte(line, ' ')
	if idx == -1 {
		return time.Time{}, line, fmt.Errorf("invalid log line: no space")
	}

	tsStr := line[:idx]
	msg := strings.TrimSpace(line[idx+1:])

	ts, err := time.Parse(time.RFC3339Nano, tsStr)
	if err != nil {
		// return original message even if timestamp cannot be parsed
		return time.Time{}, msg, err
	}

	return ts, msg, nil
}
