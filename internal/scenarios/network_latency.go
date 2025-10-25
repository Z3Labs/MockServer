package scenarios

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type NetworkLatency struct {
	latencyMs int
	stopCh    chan struct{}
	running   atomic.Bool
	startTime time.Time
	params    map[string]interface{}
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewNetworkLatency() *NetworkLatency {
	return &NetworkLatency{
		stopCh: make(chan struct{}),
		params: make(map[string]interface{}),
	}
}

func (n *NetworkLatency) Name() string {
	return "network_latency"
}

func (n *NetworkLatency) Describe() string {
	return "Adds specified latency to HTTP requests"
}

func (n *NetworkLatency) Start(ctx context.Context, params map[string]interface{}) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.running.Load() {
		n.stop()
	}

	n.ctx, n.cancel = context.WithCancel(ctx)
	n.startTime = time.Now()
	n.params = params

	latencyMs := 100
	if lm, ok := params["latency_ms"].(float64); ok {
		latencyMs = int(lm)
	} else if lm, ok := params["latency_ms"].(int); ok {
		latencyMs = lm
	}
	n.latencyMs = latencyMs

	n.running.Store(true)

	return nil
}

func (n *NetworkLatency) Stop() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.stop()
}

func (n *NetworkLatency) stop() error {
	if !n.running.Load() {
		return nil
	}

	n.running.Store(false)
	if n.cancel != nil {
		n.cancel()
	}
	close(n.stopCh)
	n.stopCh = make(chan struct{})

	return nil
}

func (n *NetworkLatency) Status() ScenarioStatus {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return ScenarioStatus{
		Running:   n.running.Load(),
		StartTime: n.startTime,
		Params:    n.params,
		Metrics: map[string]float64{
			"latency_ms": float64(n.latencyMs),
		},
	}
}

func (n *NetworkLatency) GetLatency() time.Duration {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if !n.running.Load() {
		return 0
	}
	return time.Duration(n.latencyMs) * time.Millisecond
}
