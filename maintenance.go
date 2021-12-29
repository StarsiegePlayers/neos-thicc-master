package main

import (
	"time"
)

type MaintenanceService struct {
	*time.Ticker
	time.Duration
	Services *map[ServiceID]Service
	Config   *ConfigurationService

	Service
	Logger
}

type Maintainable interface {
	Maintenance()
}

func (t *MaintenanceService) Init(args map[InitArg]interface{}) (err error) {
	t.Logger = Logger{
		Name: "maintenance",
		ID:   MaintenanceServiceID,
	}

	var ok bool
	t.Config, ok = args[InitArgConfig].(*ConfigurationService)
	if !ok {
		t.LogAlert("config %s", ErrorInvalidArgument)
		return ErrorInvalidArgument
	}

	t.Services, ok = args[InitArgServices].(*map[ServiceID]Service)
	if !ok {
		t.LogAlert("services %s", ErrorInvalidArgument)
		return ErrorInvalidArgument
	}

	t.Ticker = time.NewTicker(t.Config.Values.Advanced.Maintenance.Interval.Duration)

	return nil
}

func (t *MaintenanceService) Rehash() {
	t.Shutdown()
	t.Run()
}

func (t *MaintenanceService) Run() {
	t.Log("will run every %s", t.Config.Values.Advanced.Maintenance.Interval.String())
	for range t.C {
		for _, v := range *t.Services {
			if service, ok := v.(Maintainable); ok {
				service.Maintenance()
			}
		}
	}
}

func (t *MaintenanceService) Shutdown() {
	t.Stop()
	t.Log("shutdown complete")
}
