package main

import (
	"time"
)

type MaintenanceService struct {
	*time.Ticker
	time.Duration
	Services *map[string]Service
	Config   *Configuration

	Service
	Component
}

type Maintainable interface {
	Maintenance()
}

func (t *MaintenanceService) Init(args map[string]interface{}) (err error) {
	t.Name = "Maintenance"
	t.LogTag = "maintenance"

	var ok bool
	t.Config, ok = args["config"].(*Configuration)
	if !ok {
		t.LogAlert("config %s", ErrorInvalidArgument)
		return ErrorInvalidArgument
	}

	t.Services, ok = args["services"].(*map[string]Service)
	if !ok {
		t.LogAlert("services %s", ErrorInvalidArgument)
		return ErrorInvalidArgument
	}

	t.Ticker = time.NewTicker(t.Config.Advanced.Maintenance.Interval)

	return nil
}

func (t *MaintenanceService) Rehash() {
	t.Shutdown()
	t.Run()
}

func (t *MaintenanceService) Run() {
	t.Log("will run every %s", t.Config.Advanced.Maintenance.Interval.String())
	for range t.C {
		for _, v := range *t.Services {
			if service, ok := v.(Maintainable); ok {
				service.Maintenance()
			}
		}
	}
}

func (t *MaintenanceService) Shutdown() {
	t.Log("shutdown requested")
	t.Stop()
}
