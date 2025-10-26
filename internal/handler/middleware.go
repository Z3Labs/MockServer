package handler

import (
	"net/http"
	"time"

	"github.com/Z3Labs/MockServer/internal/scenarios"
	"github.com/Z3Labs/MockServer/internal/svc"
)

func LatencyMiddleware(svcCtx *svc.ServiceContext) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			scenario, ok := svcCtx.ScenarioManager.GetScenario("network_latency")
			if ok {
				if latencyScenario, ok := scenario.(*scenarios.NetworkLatency); ok {
					delay := latencyScenario.GetLatency()
					if delay > 0 {
						time.Sleep(delay)
					}
				}
			}
			next(w, r)
		}
	}
}
