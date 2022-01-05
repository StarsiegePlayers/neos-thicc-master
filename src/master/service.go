package master

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/StarsiegePlayers/neos-thicc-master/src/config"
	"github.com/StarsiegePlayers/neos-thicc-master/src/log"
	"github.com/StarsiegePlayers/neos-thicc-master/src/service"

	darkstar "github.com/StarsiegePlayers/darkstar-query-go/v2"
	"github.com/StarsiegePlayers/darkstar-query-go/v2/protocol"
	"github.com/StarsiegePlayers/darkstar-query-go/v2/query"
	"github.com/StarsiegePlayers/darkstar-query-go/v2/server"
)

const dummythicc = "dummythicc"

type Service struct {
	sync.Mutex

	Masters        *Servers
	Options        *protocol.Options
	IPServiceCount map[string]uint16
	ServerList     *ServerList

	Services        *map[service.ID]service.Interface
	Config          *config.Service
	TemplateService service.Interface
	STUNService     service.Interface

	pconn net.PacketConn

	DailyStats DailyStats

	Logs struct {
		Master       *log.Log
		Heartbeat    *log.Log
		Registration *log.Log
		Banned       *log.Log
	}

	service.Interface
	service.Maintainable
	service.DailyMaintainable
}

type DailyStats struct {
	UniqueUsers map[string]bool
}

type ServerInfo struct {
	*query.PingInfoQuery
	*server.Server

	SolicitedTime time.Time
}

func (s *Service) Init(services *map[service.ID]service.Interface) (err error) {
	s.Masters = new(Servers)
	_ = s.Masters.Init(services)
	s.Options = &protocol.Options{}
	s.ServerList = &ServerList{}

	s.IPServiceCount = make(map[string]uint16)
	s.DailyStats = DailyStats{
		UniqueUsers: make(map[string]bool),
	}

	s.Services = services
	s.Config = (*s.Services)[service.Config].(*config.Service)
	s.TemplateService = (*s.Services)[service.Template]
	s.STUNService = (*s.Services)[service.STUN]

	s.Logs.Master = (*s.Services)[service.Log].(*log.Service).NewLogger(service.Master)
	s.Logs.Heartbeat = (*s.Services)[service.Log].(*log.Service).NewLogger(service.HeartbeatLog)
	s.Logs.Registration = (*s.Services)[service.Log].(*log.Service).NewLogger(service.ServerRegistrationLog)
	s.Logs.Banned = (*s.Services)[service.Log].(*log.Service).NewLogger(service.BannedTrafficLog)

	s.ServerList.Init(s.STUNService)

	addrPort := fmt.Sprintf("%s:%d", s.Config.Values.Service.Listen.IP, s.Config.Values.Service.Listen.Port)

	s.pconn, err = net.ListenPacket("udp", addrPort)
	if err != nil {
		s.Logs.Master.LogAlertf("unable to bind to %s - [%s]", addrPort, err)
		return
	}

	s.Logs.Master.Logf("now listening on [%s]", fmt.Sprintf("%s:%d", s.STUNService.Get(), s.Config.Values.Service.Listen.Port))
	s.Rehash()

	return
}

func (s *Service) Run() {
	// start listening loop
	buf := make([]byte, s.Config.Values.Advanced.Network.MaxPacketSize)
	buf2 := make([]byte, s.Config.Values.Advanced.Network.MaxPacketSize)
	prevIPPort := ""

	for {
		n, addr, err := s.pconn.ReadFrom(buf)
		if err != nil {
			switch t := err.(type) {
			case *net.OpError:
				if t.Op == "read" {
					s.Logs.Master.LogAlertf("socket closed.")
				}

				continue
			default:
				s.Logs.Master.LogAlertf("read error on socket [%s]", err)
			}
			break
		}

		// dedupe packets because wtf dynamix
		if prevIPPort == addr.String() && bytes.Equal(buf[:n], buf2[:n]) {
			prevIPPort = ""
			continue
		}

		copy(buf2, buf)

		prevIPPort = addr.String()

		if addr, ok := addr.(*net.UDPAddr); ok {
			go s.serveMaster(addr, buf[:n])
		}
	}

	s.Logs.Master.LogAlertf("service stopped")
}

func (s *Service) Maintenance() {
	count := 0
	checked := 0
	fresh := 0

	for k := range s.ServerList.LocalServerList {
		removed, queried := s.CheckRemoveServer(k)

		switch {
		case removed:
			count++
		case !removed && queried:
			checked++
		case !removed && !queried:
			fresh++
		}
	}

	s.Logs.Master.Logf("{%s} removed %d stale servers, queried %d servers, %d servers still fresh\n", service.Maintenance, count, checked, fresh)
}

func (s *Service) DailyMaintenance() {
	s.Lock()
	s.Logs.Master.Logf("{%s} resetting daily user count, previous count: %d users", service.DailyMaintenance, len(s.DailyStats.UniqueUsers))
	s.DailyStats.UniqueUsers = make(map[string]bool)
	s.Unlock()
}

func (s *Service) Rehash() {
	s.Logs.Master.Logf("{%s} Reading config", service.Rehash)
	s.Masters.Rehash()

	s.Options.Debug = s.Config.Values.Advanced.Verbose
	s.Options.MaxServerPacketSize = s.Config.Values.Advanced.Network.MaxPacketSize
	s.Options.MaxNetworkPacketSize = s.Config.Values.Advanced.Network.MaxBufferSize
	s.Options.Timeout = s.Config.Values.Advanced.Network.ConnectionTimeout.Duration
}

func (s *Service) Shutdown() {
	if err := s.pconn.Close(); err != nil {
		s.Logs.Master.LogAlertf("{%s} error while closing socket [%s]", service.Shutdown, err)
	}
}

func (s *Service) CheckRemoveServer(ipPort string) (removed bool, queried bool) {
	removed = false
	queried = false
	svr := s.ServerList.LocalServerList[ipPort]
	addr := svr.Server.Address.IP.String()

	if svr.IsExpired(s.Config.Values.Service.ServerTTL.Duration) {
		err := svr.Query()
		queried = true

		if err != nil {
			s.Lock()
			s.IPServiceCount[addr]--

			if s.IPServiceCount[addr] <= 0 {
				delete(s.IPServiceCount, addr)
			}

			s.Logs.Master.Logf("removing server %s, last seen: %s, new count for ip: %d", ipPort, svr.LastSeen.Format(time.Stamp), s.IPServiceCount[addr])
			s.ServerList.Remove(ipPort)
			s.Masters.Remove(ipPort)
			s.Unlock()

			removed = true

			return
		}

		svr.LastSeen = time.Now()
		svr.SolicitedTime = time.Now()
	}

	return
}

func (s *Service) RegisterExternalServerList(servers map[string]*server.Server) (errs []error) {
	s.Logs.Master.Logf("registering %d servers from external list", len(servers))

	for k, v := range servers {
		// only add servers we don't already know about
		if _, ok := s.ServerList.LocalServerList[k]; !ok {
			s.registerHeartbeat(v.Address, k)
		}
	}

	return
}

func (s *Service) RegisterExternalServer(ipPort string) error {
	if _, ok := s.ServerList.LocalServerList[ipPort]; ok {
		// only query new servers
		addr, err := net.ResolveUDPAddr("udp", ipPort)
		if err != nil {
			return err
		}

		s.registerHeartbeat(addr, ipPort)
	}

	return nil
}

func (s *Service) serveMaster(addr *net.UDPAddr, buf []byte) {
	// we use an ip-port combo as a unique identifier
	ipPort := fmt.Sprintf("%s:%d", addr.IP.String(), addr.Port)

	// parse packet
	p := protocol.NewPacket()
	err := p.UnmarshalBinary(buf)

	if err != nil {
		switch err {
		case protocol.ErrorUnknownPacketVersion:
			s.Logs.Master.ServerAlertf(ipPort, "Unknown protocol number")
		case protocol.ErrorEmptyPacket:
			s.Logs.Master.ServerAlertf(ipPort, "Empty packet received")
		default:
			s.Logs.Master.ServerAlertf(ipPort, "Error %s while parsing packet", err)
		}

		return
	}

	isBanned := false

	for _, v := range s.Config.ParsedBannedNets {
		if v.Contains(addr.IP) {
			s.Logs.Banned.ServerAlertf(addr.IP.String(), "Received a %s packet from banned host", p.Type.String())

			if p.Type != protocol.PingInfoQuery {
				return
			}

			isBanned = true

			break
		}
	}

	switch p.Type {
	// server has sent in a heartbeat
	case protocol.MasterServerHeartbeat:
		s.registerHeartbeat(addr, ipPort)

	// client is requesting a server list
	case protocol.PingInfoQuery:
		if isBanned {
			s.sendBanned(addr, ipPort, p)
			return
		}

		s.Lock()
		s.DailyStats.UniqueUsers[ipPort] = true
		s.Unlock()
		s.sendList(addr, ipPort, p)

	default:
		s.Logs.Master.ServerAlertf(ipPort, "Received unsolicited packet type %s", p.Type.String())
	}
}

func (s *Service) registerHeartbeat(addr *net.UDPAddr, ipPort string) {
	s.Lock()

	q := darkstar.NewQuery(s.Options.Timeout, s.Options.Debug)
	q.Addresses = append(q.Addresses, ipPort)
	response, err := q.Servers()

	if len(err) > 0 || len(response) == 0 {
		s.Logs.Heartbeat.ServerAlertf(ipPort, "error during server verification [%s, %d]", err, len(response))
		s.Unlock()

		return
	}

	// only add a server to the list if it passes verification
	if !s.ServerList.Exist(ipPort) {
		s.ServerList.Add(addr, ipPort, s.Options)
	}

	s.ServerList.Update(ipPort, response[0])

	s.Unlock()

	s.registerPingInfo(addr, ipPort)
}

func (s *Service) registerPingInfo(addr *net.UDPAddr, ipPort string) {
	s.Lock()
	if !s.Masters.Exist(ipPort) {
		count := s.IPServiceCount[addr.IP.String()]
		if count+1 > s.Config.Values.Service.ServersPerIP {
			s.Logs.Registration.ServerAlertf(ipPort, "Rejecting additional server for IP - count: %d/%d", count, s.Config.Values.Service.ServersPerIP)
			s.Unlock()

			return
		}

		// log and add new
		s.Masters.Add(ipPort, addr, &s.pconn)
		count++
		s.Logs.Registration.ServerLogf(ipPort, "New Server for IP - total server count for IP: %d/%d", count, s.Config.Values.Service.ServersPerIP)
		s.IPServiceCount[addr.IP.String()] = count
	}

	LastSeen := s.Masters.UpdateServer(ipPort)
	s.Logs.Heartbeat.ServerLogf(ipPort, "Heartbeat - delta: %s", time.Since(LastSeen).String())
	s.Unlock()
}

func (s *Service) sendList(addr *net.UDPAddr, ipPort string, p *protocol.Packet) {
	m := s.Masters.Get(ipPort)
	output := m.GeneratePackets(s.Options, p.Key)

	for _, v := range output {
		_, err := s.pconn.WriteTo(v, addr)
		if err != nil {
			s.Logs.Master.ServerAlertf(ipPort, "error sending master list [%s]", err)
			return
		}
	}

	s.Logs.Master.ServerLogf(ipPort, "servers list sent")
}

func (s *Service) sendBanned(addr *net.UDPAddr, ipPort string, p *protocol.Packet) {
	m := s.Masters.Get("")
	output := m.GeneratePackets(s.Options, p.Key)

	for _, v := range output {
		_, err := s.pconn.WriteTo(v, addr)
		if err != nil {
			s.Logs.Banned.ServerAlertf(ipPort, "error sending master list [%s]", err)
			return
		}
	}

	s.Logs.Banned.ServerLogf(ipPort, "banned message sent")
}
