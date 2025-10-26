package handler

import (
	"net/http"
	"time"

	"github.com/Z3Labs/MockServer/internal/scenarios"
	"github.com/Z3Labs/MockServer/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
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
		h.writeHealthy(w, r)
		return
	}

	healthScenario, ok := scenario.(*scenarios.HealthCheckFailure)
	if !ok {
		h.writeHealthy(w, r)
		return
	}

	shouldFail, statusCode, delay := healthScenario.ShouldFail()
	
	if delay > 0 {
		time.Sleep(delay)
	}

	if shouldFail {
		w.WriteHeader(statusCode)
		httpx.OkJsonCtx(r.Context(), w, map[string]string{
			"status": "unhealthy",
		})
		return
	}

	h.writeHealthy(w, r)
}

func (h *HealthHandler) ReadyCheck(w http.ResponseWriter, r *http.Request) {
	h.HealthCheck(w, r)
}

func (h *HealthHandler) writeHealthy(w http.ResponseWriter, r *http.Request) {
	httpx.OkJsonCtx(r.Context(), w, map[string]string{
		"status": "healthy",
	})
}

func (h *HealthHandler) MockService(w http.ResponseWriter, r *http.Request) {
	scenario, ok := h.svcCtx.ScenarioManager.GetScenario("dependency")
	if !ok {
		httpx.OkJsonCtx(r.Context(), w, map[string]string{
			"status": "ok",
		})
		return
	}

	depScenario, ok := scenario.(*scenarios.DependencyFailure)
	if !ok {
		httpx.OkJsonCtx(r.Context(), w, map[string]string{
			"status": "ok",
		})
		return
	}

	failureType, active := depScenario.GetFailureType()
	if !active {
		httpx.OkJsonCtx(r.Context(), w, map[string]string{
			"status": "ok",
		})
		return
	}

	switch failureType {
	case "timeout":
		time.Sleep(30 * time.Second)
		httpx.OkJsonCtx(r.Context(), w, map[string]string{
			"status": "ok",
		})
	case "error":
		httpx.ErrorCtx(r.Context(), w, http.ErrAbortHandler)
	case "slow":
		time.Sleep(3 * time.Second)
		httpx.OkJsonCtx(r.Context(), w, map[string]string{
			"status": "ok",
		})
	default:
		httpx.OkJsonCtx(r.Context(), w, map[string]string{
			"status": "ok",
		})
	}
}
