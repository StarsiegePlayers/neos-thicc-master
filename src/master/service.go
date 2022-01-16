package master

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/StarsiegePlayers/neos-thicc-master/src/config"
	"github.com/StarsiegePlayers/neos-thicc-master/src/log"
	"github.com/StarsiegePlayers/neos-thicc-master/src/service"
	"github.com/StarsiegePlayers/neos-thicc-master/src/stats"
	"github.com/StarsiegePlayers/neos-thicc-master/src/stun"

	"github.com/StarsiegePlayers/darkstar-query-go/v2"
	"github.com/StarsiegePlayers/darkstar-query-go/v2/protocol"
	"github.com/StarsiegePlayers/darkstar-query-go/v2/query"
	"github.com/StarsiegePlayers/darkstar-query-go/v2/server"
)

const dummythicc = "dummythicc"

type Service struct {
	sync.Mutex

	Options        *protocol.Options
	IPServiceCount map[string]uint16
	ServerList     map[string]*ServerInfo

	pconn  net.PacketConn
	status service.LifeCycle

	services struct {
		Map      *map[service.ID]service.Interface
		Config   *config.Service
		Stats    *stats.Service
		Template service.Getable
		STUN     service.Getable
	}

	logs struct {
		Master       *log.Log
		Heartbeat    *log.Log
		Registration *log.Log
		Banned       *log.Log
	}

	masters struct {
		Main   *protocol.Master
		Banned *protocol.Master
	}

	service.Interface
	service.Maintainable
}

type ServerInfo struct {
	*query.PingInfoQuery
	*server.Server

	SolicitedTime time.Time
}

func (s *Service) Init(services *map[service.ID]service.Interface) (err error) {
	s.status = service.Stopped
	s.masters.Main = protocol.NewMaster()
	s.masters.Banned = protocol.NewMaster()
	s.Options = &protocol.Options{}
	s.ServerList = make(map[string]*ServerInfo)

	s.IPServiceCount = make(map[string]uint16)

	s.services.Map = services
	s.services.Config = (*s.services.Map)[service.Config].(*config.Service)
	s.services.Stats = (*s.services.Map)[service.Stats].(*stats.Service)
	s.services.Template = (*s.services.Map)[service.Template].(service.Getable)
	s.services.STUN = (*s.services.Map)[service.STUN].(service.Getable)

	s.logs.Master = (*s.services.Map)[service.Log].(*log.Service).NewLogger(service.Master)
	s.logs.Heartbeat = (*s.services.Map)[service.Log].(*log.Service).NewLogger(service.HeartbeatLog)
	s.logs.Registration = (*s.services.Map)[service.Log].(*log.Service).NewLogger(service.ServerRegistrationLog)
	s.logs.Banned = (*s.services.Map)[service.Log].(*log.Service).NewLogger(service.BannedTrafficLog)

	s.Rehash()

	return
}

func (s *Service) Run() {
	// start listening loop
	buf := make([]byte, s.services.Config.Values.Advanced.Network.MaxPacketSize)
	buf2 := make([]byte, s.services.Config.Values.Advanced.Network.MaxPacketSize)
	prevIPPort := ""
	s.status = service.Running
	addrPort := fmt.Sprintf("%s:%d", s.services.Config.Values.Service.Listen.IP, s.services.Config.Values.Service.Listen.Port)

	err := s.openPort(addrPort)
	if err != nil {
		s.logs.Master.LogAlertf("unable to bind to %s - [%w]", addrPort, err)
		return
	}

	for {
		n, addr, err := s.pconn.ReadFrom(buf)
		if err != nil {
			var e *net.OpError
			if errors.As(err, &e) && e.Op == "read" {
				s.logs.Master.LogAlertf("socket closed.")
			} else {
				s.logs.Master.LogAlertf("read error on socket [%w]", err)
			}

			break
		}

		// dedupe packets because udp
		if prevIPPort == addr.String() && bytes.Equal(buf[:n], buf2[:n]) {
			prevIPPort = ""
			continue
		}

		copy(buf2, buf)

		prevIPPort = addr.String()

		go s.serveMaster(&addr, buf[:n])
	}

	s.status = service.Stopped
	s.logs.Master.LogAlertf("service %s", s.status)
}

func (s *Service) Maintenance() {
	count := 0
	checked := 0
	fresh := 0

	for k := range s.ServerList {
		removed, queried := s.CheckRemoveServer(k)

		switch {
		case removed: // removed && *
			count++
		case queried: // !removed && queried
			checked++
		case !queried: // !removed && !queried
			fresh++
		}
	}

	s.logs.Master.Logf("{%s} removed %d stale servers, queried %d servers, %d servers still fresh\n", service.Maintenance, count, checked, fresh)
}

func (s *Service) Rehash() {
	p := s.status
	s.status = service.Rehashing

	s.Lock()
	s.logs.Master.Logf("{%s} Reading config", service.Rehash)

	s.masters.Main.MOTD = s.services.Template.Get("")
	s.masters.Main.MasterID = s.services.Config.Values.Service.ID
	s.masters.Main.CommonName = s.services.Config.Values.Service.Hostname
	s.masters.Main.MOTDJunk = dummythicc

	s.masters.Banned.MOTD = s.services.Config.Values.Service.Banned.Message
	s.masters.Banned.MasterID = s.services.Config.Values.Service.ID
	s.masters.Banned.CommonName = s.services.Config.Values.Service.Hostname
	s.masters.Banned.MOTDJunk = dummythicc

	s.Options.Debug = s.services.Config.Values.Advanced.Verbose
	s.Options.MaxServerPacketSize = s.services.Config.Values.Advanced.Network.MaxPacketSize
	s.Options.MaxNetworkPacketSize = s.services.Config.Values.Advanced.Network.MaxBufferSize
	s.Options.Timeout = s.services.Config.Values.Advanced.Network.ConnectionTimeout.Duration
	s.Unlock()

	s.status = p
}

func (s *Service) Shutdown() {
	s.status = service.Stopping

	if err := s.pconn.Close(); err != nil {
		s.logs.Master.LogAlertf("{%s} error while closing socket [%s]", service.Shutdown, err)
	}

	s.status = service.Stopped
}

func (s *Service) Status() service.LifeCycle {
	return s.status
}

func (s *Service) CheckRemoveServer(ipPort string) (removed bool, queried bool) {
	removed = false
	queried = false
	svr := s.ServerList[ipPort]
	addr := svr.Server.Address.(*net.UDPAddr)

	if svr.IsExpired(s.services.Config.Values.Service.ServerTTL.Duration) {
		err := svr.Query()
		queried = true

		if err != nil {
			s.Lock()
			s.IPServiceCount[addr.String()]--

			if s.IPServiceCount[addr.String()] <= 0 {
				delete(s.IPServiceCount, addr.String())
			}

			s.logs.Master.Logf("removing server %s, last seen: %s, new count for ip: %d", ipPort, svr.LastSeen.Format(time.Stamp), s.IPServiceCount[addr.String()])
			delete(s.ServerList, ipPort)
			delete(s.masters.Main.Servers, ipPort)
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
	s.logs.Master.Logf("registering %d servers from external list", len(servers))

	for k := range servers {
		// only add servers we don't already know about
		if _, ok := s.ServerList[k]; !ok {
			s.registerHeartbeat(&servers[k].Address, k)
		}
	}

	return
}

func (s *Service) RegisterExternalServer(ipPort string) error {
	if _, ok := s.ServerList[ipPort]; ok {
		// only query new servers
		addr, err := net.ResolveUDPAddr("udp", ipPort)
		if err != nil {
			return err
		}

		addr2 := net.Addr(addr)
		s.registerHeartbeat(&addr2, ipPort)
	}

	return nil
}

func (s *Service) openPort(addrPort string) (err error) {
	s.pconn, err = net.ListenPacket("udp", addrPort)
	if err != nil {
		return
	}

	s.logs.Master.Logf("now listening on [%s]", fmt.Sprintf("%s:%d", s.services.STUN.Get(""), s.services.Config.Values.Service.Listen.Port))

	return
}

func (s *Service) serveMaster(addr *net.Addr, buf []byte) {
	ipNet := (*addr).(*net.UDPAddr)

	// we use an ip-port combo as a unique identifier
	host, port, err := net.SplitHostPort((*addr).String())
	if err != nil {
		s.logs.Master.LogAlertf("Error parsing IP")
	}

	ipPort := fmt.Sprintf("%s:%s", host, port)

	// parse packet
	p := protocol.NewPacket()
	err = p.UnmarshalBinary(buf)

	if err != nil {
		switch {
		case errors.Is(err, protocol.ErrorUnknownPacketVersion):
			s.logs.Master.ServerAlertf(ipPort, "Unknown protocol number")
		case errors.Is(err, protocol.ErrorEmptyPacket):
			s.logs.Master.ServerAlertf(ipPort, "Empty packet received")
		default:
			s.logs.Master.ServerAlertf(ipPort, "Error while parsing packet [%w]", err)
		}

		return
	}

	isBanned := false

	for _, v := range s.services.Config.ParsedBannedNets {
		if v.Contains(ipNet.IP) {
			s.logs.Banned.ServerAlertf(ipNet.IP.String(), "Received a %s packet from banned host", p.Type.String())

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

		s.sendList(addr, ipPort, p)

	default:
		s.logs.Master.ServerAlertf(ipPort, "Received unsolicited packet type %s", p.Type.String())
	}
}

func (s *Service) registerHeartbeat(addr *net.Addr, ipPort string) {
	s.Lock()

	q := darkstar.NewQuery(s.Options.Timeout, s.Options.Debug)
	q.Addresses = append(q.Addresses, ipPort)
	response, err := q.Servers()

	if len(err) > 0 || len(response) == 0 {
		s.logs.Heartbeat.ServerAlertf(ipPort, "error during server verification [%s, %d]", err, len(response))
		s.Unlock()

		return
	}

	// only add a server to the list if it passes verification
	if _, ok := s.ServerList[ipPort]; !ok {
		s.ServerList[ipPort] = new(ServerInfo)
		s.ServerList[ipPort].PingInfoQuery = query.NewPingInfoQueryWithOptions(ipPort, s.Options)
		s.ServerList[ipPort].Server = new(server.Server)
		s.ServerList[ipPort].Server.Address = *addr
	}

	s.ServerList[ipPort].SolicitedTime = time.Now()
	s.ServerList[ipPort].LastSeen = time.Now()
	s.ServerList[ipPort].PingInfoQuery = response[0]
	go s.services.Stats.UpdatePlayerCountForServer(ipPort, response[0].PlayerCount)

	s.Unlock()

	s.registerPingInfo(addr, ipPort)
}

func (s *Service) registerPingInfo(addr *net.Addr, ipPort string) {
	s.Lock()
	ipNet := (*addr).(*net.UDPAddr)

	if _, ok := s.masters.Main.Servers[ipPort]; !ok {
		count := s.IPServiceCount[ipNet.IP.String()]
		if count+1 > s.services.Config.Values.Service.ServersPerIP {
			s.logs.Registration.ServerAlertf(ipPort, "Rejecting additional server for IP - count: %d/%d", count, s.services.Config.Values.Service.ServersPerIP)
			s.Unlock()

			return
		}

		// log and add new
		s.masters.Main.Servers[ipPort] = &server.Server{
			Address:    *addr,
			Connection: &s.pconn,
			LastSeen:   time.Now(),
		}

		count++
		s.logs.Registration.ServerLogf(ipPort, "New Server for IP - total server count for IP: %d/%d", count, s.services.Config.Values.Service.ServersPerIP)
		s.IPServiceCount[ipNet.IP.String()] = count
	}

	LastSeen := s.masters.Main.Servers[ipPort].LastSeen
	s.masters.Main.Servers[ipPort].LastSeen = time.Now()

	s.logs.Heartbeat.ServerLogf(ipPort, "Heartbeat - delta: %s", time.Since(LastSeen).String())
	s.Unlock()
}

func (s *Service) sendList(addr *net.Addr, ipPort string, p *protocol.Packet) {
	var laddr net.Addr

	for _, v := range s.services.STUN.(*stun.Service).LocalAddresses {
		if v.Contains((*addr).(*net.UDPAddr).IP) {
			laddr = net.Addr(v)
			break
		}
	}

	host, _, _ := net.SplitHostPort(ipPort)
	s.masters.Main.MOTD = s.services.Template.Get(host)
	output := s.masters.Main.GeneratePackets(s.Options, p.Key, &laddr)

	for _, v := range output {
		_, err := s.pconn.WriteTo(v, *addr)
		if err != nil {
			s.logs.Master.ServerAlertf(ipPort, "error sending master list [%s]", err)
			return
		}
	}

	s.logs.Master.ServerLogf(ipPort, "servers list sent")
}

func (s *Service) sendBanned(addr *net.Addr, ipPort string, p *protocol.Packet) {
	m := s.masters.Banned
	output := m.GeneratePackets(s.Options, p.Key, nil)

	for _, v := range output {
		_, err := s.pconn.WriteTo(v, *addr)
		if err != nil {
			s.logs.Banned.ServerAlertf(ipPort, "error sending master list [%s]", err)
			return
		}
	}

	s.logs.Banned.ServerLogf(ipPort, "banned message sent")
}
