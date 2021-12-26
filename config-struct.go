package main

import (
	"github.com/spf13/viper"
	"net"
	"sync"
	"time"
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
		ServerTTL    time.Duration
		ID           uint16
		ServersPerIP uint16
		Banned       struct {
			Networks []string
			Message  string
		}
	}

	Poll struct {
		Enabled      bool
		Interval     time.Duration
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
	}

	Advanced struct {
		Verbose bool
		Network struct {
			ConnectionTimeout time.Duration
			MaxPacketSize     uint16
			MaxBufferSize     uint16
			StunServers       []string
		}
		Maintenance struct {
			Interval time.Duration
		}
	}

	parsedBannedNets []*net.IPNet
	externalIP       string
	serviceRunning   bool
	callbackFn       func()
}

func (c *Configuration) SetDefaults(v *viper.Viper) {
	v.SetDefault("Log.ConsoleColors", true)
	v.SetDefault("Log.File", "")
	v.SetDefault("Log.Components", []string{"*"})

	v.SetDefault("Service.Listen.IP", "")
	v.SetDefault("Service.Listen.Port", 29000)

	v.SetDefault("Service.ServerTTL", 5*time.Minute)

	v.SetDefault("Service.Hostname", "")
	v.SetDefault("Service.Templates.MOTD", "")
	v.SetDefault("Service.Templates.TimeFormat", "Y-m-d H:i:s T")
	v.SetDefault("Service.ID", 01)
	v.SetDefault("Service.ServersPerIP", 15)

	v.SetDefault("Service.Banned.Message", "You've been banned!")
	v.SetDefault("Service.Banned.Networks", []string{"224.0.0.0/4"})

	v.SetDefault("Poll.Enabled", false)
	v.SetDefault("Poll.Interval", 5*time.Minute)
	v.SetDefault("Poll.KnownMasters", []string{"master1.starsiegeplayers.com:29000", "master2.starsiegeplayers.com:29000", "master3.starsiegeplayers.com:29000"})

	v.SetDefault("HTTPD.Enabled", true)
	v.SetDefault("HTTPD.Listen.IP", "")
	v.SetDefault("HTTPD.Listen.Port", "")
	v.SetDefault("HTTPD.Admins", map[string]string{})

	v.SetDefault("Advanced.Verbose", false)
	v.SetDefault("Advanced.Maintenance.Interval", 60*time.Second)
	v.SetDefault("Advanced.Network.ConnectionTimeout", 2*time.Second)
	v.SetDefault("Advanced.Network.MaxPacketSize", 512)
	v.SetDefault("Advanced.Network.MaxBufferSize", 32768)
	v.SetDefault("Advanced.Network.StunServers", []string{"stun.l.google.com:19302", "stun1.l.google.com:19302", "stun2.l.google.com:19302", "stun3.l.google.com:19302", "stun4.l.google.com:19302"})
}
