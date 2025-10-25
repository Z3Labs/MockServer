package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Z3Labs/MockServer/internal/scenarios"
	"github.com/Z3Labs/MockServer/internal/svc"
)

type HealthHandler struct {
	svcCtx *svc.ServiceContext
}

func NewHealthHandler(svcCtx *svc.ServiceContext) *HealthHandler {
	return &HealthHandler{
		svcCtx: svcCtx,
	}
}

func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	scenario, ok := h.svcCtx.ScenarioManager.GetScenario("health_check")
	if !ok {
		h.writeHealthy(w)
		return
	}

	healthScenario, ok := scenario.(*scenarios.HealthCheckFailure)
	if !ok {
		h.writeHealthy(w)
		return
	}

	shouldFail, statusCode, delay := healthScenario.ShouldFail()
	
	if delay > 0 {
		time.Sleep(delay)
	}

	if shouldFail {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "unhealthy",
		})
		return
	}

	h.writeHealthy(w)
}

func (h *HealthHandler) ReadyCheck(w http.ResponseWriter, r *http.Request) {
	h.HealthCheck(w, r)
}

func (h *HealthHandler) writeHealthy(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

func (h *HealthHandler) MockService(w http.ResponseWriter, r *http.Request) {
	scenario, ok := h.svcCtx.ScenarioManager.GetScenario("dependency")
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
		return
	}

	depScenario, ok := scenario.(*scenarios.DependencyFailure)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
		return
	}

	failureType, active := depScenario.GetFailureType()
	if !active {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
		return
	}

	switch failureType {
	case "timeout":
		time.Sleep(30 * time.Second)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	case "error":
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	case "slow":
		time.Sleep(3 * time.Second)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	default:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	}
}
