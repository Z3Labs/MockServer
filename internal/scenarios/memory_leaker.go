package scenarios

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type MemoryLeaker struct {
	leakedMemory [][]byte
	leakRateMB   int
	targetMB     int
	stopCh       chan struct{}
	running      atomic.Bool
	startTime    time.Time
	params       map[string]interface{}
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewMemoryLeaker() *MemoryLeaker {
	return &MemoryLeaker{
		leakedMemory: make([][]byte, 0),
		stopCh:       make(chan struct{}),
		params:       make(map[string]interface{}),
	}
}

func (m *MemoryLeaker) Name() string {
	return "memory_leaker"
}

func (m *MemoryLeaker) Describe() string {
	return "Continuously leaks memory at specified rate until target is reached"
}

func (m *MemoryLeaker) Start(ctx context.Context, params map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running.Load() {
		m.stop()
	}

	m.ctx, m.cancel = context.WithCancel(ctx)
	m.startTime = time.Now()
	m.params = params
	m.leakedMemory = make([][]byte, 0)

	targetMB := 1024
	if t, ok := params["target_mb"].(float64); ok {
		targetMB = int(t)
	} else if t, ok := params["target_mb"].(int); ok {
		targetMB = t
	}
	m.targetMB = targetMB

	leakRateMB := 10
	if lr, ok := params["leak_rate_mb"].(float64); ok {
		leakRateMB = int(lr)
	} else if lr, ok := params["leak_rate_mb"].(int); ok {
		leakRateMB = lr
	}
	m.leakRateMB = leakRateMB

	m.running.Store(true)
	go m.leakMemory()

	return nil
}

func (m *MemoryLeaker) leakMemory() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.mu.Lock()
			currentMB := len(m.leakedMemory) * m.leakRateMB
			if currentMB >= m.targetMB {
				m.mu.Unlock()
				return
			}

			chunk := make([]byte, m.leakRateMB*1024*1024)
			for i := range chunk {
				chunk[i] = byte(i % 256)
			}
			m.leakedMemory = append(m.leakedMemory, chunk)
			m.mu.Unlock()
		}
	}
}

func (m *MemoryLeaker) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stop()
}

func (m *MemoryLeaker) stop() error {
	if !m.running.Load() {
		return nil
	}

	m.running.Store(false)
	if m.cancel != nil {
		m.cancel()
	}
	close(m.stopCh)
	m.stopCh = make(chan struct{})
	m.leakedMemory = nil

	return nil
}

func (m *MemoryLeaker) Status() ScenarioStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	currentMB := len(m.leakedMemory) * m.leakRateMB

	return ScenarioStatus{
		Running:   m.running.Load(),
		StartTime: m.startTime,
		Params:    m.params,
		Metrics: map[string]float64{
			"current_mb":    float64(currentMB),
			"target_mb":     float64(m.targetMB),
			"leak_rate_mb":  float64(m.leakRateMB),
		},
	}
}
