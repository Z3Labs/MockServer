package scenarios

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type DependencyFailure struct {
	failureType string
	stopCh      chan struct{}
	running     atomic.Bool
	startTime   time.Time
	params      map[string]interface{}
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewDependencyFailure() *DependencyFailure {
	return &DependencyFailure{
		stopCh: make(chan struct{}),
		params: make(map[string]interface{}),
	}
}

func (d *DependencyFailure) Name() string {
	return "dependency"
}

func (d *DependencyFailure) Describe() string {
	return "Simulates dependency service failures (timeout, error, slow response)"
}

func (d *DependencyFailure) Start(ctx context.Context, params map[string]interface{}) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.running.Load() {
		d.stop()
	}

	d.ctx, d.cancel = context.WithCancel(ctx)
	d.startTime = time.Now()
	d.params = params

	d.failureType = "timeout"
	if ft, ok := params["failure_type"].(string); ok {
		d.failureType = ft
	}

	d.running.Store(true)

	return nil
}

func (d *DependencyFailure) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.stop()
}

func (d *DependencyFailure) stop() error {
	if !d.running.Load() {
		return nil
	}

	d.running.Store(false)
	if d.cancel != nil {
		d.cancel()
	}
	close(d.stopCh)
	d.stopCh = make(chan struct{})

	return nil
}

func (d *DependencyFailure) Status() ScenarioStatus {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return ScenarioStatus{
		Running:   d.running.Load(),
		StartTime: d.startTime,
		Params:    d.params,
		Metrics:   map[string]float64{},
	}
}

func (d *DependencyFailure) GetFailureType() (string, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if !d.running.Load() {
		return "", false
	}
	return d.failureType, true
}
