package scenarios

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type CPUBurner struct {
	targetPercent int
	stopCh        chan struct{}
	running       atomic.Bool
	startTime     time.Time
	params        map[string]interface{}
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewCPUBurner() *CPUBurner {
	return &CPUBurner{
		stopCh: make(chan struct{}),
		params: make(map[string]interface{}),
	}
}

func (c *CPUBurner) Name() string {
	return "cpu_burner"
}

func (c *CPUBurner) Describe() string {
	return "Increases CPU usage to specified percentage"
}

func (c *CPUBurner) Start(ctx context.Context, params map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running.Load() {
		c.stop()
	}

	c.ctx, c.cancel = context.WithCancel(ctx)
	c.startTime = time.Now()
	c.params = params

	targetPercent := 50
	if tp, ok := params["target_percent"].(float64); ok {
		targetPercent = int(tp)
	} else if tp, ok := params["target_percent"].(int); ok {
		targetPercent = tp
	}
	c.targetPercent = targetPercent

	numCores := runtime.NumCPU()
	c.running.Store(true)

	for i := 0; i < numCores; i++ {
		go c.burnCPU()
	}

	return nil
}

func (c *CPUBurner) burnCPU() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.stopCh:
			return
		default:
			for j := 0; j < 1000000; j++ {
				_ = j * j
			}
			sleepTime := time.Duration((100-c.targetPercent)*10) * time.Microsecond
			if sleepTime > 0 {
				time.Sleep(sleepTime)
			}
		}
	}
}

func (c *CPUBurner) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.stop()
}

func (c *CPUBurner) stop() error {
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

func (c *CPUBurner) Status() ScenarioStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return ScenarioStatus{
		Running:   c.running.Load(),
		StartTime: c.startTime,
		Params:    c.params,
		Metrics: map[string]float64{
			"target_percent": float64(c.targetPercent),
			"num_cores":      float64(runtime.NumCPU()),
		},
	}
}
