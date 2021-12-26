package main

import (
	"sync"
	"time"

	darkstar "github.com/StarsiegePlayers/darkstar-query-go/v2"
	"github.com/StarsiegePlayers/darkstar-query-go/v2/query"
	"github.com/StarsiegePlayers/darkstar-query-go/v2/server"
)

type PollService struct {
	sync.Mutex
	*time.Ticker
	Config         *Configuration
	Services       *map[ServiceID]Service
	PollMasterInfo *PollMasterInfo
	MasterService  *MasterService
	Service
	Logger
}

type PollMasterInfo struct {
	Masters []*query.MasterQuery
	Games   map[string]*server.Server
	Errors  []error
}

func (p *PollService) Init(args map[InitArg]interface{}) (err error) {
	p.Logger = Logger{
		Name: "poll",
		ID:   PollServiceID,
	}

	var ok bool
	p.Config, ok = args[InitArgConfig].(*Configuration)
	if !ok {
		return ErrorInvalidArgument
	}

	p.Services, ok = args[InitArgServices].(*map[ServiceID]Service)
	if !ok {
		return ErrorInvalidArgument
	}

	p.MasterService, ok = (*p.Services)[MasterServiceID].(*MasterService)
	if !ok {
		return ErrorInvalidArgument
	}

	p.Ticker = time.NewTicker(p.Config.Poll.Interval)

	return
}

func (p *PollService) Rehash() {
	p.Shutdown()
	p.Run()
}

func (p *PollService) Run() {
	p.Log("will run every %s", p.Config.Poll.Interval.String())
	p.Log("known masters are %s", p.Config.Poll.KnownMasters)
	p.query()
	for range p.C {
		p.query()
	}
}

func (p *PollService) query() {
	q := darkstar.NewQuery(p.Config.Advanced.Network.ConnectionTimeout, p.Config.Advanced.Verbose)
	q.Addresses = p.Config.Poll.KnownMasters

	pm := new(PollMasterInfo)
	pm.Masters, pm.Games, pm.Errors = q.Masters()
	p.Log("found %d games on %d masters", len(pm.Games), len(pm.Masters))

	p.Lock()
	p.PollMasterInfo = pm
	p.Unlock()

	p.MasterService.RegisterExternalServerList(pm.Games)
}

func (p *PollService) Shutdown() {
	p.Stop()
	p.Log("shutdown complete")
}
