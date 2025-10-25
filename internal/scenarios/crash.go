package scenarios

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type CrashSimulator struct {
	crashDelay int
	stopCh     chan struct{}
	running    atomic.Bool
	startTime  time.Time
	params     map[string]interface{}
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewCrashSimulator() *CrashSimulator {
	return &CrashSimulator{
		stopCh: make(chan struct{}),
		params: make(map[string]interface{}),
	}
}

func (c *CrashSimulator) Name() string {
	return "crash"
}

func (c *CrashSimulator) Describe() string {
	return "Simulates service crash after specified delay"
}

func (c *CrashSimulator) Start(ctx context.Context, params map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running.Load() {
		c.stop()
	}

	c.ctx, c.cancel = context.WithCancel(ctx)
	c.startTime = time.Now()
	c.params = params

	crashDelay := 10
	if cd, ok := params["crash_delay"].(float64); ok {
		crashDelay = int(cd)
	} else if cd, ok := params["crash_delay"].(int); ok {
		crashDelay = cd
	}
	c.crashDelay = crashDelay

	c.running.Store(true)
	go c.scheduleCrash()

	return nil
}

func (c *CrashSimulator) scheduleCrash() {
	timer := time.NewTimer(time.Duration(c.crashDelay) * time.Second)
	defer timer.Stop()

	select {
	case <-c.ctx.Done():
		return
	case <-c.stopCh:
		return
	case <-timer.C:
		os.Exit(1)
	}
}

func (c *CrashSimulator) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.stop()
}

func (c *CrashSimulator) stop() error {
	if !c.running.Load() {
		return nil
	}

	c.running.Store(false)
	if c.cancel != nil {
		c.cancel()
	}
	close(c.stopCh)
	c.stopCh = make(chan struct{})

	return nil
}

func (c *CrashSimulator) Status() ScenarioStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return ScenarioStatus{
		Running:   c.running.Load(),
		StartTime: c.startTime,
		Params:    c.params,
		Metrics: map[string]float64{
			"crash_delay": float64(c.crashDelay),
		},
	}
}
