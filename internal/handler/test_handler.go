package handler

import (
	"net/http"
	"time"

	"github.com/Z3Labs/MockServer/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type TestHandler struct {
	svcCtx *svc.ServiceContext
}

func NewTestHandler(svcCtx *svc.ServiceContext) *TestHandler {
	return &TestHandler{
		svcCtx: svcCtx,
	}
}

func (h *TestHandler) Test10ms(w http.ResponseWriter, r *http.Request) {
	time.Sleep(10 * time.Millisecond)
	httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{
		"message":  "test endpoint with 10ms sleep",
		"sleep_ms": 10,
	})
}

func (h *TestHandler) Test30ms(w http.ResponseWriter, r *http.Request) {
	time.Sleep(30 * time.Millisecond)
	httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{
		"message":  "test endpoint with 30ms sleep",
		"sleep_ms": 30,
	})
}
