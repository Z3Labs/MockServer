package scenarios

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type HealthCheckFailure struct {
	failureMode string
	statusCode  int
	failRate    float64
	stopCh      chan struct{}
	running     atomic.Bool
	startTime   time.Time
	params      map[string]interface{}
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewHealthCheckFailure() *HealthCheckFailure {
	return &HealthCheckFailure{
		stopCh:     make(chan struct{}),
		params:     make(map[string]interface{}),
		statusCode: 503,
	}
}

func (h *HealthCheckFailure) Name() string {
	return "health_check"
}

func (h *HealthCheckFailure) Describe() string {
	return "Controls health check endpoint to return failures"
}

func (h *HealthCheckFailure) Start(ctx context.Context, params map[string]interface{}) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.running.Load() {
		h.stop()
	}

	h.ctx, h.cancel = context.WithCancel(ctx)
	h.startTime = time.Now()
	h.params = params

	h.failureMode = "always"
	if fm, ok := params["failure_mode"].(string); ok {
		h.failureMode = fm
	}

	h.statusCode = 503
	if sc, ok := params["status_code"].(float64); ok {
		h.statusCode = int(sc)
	} else if sc, ok := params["status_code"].(int); ok {
		h.statusCode = sc
	}

	h.failRate = 0.5
	if fr, ok := params["fail_rate"].(float64); ok {
		h.failRate = fr
	}

	h.running.Store(true)

	return nil
}

func (h *HealthCheckFailure) Stop() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.stop()
}

func (h *HealthCheckFailure) stop() error {
	if !h.running.Load() {
		return nil
	}

	h.running.Store(false)
	if h.cancel != nil {
		h.cancel()
	}
	close(h.stopCh)
	h.stopCh = make(chan struct{})

	return nil
}

func (h *HealthCheckFailure) Status() ScenarioStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return ScenarioStatus{
		Running:   h.running.Load(),
		StartTime: h.startTime,
		Params:    h.params,
		Metrics: map[string]float64{
			"status_code": float64(h.statusCode),
			"fail_rate":   h.failRate,
		},
	}
}

func (h *HealthCheckFailure) ShouldFail() (bool, int, time.Duration) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if !h.running.Load() {
		return false, 200, 0
	}

	switch h.failureMode {
	case "always":
		return true, h.statusCode, 0
	case "intermittent":
		if rand.Float64() < h.failRate {
			return true, 503, 0
		}
		return false, 200, 0
	case "delayed":
		return false, 200, 10 * time.Second
	default:
		return false, 200, 0
	}
}
