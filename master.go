package main

import (
	"bytes"
	"fmt"
	"github.com/StarsiegePlayers/darkstar-query-go/v2/query"
	"net"
	"sync"
	"time"

	darkstar "github.com/StarsiegePlayers/darkstar-query-go/v2"
	"github.com/StarsiegePlayers/darkstar-query-go/v2/protocol"
	"github.com/StarsiegePlayers/darkstar-query-go/v2/server"
)

type MasterService struct {
	sync.Mutex
	Master         *protocol.Master
	BannedMaster   *protocol.Master
	Options        *protocol.Options
	IPServiceCount map[string]uint16
	ServerList     map[string]*ServerInfo
	Config         *Configuration

	pconn net.PacketConn

	DailyStats DailyStats

	Service
	Component
	Maintainable
	DailyMaintainable
}

type DailyStats struct {
	UniqueUsers map[string]bool
}

type ServerInfo struct {
	*query.PingInfoQuery
	*server.Server

	SolicitedTime time.Time
}

func NewMaster(config *Configuration) (thisMaster *MasterService, err error) {
	thisMaster = &MasterService{
		Config: config,
	}
	return
}

func (s *MasterService) Init(args map[string]interface{}) (err error) {
	s.Component = Component{
		Name:   "Master Service",
		LogTag: "master",
	}

	s.Master = protocol.NewMaster()
	s.BannedMaster = protocol.NewMaster()
	s.Options = &protocol.Options{}
	s.ServerList = make(map[string]*ServerInfo)
	s.DailyStats = DailyStats{
		UniqueUsers: make(map[string]bool),
	}

	var ok bool
	s.Config, ok = args["config"].(*Configuration)
	if !ok {
		s.LogAlert("config %s", ErrorInvalidArgument)
		return ErrorInvalidArgument
	}

	addrPort := fmt.Sprintf("%s:%d", s.Config.Service.Listen.IP, s.Config.Service.Listen.Port)
	s.pconn, err = net.ListenPacket("udp", addrPort)
	if err != nil {
		s.LogAlert("unable to bind to %s - [%s]", addrPort, err)
		return
	}
	externalIP := s.Config.externalIP
	externalAddrPort := fmt.Sprintf("%s:%d", externalIP, s.Config.Service.Listen.Port)
	s.Log("now listening on [%s]", externalAddrPort)

	return
}

func (s *MasterService) Run() {
	// start listening loop
	buf := make([]byte, s.Config.Advanced.Network.MaxPacketSize)
	buf2 := make([]byte, s.Config.Advanced.Network.MaxPacketSize)
	prevIPPort := ""
	for serviceRunning {
		n, addr, err := s.pconn.ReadFrom(buf)
		if err != nil {
			switch t := err.(type) {
			case *net.OpError:
				if t.Op == "read" {
					s.LogAlert("socket closed.")
				}
				continue
			default:
				s.LogAlert("read error on socket [%s]", err)
			}
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
}

func (s *MasterService) Maintenance() {
	count := 0
	for k, v := range s.ServerList {
		if v.IsExpired(s.Config.Service.ServerTTL) {
			s.Lock()
			s.Log("[maintenance] removing server %s, last seen: %s", v.String(), v.LastSeen.Format(time.Stamp))
			delete(s.ServerList, k)
			s.IPServiceCount[k]--
			if s.IPServiceCount[k] <= 0 {
				delete(s.IPServiceCount, k)
			}
			s.Unlock()
			count++
		}
	}
	s.Log("[maintenance] cleaned up %d stale servers\n", count)
}

func (s *MasterService) DailyMaintenance() {
	s.Lock()
	s.Log("[daily-maintenance] resetting daily user count, last count: %d users", len(s.DailyStats.UniqueUsers))
	s.DailyStats.UniqueUsers = make(map[string]bool)
	s.Unlock()
}

func (s *MasterService) Rehash() {
	s.Master.MOTD = s.Config.Service.MOTD
	s.Master.MasterID = s.Config.Service.ID
	s.Master.CommonName = s.Config.Service.Hostname

	s.BannedMaster.MOTD = s.Config.Service.Banned.Message
	s.BannedMaster.MasterID = s.Config.Service.ID
	s.BannedMaster.CommonName = s.Config.Service.Hostname

	s.Options.MaxServerPacketSize = s.Config.Advanced.Network.MaxPacketSize
}

func (s *MasterService) Shutdown() {
	err := s.pconn.Close()
	if err != nil {
		s.LogAlert("error while closing socket [%s]", err)
	}
	return
}

func (s *MasterService) serveMaster(addr *net.UDPAddr, buf []byte) {
	// we use an ip-port combo as a unique identifier
	ipPort := fmt.Sprintf("%s:%d", addr.IP.String(), addr.Port)

	// parse packet
	p := protocol.NewPacket()
	err := p.UnmarshalBinary(buf)
	if err != nil {
		switch err {
		case protocol.ErrorUnknownPacketVersion:
			s.ServerAlert(ipPort, "Unknown protocol number")
		case protocol.ErrorEmptyPacket:
			s.ServerAlert(ipPort, "Empty packet received")
		default:
			s.ServerAlert(ipPort, "Error %s while parsing packet", err)
		}
		return
	}

	isBanned := false
	for _, v := range s.Config.parsedBannedNets {
		if v.Contains(addr.IP) {
			s.ServerAlert(addr.IP.String(), "Received a %s packet from banned host", p.Type.String())
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
		s.registerHeartbeat(addr, ipPort, p)
		break

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
		break

	default:
		s.ServerAlert(ipPort, "Received unsolicited packet type %s", p.Type.String())
	}
}

func (s *MasterService) registerHeartbeat(addr *net.UDPAddr, ipPort string, p *protocol.Packet) {
	s.Lock()
	s.ServerList[ipPort].SolicitedTime = time.Now()
	s.Unlock()

	q := darkstar.NewQuery(2*time.Second, true)
	q.Addresses = append(q.Addresses, ipPort)
	response, err := q.Servers()
	if len(err) > 0 || len(response) <= 0 {
		s.ServerAlert(ipPort, "error during server verification [%s, %d]", err, len(response))
		return
	}

	s.registerPingInfo(addr, ipPort, p)

}

func (s *MasterService) registerPingInfo(addr *net.UDPAddr, ipPort string, p *protocol.Packet) {
	s.Lock()
	if _, exist := s.Master.Servers[ipPort]; !exist {
		count := s.IPServiceCount[addr.IP.String()]
		if uint16(count)+1 > s.Config.Service.ServersPerIP {
			s.ServerAlert(ipPort, "Rejecting additional server for IP - count: %d/%d", count, s.Config.Service.ServersPerIP)
			s.Unlock()
			return
		}

		// log and add new
		s.ServerLog(ipPort, "Heartbeat - New Server")
		s.Master.Servers[ipPort] = &server.Server{
			Address:    addr,
			Connection: &s.pconn,
			LastSeen:   time.Now(),
		}
		count++
		s.ServerLog(ipPort, "New Server for IP - total server count for IP: %d/%d", count, s.Config.Service.ServersPerIP)
		s.IPServiceCount[addr.IP.String()] = count
	}

	svr := s.Master.Servers[ipPort]
	s.ServerLog(ipPort, "Heartbeat - delta: %s", time.Now().Sub(svr.LastSeen).String())
	s.Master.Servers[ipPort].LastSeen = time.Now()
	s.Unlock()
}

func (s *MasterService) sendList(addr *net.UDPAddr, ipPort string, p *protocol.Packet) {
	output := s.Master.GeneratePackets(s.Options, p.Key)
	for _, v := range output {
		_, err := s.pconn.WriteTo(v, addr)
		if err != nil {
			s.ServerAlert(ipPort, "error sending master list [%s]", err)
			return
		}
	}
	s.ServerLog(ipPort, "servers list sent")
}

func (s *MasterService) sendBanned(addr *net.UDPAddr, ipPort string, p *protocol.Packet) {
	output := s.BannedMaster.GeneratePackets(s.Options, p.Key)
	for _, v := range output {
		_, err := s.pconn.WriteTo(v, addr)
		if err != nil {
			s.ServerAlert(ipPort, "error sending master list [%s]", err)
			return
		}
	}
	s.ServerLog(ipPort, "banned message sent")
}
