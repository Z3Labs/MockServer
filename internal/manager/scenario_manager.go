package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Z3Labs/MockServer/internal/scenarios"
)

type ScenarioManager struct {
	scenarios      map[string]scenarios.Scenario
	currentSession *ScenarioSession
	mu             sync.RWMutex
}

type ScenarioSession struct {
	SessionId     string
	Scenarios     []string
	StartTime     time.Time
	CancelFunc    context.CancelFunc
	RecoveryTimer *time.Timer
}

type ScenarioInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Running     bool   `json:"running"`
}

func NewScenarioManager() *ScenarioManager {
	sm := &ScenarioManager{
		scenarios: make(map[string]scenarios.Scenario),
	}

	sm.registerScenarios()

	return sm
}

func (sm *ScenarioManager) registerScenarios() {
	sm.Register(scenarios.NewCPUBurner())
	sm.Register(scenarios.NewMemoryLeaker())
	sm.Register(scenarios.NewNetworkLatency())
	sm.Register(scenarios.NewHealthCheckFailure())
	sm.Register(scenarios.NewGoroutineLeak())
	sm.Register(scenarios.NewDiskIO())
	sm.Register(scenarios.NewCrashSimulator())
	sm.Register(scenarios.NewDependencyFailure())
}

func (sm *ScenarioManager) Register(scenario scenarios.Scenario) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.scenarios[scenario.Name()] = scenario
}

func (sm *ScenarioManager) Start(ctx context.Context, name string, params map[string]interface{}) error {
	sm.mu.RLock()
	scenario, ok := sm.scenarios[name]
	sm.mu.RUnlock()

	if !ok {
		return fmt.Errorf("scenario %s not found", name)
	}

	return scenario.Start(ctx, params)
}

func (sm *ScenarioManager) Stop(name string) error {
	sm.mu.RLock()
	scenario, ok := sm.scenarios[name]
	sm.mu.RUnlock()

	if !ok {
		return fmt.Errorf("scenario %s not found", name)
	}

	return scenario.Stop()
}

func (sm *ScenarioManager) Status(name string) (scenarios.ScenarioStatus, error) {
	sm.mu.RLock()
	scenario, ok := sm.scenarios[name]
	sm.mu.RUnlock()

	if !ok {
		return scenarios.ScenarioStatus{}, fmt.Errorf("scenario %s not found", name)
	}

	return scenario.Status(), nil
}

func (sm *ScenarioManager) List() []ScenarioInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	list := make([]ScenarioInfo, 0, len(sm.scenarios))
	for name, scenario := range sm.scenarios {
		status := scenario.Status()
		list = append(list, ScenarioInfo{
			Name:        name,
			Description: scenario.Describe(),
			Running:     status.Running,
		})
	}

	return list
}

func (sm *ScenarioManager) GetScenario(name string) (scenarios.Scenario, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	scenario, ok := sm.scenarios[name]
	return scenario, ok
}

func (sm *ScenarioManager) StartComposite(configs []ScenarioConfig) (*CompositeScenarioResp, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.currentSession != nil {
		sm.stopCurrentSession()
	}

	ctx, cancel := context.WithCancel(context.Background())
	session := &ScenarioSession{
		SessionId:  fmt.Sprintf("session-%d", time.Now().Unix()),
		Scenarios:  []string{},
		StartTime:  time.Now(),
		CancelFunc: cancel,
	}

	details := make([]ScenarioDetail, 0, len(configs))
	var maxDuration int

	for _, config := range configs {
		scenario, ok := sm.scenarios[config.Name]
		if !ok {
			details = append(details, ScenarioDetail{
				Name:    config.Name,
				Success: false,
				Error:   fmt.Sprintf("scenario %s not found", config.Name),
			})
			continue
		}

		err := scenario.Start(ctx, config.Params)
		if err != nil {
			details = append(details, ScenarioDetail{
				Name:    config.Name,
				Success: false,
				Error:   err.Error(),
			})
		} else {
			session.Scenarios = append(session.Scenarios, config.Name)
			details = append(details, ScenarioDetail{
				Name:    config.Name,
				Success: true,
			})
		}

		if config.Duration > maxDuration {
			maxDuration = config.Duration
		}
	}

	if maxDuration > 0 {
		session.RecoveryTimer = time.AfterFunc(time.Duration(maxDuration)*time.Second, func() {
			sm.StopAllScenarios()
		})
	}

	sm.currentSession = session

	status := "success"
	for _, detail := range details {
		if !detail.Success {
			status = "partial"
			break
		}
	}
	if len(session.Scenarios) == 0 {
		status = "failed"
	}

	return &CompositeScenarioResp{
		SessionId: session.SessionId,
		Scenarios: session.Scenarios,
		Status:    status,
		Details:   details,
	}, nil
}

func (sm *ScenarioManager) stopCurrentSession() {
	if sm.currentSession == nil {
		return
	}

	if sm.currentSession.RecoveryTimer != nil {
		sm.currentSession.RecoveryTimer.Stop()
	}

	sm.currentSession.CancelFunc()

	for _, scenarioName := range sm.currentSession.Scenarios {
		if scenario, ok := sm.scenarios[scenarioName]; ok {
			scenario.Stop()
		}
	}

	sm.currentSession = nil
}

func (sm *ScenarioManager) StopAllScenarios() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.stopCurrentSession()

	return nil
}

func (sm *ScenarioManager) GetCurrentSession() *CompositeScenarioResp {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.currentSession == nil {
		return &CompositeScenarioResp{
			Status: "no active session",
		}
	}

	details := make([]ScenarioDetail, 0, len(sm.currentSession.Scenarios))
	for _, name := range sm.currentSession.Scenarios {
		if scenario, ok := sm.scenarios[name]; ok {
			status := scenario.Status()
			details = append(details, ScenarioDetail{
				Name:    name,
				Success: status.Running,
			})
		}
	}

	return &CompositeScenarioResp{
		SessionId: sm.currentSession.SessionId,
		Scenarios: sm.currentSession.Scenarios,
		Status:    "running",
		Details:   details,
	}
}

type ScenarioConfig struct {
	Name     string                 `json:"name"`
	Params   map[string]interface{} `json:"params"`
	Duration int                    `json:"duration,omitempty"`
}

type CompositeScenarioResp struct {
	SessionId string           `json:"session_id"`
	Scenarios []string         `json:"scenarios"`
	Status    string           `json:"status"`
	Details   []ScenarioDetail `json:"details"`
}

type ScenarioDetail struct {
	Name    string `json:"name"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
