package master

import (
	"fmt"
	"github.com/StarsiegePlayers/darkstar-query-go/v2/protocol"
	"github.com/StarsiegePlayers/darkstar-query-go/v2/query"
	"github.com/StarsiegePlayers/darkstar-query-go/v2/server"
	"github.com/StarsiegePlayers/neos-thicc-master/src/service"
	"net"
	"sync"
	"time"
)

type ServerList struct {
	sync.Mutex

	stun            service.Interface
	ServerList      map[string]*ServerInfo
	LocalServerList map[string]*ServerInfo
}

func (s *ServerList) Init(stun service.Interface) {
	s.stun = stun
	s.ServerList = make(map[string]*ServerInfo)
	s.LocalServerList = make(map[string]*ServerInfo)
}

func (s *ServerList) Add(addr *net.UDPAddr, ipPort string, options *protocol.Options) {
	s.Lock()
	s.LocalServerList[ipPort] = new(ServerInfo)
	s.LocalServerList[ipPort].PingInfoQuery = query.NewPingInfoQueryWithOptions(ipPort, options)
	s.LocalServerList[ipPort].Server = new(server.Server)
	s.LocalServerList[ipPort].Server.Address = addr

	host, port, _ := net.SplitHostPort(ipPort)
	if host == service.LocalhostAddress {
		ipPort = fmt.Sprintf("%s:%s", s.stun.Get(), port)
		addr, _ = net.ResolveUDPAddr("udp", ipPort)
	}

	s.ServerList[ipPort] = new(ServerInfo)
	s.ServerList[ipPort].PingInfoQuery = query.NewPingInfoQueryWithOptions(ipPort, options)
	s.ServerList[ipPort].Server = new(server.Server)
	s.ServerList[ipPort].Server.Address = addr

	s.Unlock()
}

func (s *ServerList) Exist(ipPort string) bool {
	_, ok := s.LocalServerList[ipPort]
	return ok
}

func (s *ServerList) Update(ipPort string, response *query.PingInfoQuery) {
	s.Lock()
	s.LocalServerList[ipPort].SolicitedTime = time.Now()
	s.LocalServerList[ipPort].LastSeen = time.Now()
	s.LocalServerList[ipPort].PingInfoQuery = response

	host, port, _ := net.SplitHostPort(ipPort)
	if host == "127.0.0.1" {
		ipPort = fmt.Sprintf("%s:%s", s.stun.Get(), port)
	}

	s.ServerList[ipPort].SolicitedTime = time.Now()
	s.ServerList[ipPort].LastSeen = time.Now()
	s.ServerList[ipPort].PingInfoQuery = response

	s.Unlock()
}

func (s *ServerList) Remove(ipPort string) {
	s.Lock()
	delete(s.LocalServerList, ipPort)

	host, port, _ := net.SplitHostPort(ipPort)
	if host == "127.0.0.1" {
		ipPort = fmt.Sprintf("%s:%s", s.stun.Get(), port)
	}

	delete(s.ServerList, ipPort)

	s.Unlock()
}
