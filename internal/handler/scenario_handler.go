package handler

import (
	"net/http"

	"github.com/Z3Labs/MockServer/internal/manager"
	"github.com/Z3Labs/MockServer/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
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
	var pathParams struct {
		Scenario string `path:"scenario"`
	}
	if err := httpx.Parse(r, &pathParams); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	var req map[string]interface{}
	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	configs := []manager.ScenarioConfig{
		{
			Name:   pathParams.Scenario,
			Params: req,
		},
	}

	resp, err := h.svcCtx.ScenarioManager.StartComposite(configs)
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	httpx.OkJsonCtx(r.Context(), w, resp)
}

func (h *ScenarioHandler) StopScenario(w http.ResponseWriter, r *http.Request) {
	var pathParams struct {
		Scenario string `path:"scenario"`
	}
	if err := httpx.Parse(r, &pathParams); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	err := h.svcCtx.ScenarioManager.Stop(pathParams.Scenario)
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	status, _ := h.svcCtx.ScenarioManager.Status(pathParams.Scenario)

	httpx.OkJsonCtx(r.Context(), w, status)
}

func (h *ScenarioHandler) GetScenarioStatus(w http.ResponseWriter, r *http.Request) {
	var pathParams struct {
		Scenario string `path:"scenario"`
	}
	if err := httpx.Parse(r, &pathParams); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	status, err := h.svcCtx.ScenarioManager.Status(pathParams.Scenario)
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	httpx.OkJsonCtx(r.Context(), w, status)
}

func (h *ScenarioHandler) ListScenarios(w http.ResponseWriter, r *http.Request) {
	scenarios := h.svcCtx.ScenarioManager.List()

	resp := map[string]interface{}{
		"scenarios": scenarios,
	}

	httpx.OkJsonCtx(r.Context(), w, resp)
}

func (h *ScenarioHandler) StartCompositeScenario(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Scenarios []manager.ScenarioConfig `json:"scenarios"`
	}

	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	resp, err := h.svcCtx.ScenarioManager.StartComposite(req.Scenarios)
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	httpx.OkJsonCtx(r.Context(), w, resp)
}

func (h *ScenarioHandler) StopAllScenarios(w http.ResponseWriter, r *http.Request) {
	err := h.svcCtx.ScenarioManager.StopAllScenarios()
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	httpx.OkJsonCtx(r.Context(), w, map[string]string{
		"status": "all scenarios stopped",
	})
}

func (h *ScenarioHandler) GetCurrentSession(w http.ResponseWriter, r *http.Request) {
	resp := h.svcCtx.ScenarioManager.GetCurrentSession()

	httpx.OkJsonCtx(r.Context(), w, resp)
}
