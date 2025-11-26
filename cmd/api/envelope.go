package main

import (
	"encoding/json"
	"time"
)

type Envelope struct {
	Event     string          `json:"event"`
	Payload   json.RawMessage `json:"payload"`
	PodName   string          `json:"-"`
	NodeName  string          `json:"-"`
	Component string          `json:"-"`
	Timestamp time.Time       `json:"-"`
}
