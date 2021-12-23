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
	Services       *map[string]Service
	PollMasterInfo *PollMasterInfo
	Service
	Logger
}

type PollMasterInfo struct {
	Masters []*query.MasterQuery
	Games   map[string]*server.Server
	Errors  []error
}

func (p *PollService) Init(args map[string]interface{}) (err error) {
	p.Logger = Logger{
		Name:   "Server Poll",
		LogTag: "poll",
	}
	var ok bool
	p.Config, ok = args["config"].(*Configuration)
	if !ok {
		return ErrorInvalidArgument
	}

	p.Services, ok = args["services"].(*map[string]Service)
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

	(*p.Services)["master"].(*MasterService).RegisterExternalServerList(pm.Games)
}

func (p *PollService) Shutdown() {
	p.Log("shutdown requested")
	p.Stop()
}
