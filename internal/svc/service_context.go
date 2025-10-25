package svc

import (
	"github.com/Z3Labs/MockServer/internal/config"
	"github.com/Z3Labs/MockServer/internal/manager"
)

type ServiceContext struct {
	Config          config.Config
	ScenarioManager *manager.ScenarioManager
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:          c,
		ScenarioManager: manager.NewScenarioManager(),
	}
}
