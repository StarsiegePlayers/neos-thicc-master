package maintenance

import (
	"time"

	"github.com/StarsiegePlayers/neos-thicc-master/src/config"
	"github.com/StarsiegePlayers/neos-thicc-master/src/log"
	"github.com/StarsiegePlayers/neos-thicc-master/src/service"
)

const DayDuration = time.Hour * 24

type DailyService struct {
	*time.Timer
	time.Duration

	status   service.LifeCycle
	next     time.Time
	services struct {
		Map    *map[service.ID]service.Interface
		Config *config.Service
	}
	logs struct {
		DailyMaintenance *log.Log
	}

	service.Interface
	service.Runnable
}

func (s *DailyService) Init(services *map[service.ID]service.Interface) (err error) {
	s.services.Map = services
	s.services.Config = (*s.services.Map)[service.Config].(*config.Service)
	s.logs.DailyMaintenance = (*s.services.Map)[service.Log].(*log.Service).NewLogger(service.DailyMaintenance)
	s.status = service.Starting

	s.update()

	return nil
}

func (s *DailyService) Run() {
	s.logs.DailyMaintenance.Logf("will run next on %s", s.next.Format(time.Stamp))
	s.status = service.Running

	for range s.C {
		for _, v := range *s.services.Map {
			if sv, ok := v.(service.DailyMaintainable); ok {
				sv.DailyMaintenance()
			}
		}

		s.update()
		s.logs.DailyMaintenance.Logf("will run next at %s", s.next.Format(time.Stamp))
	}

	s.status = service.Stopped
}

func (s *DailyService) Rehash() {
	// noop
}

func (s *DailyService) Shutdown() {
	s.status = service.Stopping
	s.Stop()
	s.status = service.Stopped

	s.logs.DailyMaintenance.Logf("shutdown complete")
}

func (s *DailyService) Status() service.LifeCycle {
	return s.status
}

func (s *DailyService) update() {
	now := time.Now()
	s.next = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	if !s.next.After(time.Now()) {
		s.next = s.next.Add(DayDuration)
	}

	s.Duration = time.Until(s.next)
	if s.Timer == nil {
		s.Timer = time.NewTimer(s.Duration)
	} else {
		s.Timer.Reset(s.Duration)
	}
}
