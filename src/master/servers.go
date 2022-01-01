package master

import (
	"fmt"
	"net"
	"time"

	"github.com/StarsiegePlayers/neos-thicc-master/src/config"
	"github.com/StarsiegePlayers/neos-thicc-master/src/service"

	"github.com/StarsiegePlayers/darkstar-query-go/v2/protocol"
	"github.com/StarsiegePlayers/darkstar-query-go/v2/server"
)

type Servers struct {
	Master       *protocol.Master
	BannedMaster *protocol.Master
	LocalMaster  *protocol.Master

	Options *protocol.Options

	Services        *map[service.ID]service.Interface
	Config          *config.Service
	TemplateService service.Interface
	STUNService     service.Interface
}

func (s *Servers) Init(services *map[service.ID]service.Interface) (err error) {
	s.Master = protocol.NewMaster()
	s.BannedMaster = protocol.NewMaster()
	s.LocalMaster = protocol.NewMaster()

	s.Options = &protocol.Options{}

	s.Services = services
	s.Config = (*s.Services)[service.Config].(*config.Service)
	s.TemplateService = (*s.Services)[service.Template]
	s.STUNService = (*s.Services)[service.STUN]

	return
}

func (s *Servers) Rehash() {
	s.Master.MOTD = s.TemplateService.Get()
	s.Master.MasterID = s.Config.Values.Service.ID
	s.Master.CommonName = s.Config.Values.Service.Hostname
	s.Master.MOTDJunk = dummythicc

	s.LocalMaster.MOTD = s.TemplateService.Get()
	s.LocalMaster.MasterID = s.Config.Values.Service.ID
	s.LocalMaster.CommonName = s.Config.Values.Service.Hostname
	s.LocalMaster.MOTDJunk = dummythicc

	s.BannedMaster.MOTD = s.Config.Values.Service.Banned.Message
	s.BannedMaster.MasterID = s.Config.Values.Service.ID
	s.BannedMaster.CommonName = s.Config.Values.Service.Hostname
	s.BannedMaster.MOTDJunk = dummythicc

	s.Options.Debug = s.Config.Values.Advanced.Verbose
	s.Options.MaxServerPacketSize = s.Config.Values.Advanced.Network.MaxPacketSize
	s.Options.MaxNetworkPacketSize = s.Config.Values.Advanced.Network.MaxBufferSize
	s.Options.Timeout = s.Config.Values.Advanced.Network.ConnectionTimeout.Duration
}

func (s *Servers) Exist(ipPort string) (exist bool) {
	_, exist = s.Master.Servers[ipPort]

	return
}

func (s *Servers) Get(ipPort string) *protocol.Master {
	if ipPort == "" {
		return s.BannedMaster
	}

	host, _, _ := net.SplitHostPort(ipPort)
	if host == service.LocalhostAddress {
		s.LocalMaster.MOTD = s.TemplateService.Get()
		return s.LocalMaster
	}

	s.Master.MOTD = s.TemplateService.Get()

	return s.Master
}

func (s *Servers) UpdateServer(ipPort string) time.Time {
	out := s.LocalMaster.Servers[ipPort].LastSeen
	s.LocalMaster.Servers[ipPort].LastSeen = time.Now()

	host, port, _ := net.SplitHostPort(ipPort)
	if host == service.LocalhostAddress {
		ipPort = fmt.Sprintf("%s:%s", s.STUNService.Get(), port)
	}

	s.Master.Servers[ipPort].LastSeen = time.Now()

	return out
}

func (s *Servers) Add(ipPort string, addr *net.UDPAddr, pconn *net.PacketConn) {
	now := time.Now()
	s.LocalMaster.Servers[ipPort] = &server.Server{
		Address:    addr,
		Connection: pconn,
		LastSeen:   now,
	}

	host, port, _ := net.SplitHostPort(ipPort)
	if host == service.LocalhostAddress {
		ipPort = fmt.Sprintf("%s:%s", s.STUNService.Get(), port)

		lAddr, err := net.ResolveUDPAddr("udp", ipPort)
		if err != nil {
			lAddr = addr
		}

		s.Master.Servers[ipPort] = &server.Server{
			Address:    lAddr,
			Connection: pconn,
			LastSeen:   now,
		}
	} else {
		s.Master.Servers[ipPort] = s.LocalMaster.Servers[ipPort]
	}
}

func (s *Servers) Remove(ipPort string) {
	delete(s.LocalMaster.Servers, ipPort)

	host, port, _ := net.SplitHostPort(ipPort)
	if host == service.LocalhostAddress {
		ipPort = fmt.Sprintf("%s:%s", s.STUNService.Get(), port)
	}

	delete(s.Master.Servers, ipPort)
}
