package stun

import (
	"bytes"
	"net"

	"github.com/StarsiegePlayers/neos-thicc-master/src/config"
	"github.com/StarsiegePlayers/neos-thicc-master/src/log"
	"github.com/StarsiegePlayers/neos-thicc-master/src/service"

	"github.com/pion/stun"
)

type Service struct {
	stunServers  []string
	cachedOutput string

	LocalAddresses []*net.IPNet

	services struct {
		Map    *map[service.ID]service.Interface
		Config *config.Service
	}

	log    *log.Log
	status service.LifeCycle

	service.Interface
	service.Rehashable
}

func (s *Service) Init(services *map[service.ID]service.Interface) error {
	s.log = (*services)[service.Log].(*log.Service).NewLogger(service.STUN)
	s.services.Map = services
	s.services.Config = (*services)[service.Config].(*config.Service)
	s.status = service.Static
	s.Rehash()

	return nil
}

func (s *Service) Status() service.LifeCycle {
	return s.status
}

func (s *Service) Rehash() {
	p := s.status
	s.status = service.Rehashing
	s.stunServers = s.services.Config.Values.Advanced.Network.StunServers
	s.LocalAddresses = s.generateUniqueLocalAddresses()
	s.cachedOutput = ""
	s.cachedOutput = s.Get("")
	s.status = p
}

func (s *Service) Get(string) string {
	if s.cachedOutput != "" {
		return s.cachedOutput
	}

	for _, stunServer := range s.stunServers {
		c, err := stun.Dial("udp4", stunServer)
		if err != nil {
			s.log.LogAlertf("dial error [%s]", err)
			continue
		}

		if err = c.Do(stun.MustBuild(stun.TransactionID, stun.BindingRequest), func(res stun.Event) {
			if res.Error != nil {
				s.log.LogAlertf("packet building error [%s]", res.Error)
				return
			}
			var xorAddr stun.XORMappedAddress
			if getErr := xorAddr.GetFrom(res.Message); getErr != nil {
				s.log.LogAlertf("xorAddress error [%s]", getErr)
				return
			}
			s.cachedOutput = xorAddr.IP.String()
		}); err != nil {
			s.log.LogAlertf("error during STUN-do [%s]", err)
			continue
		}

		if err := c.Close(); err != nil {
			s.log.LogAlertf("error closing STUN client [%s]", err)
		}

		if s.cachedOutput != "" {
			break
		}
	}

	return s.cachedOutput
}

func (s *Service) IsInLocalNets(host string) (bool, net.IP) {
	for _, v := range s.LocalAddresses {
		if v.Contains(net.ParseIP(host)) {
			return true, v.IP
		}
	}
	return false, nil
}

func (s *Service) generateUniqueLocalAddresses() (output []*net.IPNet) {
	addressList := make([]*net.IPNet, 0)

	ifaces, err := net.Interfaces()
	if err != nil {
		return
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		for k := range addrs {
			if addrs[k].(*net.IPNet).IP.To4() != nil {
				addressList = append(addressList, addrs[k].(*net.IPNet))
			}
		}
	}

	output = make([]*net.IPNet, 0)

	for k, v := range addressList {
		matchFound := false

		for k2, v2 := range addressList {
			if k == k2 || k < k2 {
				continue
			}

			matchFound = bytes.Equal(v.Mask, v2.Mask) && v.Contains(v2.IP) && v2.Contains(v.IP)
			if matchFound {
				break
			}
		}

		if !matchFound {
			output = append(output, v)
		}
	}

	return
}
