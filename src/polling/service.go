package polling

import (
	"sync"
	"time"

	"github.com/StarsiegePlayers/neos-thicc-master/src/config"
	"github.com/StarsiegePlayers/neos-thicc-master/src/log"
	"github.com/StarsiegePlayers/neos-thicc-master/src/master"
	"github.com/StarsiegePlayers/neos-thicc-master/src/service"

	darkstar "github.com/StarsiegePlayers/darkstar-query-go/v2"
	"github.com/StarsiegePlayers/darkstar-query-go/v2/query"
	"github.com/StarsiegePlayers/darkstar-query-go/v2/server"
)

type Service struct {
	sync.Mutex
	*time.Ticker

	PollMasterInfo *PollMasterInfo

	services struct {
		Map    *map[service.ID]service.Interface
		Config *config.Service
		Master *master.Service
	}
	status   service.LifeCycle
	duration time.Duration
	log      *log.Log

	service.Interface
	service.Runnable
}

type PollMasterInfo struct {
	Masters []*query.MasterQuery
	Games   map[string]*server.Server
	Errors  []error
}

func (s *Service) Init(services *map[service.ID]service.Interface) (err error) {
	s.services.Map = services
	s.services.Config = (*s.services.Map)[service.Config].(*config.Service)
	s.services.Master = (*s.services.Map)[service.Master].(*master.Service)
	s.log = (*s.services.Map)[service.Log].(*log.Service).NewLogger(service.Poll)
	s.status = service.Starting

	return
}

func (s *Service) Rehash() {
	if s.duration != s.services.Config.Values.Poll.Interval.Duration {
		s.log.Logf("restarting poll service")
		s.Stop()

		go s.Run()
	}
}

func (s *Service) Run() {
	s.status = service.Running
	s.duration = s.services.Config.Values.Poll.Interval.Duration
	s.log.Logf("will run every %s", s.duration.String())
	s.log.Logf("known masters are %s", s.services.Config.Values.Poll.KnownMasters)

	s.Ticker = time.NewTicker(s.duration)
	s.query()

	for range s.C {
		s.query()
	}
}

func (s *Service) query() {
	q := darkstar.NewQuery(s.services.Config.Values.Advanced.Network.ConnectionTimeout.Duration, s.services.Config.Values.Advanced.Verbose)
	q.Addresses = s.services.Config.Values.Poll.KnownMasters

	pm := new(PollMasterInfo)
	pm.Masters, pm.Games, pm.Errors = q.Masters()
	s.log.Logf("found %d games on %d masters", len(pm.Games), len(pm.Masters))

	s.Lock()
	s.PollMasterInfo = pm
	s.Unlock()

	s.services.Master.RegisterExternalServerList(pm.Games)
}

func (s *Service) Shutdown() {
	s.status = service.Stopping
	s.Stop()
	s.status = service.Stopped
	s.log.Logf("shutdown complete")
}
