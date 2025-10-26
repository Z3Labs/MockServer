package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/Z3Labs/MockServer/internal/config"
	"github.com/Z3Labs/MockServer/internal/handler"
	"github.com/Z3Labs/MockServer/internal/svc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/mockserver.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	svcCtx := svc.NewServiceContext(c)

	scenarioHandler := handler.NewScenarioHandler(svcCtx)
	healthHandler := handler.NewHealthHandler(svcCtx)

	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/api/v1/scenarios/:scenario/start",
		Handler: scenarioHandler.StartScenario,
	})
	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/api/v1/scenarios/:scenario/stop",
		Handler: scenarioHandler.StopScenario,
	})
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/api/v1/scenarios/:scenario/status",
		Handler: scenarioHandler.GetScenarioStatus,
	})
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/api/v1/scenarios",
		Handler: scenarioHandler.ListScenarios,
	})

	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/api/v1/composite/start",
		Handler: scenarioHandler.StartCompositeScenario,
	})
	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/api/v1/composite/stop",
		Handler: scenarioHandler.StopAllScenarios,
	})
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/api/v1/composite/status",
		Handler: scenarioHandler.GetCurrentSession,
	})

	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/health",
		Handler: healthHandler.HealthCheck,
	})
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/ready",
		Handler: healthHandler.ReadyCheck,
	})
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/api/v1/mock-service",
		Handler: healthHandler.MockService,
	})

	server.Use(handler.LatencyMiddleware(svcCtx))

	fmt.Printf("Starting MockServer at %s:%d\n", c.Host, c.Port)
	server.Start()
}
