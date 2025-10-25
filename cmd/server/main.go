package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Z3Labs/MockServer/internal/config"
	"github.com/Z3Labs/MockServer/internal/handler"
	"github.com/Z3Labs/MockServer/internal/svc"
	"gopkg.in/yaml.v3"
)

var configFile = flag.String("f", "etc/mockserver.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	data, err := os.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	if err := yaml.Unmarshal(data, &c); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	svcCtx := svc.NewServiceContext(c)

	scenarioHandler := handler.NewScenarioHandler(svcCtx)
	healthHandler := handler.NewHealthHandler(svcCtx)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/scenarios/{scenario}/start", scenarioHandler.StartScenario)
	mux.HandleFunc("POST /api/v1/scenarios/{scenario}/stop", scenarioHandler.StopScenario)
	mux.HandleFunc("GET /api/v1/scenarios/{scenario}/status", scenarioHandler.GetScenarioStatus)
	mux.HandleFunc("GET /api/v1/scenarios", scenarioHandler.ListScenarios)

	mux.HandleFunc("POST /api/v1/composite/start", scenarioHandler.StartCompositeScenario)
	mux.HandleFunc("POST /api/v1/composite/stop", scenarioHandler.StopAllScenarios)
	mux.HandleFunc("GET /api/v1/composite/status", scenarioHandler.GetCurrentSession)

	mux.HandleFunc("GET /health", healthHandler.HealthCheck)
	mux.HandleFunc("GET /ready", healthHandler.ReadyCheck)
	mux.HandleFunc("GET /api/v1/mock-service", healthHandler.MockService)

	wrappedMux := handler.LatencyMiddleware(svcCtx)(mux)

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	log.Printf("Starting MockServer at %s", addr)

	if err := http.ListenAndServe(addr, wrappedMux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
