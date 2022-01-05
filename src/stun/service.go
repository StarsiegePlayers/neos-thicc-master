package stun

import (
	"github.com/StarsiegePlayers/neos-thicc-master/src/config"
	"github.com/StarsiegePlayers/neos-thicc-master/src/log"
	"github.com/StarsiegePlayers/neos-thicc-master/src/service"

	"github.com/pion/stun"
)

type Service struct {
	services      *map[service.ID]service.Interface
	configService *config.Service
	stunServers   []string
	cachedOutput  string

	*log.Log
	service.Interface
}

func (s *Service) Init(services *map[service.ID]service.Interface) error {
	s.Log = (*services)[service.Log].(*log.Service).NewLogger(service.STUN)
	s.configService = (*services)[service.Config].(*config.Service)
	s.Rehash()

	return nil
}

func (s *Service) Rehash() {
	s.stunServers = s.configService.Values.Advanced.Network.StunServers
	s.cachedOutput = ""
}

func (s *Service) Run() {
	// noop
}

func (s *Service) Shutdown() {
	// noop
}

func (s *Service) Get() string {
	if s.cachedOutput != "" {
		return s.cachedOutput
	}

	for _, stunServer := range s.stunServers {
		c, err := stun.Dial("udp4", stunServer)
		if err != nil {
			s.LogAlertf("dial error [%s]", err)
			continue
		}

		if err = c.Do(stun.MustBuild(stun.TransactionID, stun.BindingRequest), func(res stun.Event) {
			if res.Error != nil {
				s.LogAlertf("packet building error [%s]", res.Error)
				return
			}
			var xorAddr stun.XORMappedAddress
			if getErr := xorAddr.GetFrom(res.Message); getErr != nil {
				s.LogAlertf("xorAddress error [%s]", getErr)
				return
			}
			s.cachedOutput = xorAddr.IP.String()
		}); err != nil {
			s.LogAlertf("error during STUN-do [%s]", err)
			continue
		}

		if err := c.Close(); err != nil {
			s.LogAlertf("error closing STUN client [%s]", err)
		}

		if s.cachedOutput != "" {
			break
		}
	}

	return s.cachedOutput
}
