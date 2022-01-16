package stats

import (
	"sync"

	"github.com/StarsiegePlayers/neos-thicc-master/src/log"
	"github.com/StarsiegePlayers/neos-thicc-master/src/service"
)

type Service struct {
	sync.Mutex

	stats struct {
		DailyHosts  map[string]int
		ActiveGames map[string]byte
	}
	services struct {
		Map *map[service.ID]service.Interface
	}
	logs struct {
		Stats *log.Log
	}

	service.Interface
	service.DailyMaintainable
}

func (s *Service) Init(services *map[service.ID]service.Interface) (err error) {
	s.services.Map = services
	s.stats.DailyHosts = make(map[string]int)
	s.stats.ActiveGames = make(map[string]byte)
	s.logs.Stats = (*s.services.Map)[service.Log].(*log.Service).NewLogger(service.BannedTrafficLog)

	return
}

func (s *Service) Status() service.LifeCycle {
	return service.Static
}

func (s *Service) DailyMaintenance() {
	s.logs.Stats.Logf("{%s} resetting daily user count, previous count: %d users", service.DailyMaintenance, len(s.stats.DailyHosts))
	s.Lock()
	s.stats.DailyHosts = make(map[string]int)
	s.Unlock()
}

func (s *Service) AddDailyClientNumber(host string) (out int) {
	s.Lock()
	out = s.GetDailyClientsTotal()
	s.stats.DailyHosts[host] = out
	s.Unlock()

	return
}

func (s *Service) GetDailyHostNumber(host string) (out int) {
	var ok bool
	if out, ok = s.stats.DailyHosts[host]; !ok {
		out = s.AddDailyClientNumber(host)
	}

	return
}

func (s *Service) GetDailyClientsTotal() int {
	return len(s.stats.DailyHosts) + 1
}

func (s *Service) GetTotalServersWithPlayers() int {
	return len(s.stats.ActiveGames)
}

func (s *Service) UpdatePlayerCountForServer(ipPort string, count byte) {
	s.Lock()
	if count <= 0 {
		delete(s.stats.ActiveGames, ipPort)
	} else {
		s.stats.ActiveGames[ipPort] = count
	}
	s.Unlock()
}

func (s *Service) RemoveServer(ipPort string) {
	s.Lock()
	delete(s.stats.ActiveGames, ipPort)
	s.Unlock()
}
