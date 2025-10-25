package scenarios

import (
	"context"
	"time"
)

type Scenario interface {
	Name() string
	Start(ctx context.Context, params map[string]interface{}) error
	Stop() error
	Status() ScenarioStatus
	Describe() string
}

type ScenarioStatus struct {
	Running   bool                   `json:"running"`
	StartTime time.Time              `json:"start_time,omitempty"`
	Params    map[string]interface{} `json:"params"`
	Metrics   map[string]float64     `json:"metrics"`
}
