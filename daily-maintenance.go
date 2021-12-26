package main

import (
	"time"
)

type DailyMaintenanceService struct {
	*time.Timer
	Next     time.Time
	Diff     time.Duration
	Services *map[ServiceID]Service
	Config   *Configuration

	Service
	Logger
}

type DailyMaintainable interface {
	DailyMaintenance()
}

func (t *DailyMaintenanceService) Init(args map[InitArg]interface{}) (err error) {
	t.Logger = Logger{
		Name: "daily-maintenance",
		ID:   DailyMaintenanceServiceID,
	}
	var ok bool
	t.Config, ok = args[InitArgConfig].(*Configuration)
	if !ok {
		return ErrorInvalidArgument
	}

	t.Services, ok = args[InitArgServices].(*map[ServiceID]Service)
	if !ok {
		return ErrorInvalidArgument
	}

	t.update()

	return nil
}

func (t *DailyMaintenanceService) Run() {
	t.Log("will run next on %s", t.Next.Format(time.Stamp))
	for range t.C {
		for _, v := range *t.Services {
			if service, ok := v.(DailyMaintainable); ok {
				service.DailyMaintenance()
			}
		}
		t.update()
		t.Log("will run next at %s", t.Next.Format(time.Stamp))
	}
}

func (t *DailyMaintenanceService) Rehash() {
	t.Shutdown()
	t.update()
	t.Run()
}

func (t *DailyMaintenanceService) update() {
	now := time.Now()
	t.Next = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	if !t.Next.After(time.Now()) {
		t.Next = t.Next.Add(24 * time.Hour)
	}

	t.Diff = t.Next.Sub(time.Now())
	if t.Timer == nil {
		t.Timer = time.NewTimer(t.Diff)
	} else {
		t.Timer.Reset(t.Diff)
	}
}

func (t *DailyMaintenanceService) Shutdown() {
	t.Stop()
	t.Log("shutdown complete")
}
