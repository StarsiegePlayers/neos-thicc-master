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

	return
}

func (p *PollService) Rehash() {
	p.Shutdown()
	p.Run()
}

func (p *PollService) Run() {
	p.Log("will run every %s", p.Config.Poll.Interval.String())
	for range p.C {
		q := darkstar.NewQuery(p.Config.Advanced.Network.ConnectionTimeout, true)
		q.Addresses = p.Config.Poll.KnownMasters

		pm := new(PollMasterInfo)
		pm.Masters, pm.Games, pm.Errors = q.Masters()
		p.Log("found %d games on %d masters", len(pm.Masters), len(pm.Games))

		p.Lock()
		p.PollMasterInfo = pm
		p.Unlock()
	}
}

func (p *PollService) Shutdown() {
	p.Log("shutdown requested")
	p.Stop()
}
