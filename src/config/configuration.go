package config

import (
	"github.com/StarsiegePlayers/neos-thicc-master/src/service"
	"sync"
)

type Configuration struct {
	sync.Mutex

	Log struct {
		ConsoleColors bool
		File          string
		Components    []string
	}

	Service struct {
		Listen struct {
			IP   string
			Port uint16
		}
		Hostname  string
		Templates struct {
			MOTD       string
			TimeFormat string
		}
		ServerTTL    Duration
		ID           uint16
		ServersPerIP uint16
		Banned       struct {
			Networks []string
			Message  string
		}
	}

	Poll struct {
		Enabled      bool
		Interval     Duration
		KnownMasters []string
	}

	HTTPD struct {
		Enabled bool
		Listen  struct {
			IP   string
			Port uint16
		}
		Admins  map[string]string
		Secrets struct {
			Authentication string
			Refresh        string
		}
		MaxRequestsPerMinute int
	}

	Advanced struct {
		Verbose bool
		Network struct {
			ConnectionTimeout Duration
			MaxPacketSize     uint16
			MaxBufferSize     uint16
			StunServers       []string
		}
		Maintenance struct {
			Interval Duration
		}
	}
}

func (s *Service) SetDefaults() {
	components := make([]string, 0)
	for _, v := range service.List {
		components = append(components, v.Tag)
	}

	s.viper.SetDefault("Log.ConsoleColors", true)
	s.viper.SetDefault("Log.File", "")
	s.viper.SetDefault("Log.Components", components)

	s.viper.SetDefault("Service.Listen.IP", "")
	s.viper.SetDefault("Service.Listen.Port", 29000) //nolint:gomnd

	s.viper.SetDefault("Service.ServerTTL", "5m")

	s.viper.SetDefault("Service.Hostname", "")
	s.viper.SetDefault("Service.Templates.MOTD", "")
	s.viper.SetDefault("Service.Templates.TimeFormat", "Y-m-d H:i:s T")
	s.viper.SetDefault("Service.ID", 99)           //nolint:gomnd
	s.viper.SetDefault("Service.ServersPerIP", 30) //nolint:gomnd

	s.viper.SetDefault("Service.Banned.Message", "You've been banned!")
	s.viper.SetDefault("Service.Banned.Networks", []string{"224.0.0.0/4"})

	s.viper.SetDefault("Poll.Enabled", false)
	s.viper.SetDefault("Poll.Interval", "5m")
	s.viper.SetDefault("Poll.KnownMasters", []string{"master1.starsiegeplayers.com:29000", "master2.starsiegeplayers.com:29000", "master3.starsiegeplayers.com:29000"})

	s.viper.SetDefault("HTTPD.Enabled", true)
	s.viper.SetDefault("HTTPD.Listen.IP", "")
	s.viper.SetDefault("HTTPD.Listen.Port", "")
	s.viper.SetDefault("HTTPD.Admins", map[string]string{})
	s.viper.SetDefault("HTTPD.MaxRequestsPerMinute", 15) //nolint:gomnd

	s.viper.SetDefault("Advanced.Verbose", false)
	s.viper.SetDefault("Advanced.Maintenance.Interval", "1m")
	s.viper.SetDefault("Advanced.Network.ConnectionTimeout", "2s")
	s.viper.SetDefault("Advanced.Network.MaxPacketSize", 512)   //nolint:gomnd
	s.viper.SetDefault("Advanced.Network.MaxBufferSize", 32768) //nolint:gomnd
	s.viper.SetDefault("Advanced.Network.StunServers", []string{"stun.l.google.com:19302", "stun1.l.google.com:19302", "stun2.l.google.com:19302", "stun3.l.google.com:19302", "stun4.l.google.com:19302"})
}

func (s *Service) UpdateValues(c *Configuration) error {
	values := c
	s.Values = values

	return s.Write()
}

func (s *Service) setValues() {
	f := s.Values
	s.viper.Set("Log.ConsoleColors", f.Log.ConsoleColors)
	s.viper.Set("Log.File", f.Log.File)
	s.viper.Set("Log.Components", f.Log.Components)

	s.viper.Set("Service.Listen.IP", f.Service.Listen.IP)
	s.viper.Set("Service.Listen.Port", f.Service.Listen.Port)

	s.viper.Set("Service.ServerTTL", f.Service.ServerTTL)

	s.viper.Set("Service.Hostname", f.Service.Hostname)
	s.viper.Set("Service.Templates.MOTD", f.Service.Templates.MOTD)
	s.viper.Set("Service.Templates.TimeFormat", f.Service.Templates.TimeFormat)
	s.viper.Set("Service.ID", f.Service.ID)
	s.viper.Set("Service.ServersPerIP", f.Service.ServersPerIP)

	s.viper.Set("Service.Banned.Message", f.Service.Banned.Message)
	s.viper.Set("Service.Banned.Networks", f.Service.Banned.Networks)

	s.viper.Set("Poll.Enabled", f.Poll.Enabled)
	s.viper.Set("Poll.Interval", f.Poll.Interval)
	s.viper.Set("Poll.KnownMasters", f.Poll.KnownMasters)

	s.viper.Set("HTTPD.Enabled", f.HTTPD.Enabled)
	s.viper.Set("HTTPD.Listen.IP", f.HTTPD.Listen.IP)
	s.viper.Set("HTTPD.Listen.Port", f.HTTPD.Listen.Port)
	s.viper.Set("HTTPD.Admins", f.HTTPD.Admins)
	s.viper.Set("HTTPD.Secrets.Authentication", f.HTTPD.Secrets.Authentication)
	s.viper.Set("HTTPD.Secrets.Refresh", f.HTTPD.Secrets.Refresh)

	s.viper.Set("Advanced.Verbose", f.Advanced.Verbose)
	s.viper.Set("Advanced.Maintenance.Interval", f.Advanced.Maintenance.Interval)
	s.viper.Set("Advanced.Network.ConnectionTimeout", f.Advanced.Network.ConnectionTimeout)
	s.viper.Set("Advanced.Network.MaxPacketSize", f.Advanced.Network.MaxPacketSize)
	s.viper.Set("Advanced.Network.MaxBufferSize", f.Advanced.Network.MaxBufferSize)
	s.viper.Set("Advanced.Network.StunServers", f.Advanced.Network.StunServers)
}

func (s *Service) Write() (err error) {
	s.Values.Lock()
	s.setValues()
	err = s.viper.WriteConfig()
	s.Values.Unlock()

	return
}
