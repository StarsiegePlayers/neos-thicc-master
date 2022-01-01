package maintenance

import (
	"time"

	"github.com/StarsiegePlayers/neos-thicc-master/src/config"
	"github.com/StarsiegePlayers/neos-thicc-master/src/log"
	"github.com/StarsiegePlayers/neos-thicc-master/src/service"
)

type DailyService struct {
	*time.Timer
	Next     time.Time
	Diff     time.Duration
	Services *map[service.ID]service.Interface
	Config   *config.Service

	*log.Log
	service.Interface
}

func (s *DailyService) Init(services *map[service.ID]service.Interface) (err error) {
	s.Services = services
	s.Config = (*s.Services)[service.Config].(*config.Service)
	s.Log = (*s.Services)[service.Logger].(*log.Service).NewLogger(service.DailyMaintenance)

	s.update()

	return nil
}

func (s *DailyService) Run() {
	s.Logf("will run next on %s", s.Next.Format(time.Stamp))

	for range s.C {
		for _, v := range *s.Services {
			if sv, ok := v.(service.DailyMaintainable); ok {
				sv.DailyMaintenance()
			}
		}

		s.update()
		s.Logf("will run next at %s", s.Next.Format(time.Stamp))
	}
}

func (s *DailyService) Rehash() {
	s.Shutdown()
	s.update()
	s.Run()
}

func (s *DailyService) update() {
	now := time.Now()
	s.Next = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	if !s.Next.After(time.Now()) {
		s.Next = s.Next.Add(24 * time.Hour) //nolint:gomnd
	}

	s.Diff = time.Until(s.Next)
	if s.Timer == nil {
		s.Timer = time.NewTimer(s.Diff)
	} else {
		s.Timer.Reset(s.Diff)
	}
}

func (s *DailyService) Shutdown() {
	s.Stop()
	s.Logf("shutdown complete")
}
