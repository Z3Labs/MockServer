package scenarios

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type DiskIO struct {
	writeRateMB int
	filePath    string
	stopCh      chan struct{}
	running     atomic.Bool
	startTime   time.Time
	params      map[string]interface{}
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewDiskIO() *DiskIO {
	return &DiskIO{
		stopCh:   make(chan struct{}),
		params:   make(map[string]interface{}),
		filePath: "/tmp/mock-server-io-test",
	}
}

func (d *DiskIO) Name() string {
	return "disk_io"
}

func (d *DiskIO) Describe() string {
	return "Generates high disk IO by writing data at specified rate"
}

func (d *DiskIO) Start(ctx context.Context, params map[string]interface{}) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.running.Load() {
		d.stop()
	}

	d.ctx, d.cancel = context.WithCancel(ctx)
	d.startTime = time.Now()
	d.params = params

	writeRateMB := 50
	if wr, ok := params["write_rate_mb"].(float64); ok {
		writeRateMB = int(wr)
	} else if wr, ok := params["write_rate_mb"].(int); ok {
		writeRateMB = wr
	}
	d.writeRateMB = writeRateMB

	d.running.Store(true)
	go d.performIO()

	return nil
}

func (d *DiskIO) performIO() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	data := make([]byte, d.writeRateMB*1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	counter := 0
	for {
		select {
		case <-d.ctx.Done():
			return
		case <-d.stopCh:
			return
		case <-ticker.C:
			filePath := fmt.Sprintf("%s-%d", d.filePath, counter)
			f, err := os.Create(filePath)
			if err != nil {
				continue
			}
			f.Write(data)
			f.Sync()
			f.Close()
			os.Remove(filePath)
			counter++
		}
	}
}

func (d *DiskIO) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.stop()
}

func (d *DiskIO) stop() error {
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

func (d *DiskIO) Status() ScenarioStatus {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return ScenarioStatus{
		Running:   d.running.Load(),
		StartTime: d.startTime,
		Params:    d.params,
		Metrics: map[string]float64{
			"write_rate_mb": float64(d.writeRateMB),
		},
	}
}
