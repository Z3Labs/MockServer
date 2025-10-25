package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Z3Labs/MockServer/internal/manager"
	"github.com/Z3Labs/MockServer/internal/svc"
)

type ScenarioHandler struct {
	svcCtx *svc.ServiceContext
}

func NewScenarioHandler(svcCtx *svc.ServiceContext) *ScenarioHandler {
	return &ScenarioHandler{
		svcCtx: svcCtx,
	}
}

func (h *ScenarioHandler) StartScenario(w http.ResponseWriter, r *http.Request) {
	scenarioName := r.PathValue("scenario")
	
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	configs := []manager.ScenarioConfig{
		{
			Name:   scenarioName,
			Params: req,
		},
	}

	resp, err := h.svcCtx.ScenarioManager.StartComposite(configs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *ScenarioHandler) StopScenario(w http.ResponseWriter, r *http.Request) {
	scenarioName := r.PathValue("scenario")

	err := h.svcCtx.ScenarioManager.Stop(scenarioName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	status, _ := h.svcCtx.ScenarioManager.Status(scenarioName)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *ScenarioHandler) GetScenarioStatus(w http.ResponseWriter, r *http.Request) {
	scenarioName := r.PathValue("scenario")

	status, err := h.svcCtx.ScenarioManager.Status(scenarioName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *ScenarioHandler) ListScenarios(w http.ResponseWriter, r *http.Request) {
	scenarios := h.svcCtx.ScenarioManager.List()

	resp := map[string]interface{}{
		"scenarios": scenarios,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *ScenarioHandler) StartCompositeScenario(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Scenarios []manager.ScenarioConfig `json:"scenarios"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.svcCtx.ScenarioManager.StartComposite(req.Scenarios)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *ScenarioHandler) StopAllScenarios(w http.ResponseWriter, r *http.Request) {
	err := h.svcCtx.ScenarioManager.StopAllScenarios()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "all scenarios stopped",
	})
}

func (h *ScenarioHandler) GetCurrentSession(w http.ResponseWriter, r *http.Request) {
	resp := h.svcCtx.ScenarioManager.GetCurrentSession()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
