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

type PollService struct {
	sync.Mutex
	*time.Ticker

	Services       *map[service.ID]service.Interface
	Config         *config.Service
	MasterService  *master.Service
	PollMasterInfo *PollMasterInfo

	*log.Log
	service.Interface
}

type PollMasterInfo struct {
	Masters []*query.MasterQuery
	Games   map[string]*server.Server
	Errors  []error
}

func (p *PollService) Init(services *map[service.ID]service.Interface) (err error) {

	p.Services = services
	p.Config = (*p.Services)[service.Config].(*config.Service)
	p.MasterService = (*p.Services)[service.Master].(*master.Service)
	p.Log = (*p.Services)[service.Log].(*log.Service).NewLogger(service.Poll)

	p.Ticker = time.NewTicker(p.Config.Values.Poll.Interval.Duration)

	return
}

func (p *PollService) Rehash() {
	p.Shutdown()
	p.Run()
}

func (p *PollService) Run() {
	p.Logf("will run every %s", p.Config.Values.Poll.Interval.String())
	p.Logf("known masters are %s", p.Config.Values.Poll.KnownMasters)
	p.query()

	for range p.C {
		p.query()
	}
}

func (p *PollService) query() {
	q := darkstar.NewQuery(p.Config.Values.Advanced.Network.ConnectionTimeout.Duration, p.Config.Values.Advanced.Verbose)
	q.Addresses = p.Config.Values.Poll.KnownMasters

	pm := new(PollMasterInfo)
	pm.Masters, pm.Games, pm.Errors = q.Masters()
	p.Logf("found %d games on %d masters", len(pm.Games), len(pm.Masters))

	p.Lock()
	p.PollMasterInfo = pm
	p.Unlock()

	p.MasterService.RegisterExternalServerList(pm.Games)
}

func (p *PollService) Shutdown() {
	p.Stop()
	p.Logf("shutdown complete")
}
