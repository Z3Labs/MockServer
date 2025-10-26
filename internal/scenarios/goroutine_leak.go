package scenarios

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type GoroutineLeak struct {
	leakRate  int
	stopCh    chan struct{}
	running   atomic.Bool
	startTime time.Time
	params    map[string]interface{}
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewGoroutineLeak() *GoroutineLeak {
	return &GoroutineLeak{
		stopCh: make(chan struct{}),
		params: make(map[string]interface{}),
	}
}

func (g *GoroutineLeak) Name() string {
	return "goroutine_leak"
}

func (g *GoroutineLeak) Describe() string {
	return "Creates goroutines that never exit, causing goroutine leak"
}

func (g *GoroutineLeak) Start(ctx context.Context, params map[string]interface{}) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.running.Load() {
		g.stop()
	}

	g.ctx, g.cancel = context.WithCancel(ctx)
	g.startTime = time.Now()
	g.params = params

	leakRate := 100
	if lr, ok := params["goroutines_per_second"].(float64); ok {
		leakRate = int(lr)
	} else if lr, ok := params["goroutines_per_second"].(int); ok {
		leakRate = lr
	}
	g.leakRate = leakRate

	g.running.Store(true)
	go g.leakGoroutines()

	return nil
}

func (g *GoroutineLeak) leakGoroutines() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-g.ctx.Done():
			return
		case <-g.stopCh:
			return
		case <-ticker.C:
			for i := 0; i < g.leakRate; i++ {
				go func() {
					select {}
				}()
			}
		}
	}
}

func (g *GoroutineLeak) Stop() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.stop()
}

func (g *GoroutineLeak) stop() error {
	if !g.running.Load() {
		return nil
	}

	g.running.Store(false)
	if g.cancel != nil {
		g.cancel()
	}
	close(g.stopCh)
	g.stopCh = make(chan struct{})

	return nil
}

func (g *GoroutineLeak) Status() ScenarioStatus {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return ScenarioStatus{
		Running:   g.running.Load(),
		StartTime: g.startTime,
		Params:    g.params,
		Metrics: map[string]float64{
			"leak_rate":          float64(g.leakRate),
			"current_goroutines": float64(runtime.NumGoroutine()),
		},
	}
}
