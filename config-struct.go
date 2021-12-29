package main

import (
	"encoding/json"
	"errors"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
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

	parsedBannedNets []*net.IPNet
	externalIP       string
	serviceRunning   bool
	callbackFn       func()
}

type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}

// StringToCustomDurationHookFunc returns a DecodeHookFunc that converts
// strings to Duration.
func StringToCustomDurationHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(Duration{}) {
			return data, nil
		}

		// Convert it by parsing
		d, err := time.ParseDuration(data.(string))
		return Duration{d}, err
	}
}

func (c *ConfigurationService) SetDefaults() {
	c.SetDefault("Log.ConsoleColors", true)
	c.SetDefault("Log.File", "")
	c.SetDefault("Log.Components", []string{"*"})

	c.SetDefault("Service.Listen.IP", "")
	c.SetDefault("Service.Listen.Port", 29000)

	c.SetDefault("Service.ServerTTL", 5*time.Minute)

	c.SetDefault("Service.Hostname", "")
	c.SetDefault("Service.Templates.MOTD", "")
	c.SetDefault("Service.Templates.TimeFormat", "Y-m-d H:i:s T")
	c.SetDefault("Service.ID", 01)
	c.SetDefault("Service.ServersPerIP", 15)

	c.SetDefault("Service.Banned.Message", "You've been banned!")
	c.SetDefault("Service.Banned.Networks", []string{"224.0.0.0/4"})

	c.SetDefault("Poll.Enabled", false)
	c.SetDefault("Poll.Interval", 5*time.Minute)
	c.SetDefault("Poll.KnownMasters", []string{"master1.starsiegeplayers.com:29000", "master2.starsiegeplayers.com:29000", "master3.starsiegeplayers.com:29000"})

	c.SetDefault("HTTPD.Enabled", true)
	c.SetDefault("HTTPD.Listen.IP", "")
	c.SetDefault("HTTPD.Listen.Port", "")
	c.SetDefault("HTTPD.Admins", map[string]string{})
	c.SetDefault("HTTPD.MaxRequestsPerMinute", 15)

	c.SetDefault("Advanced.Verbose", false)
	c.SetDefault("Advanced.Maintenance.Interval", 60*time.Second)
	c.SetDefault("Advanced.Network.ConnectionTimeout", 2*time.Second)
	c.SetDefault("Advanced.Network.MaxPacketSize", 512)
	c.SetDefault("Advanced.Network.MaxBufferSize", 32768)
	c.SetDefault("Advanced.Network.StunServers", []string{"stun.l.google.com:19302", "stun1.l.google.com:19302", "stun2.l.google.com:19302", "stun3.l.google.com:19302", "stun4.l.google.com:19302"})
}

func (c *ConfigurationService) Write() error {
	c.Values.Lock()
	f := c.Values
	c.Set("Log.ConsoleColors", f.Log.ConsoleColors)
	c.Set("Log.File", f.Log.File)
	c.Set("Log.Components", f.Log.Components)

	c.Set("Service.Listen.IP", f.Service.Listen.IP)
	c.Set("Service.Listen.Port", f.Service.Listen.Port)

	c.Set("Service.ServerTTL", f.Service.ServerTTL)

	c.Set("Service.Hostname", f.Service.Hostname)
	c.Set("Service.Templates.MOTD", f.Service.Templates.MOTD)
	c.Set("Service.Templates.TimeFormat", f.Service.Templates.TimeFormat)
	c.Set("Service.ID", f.Service.ID)
	c.Set("Service.ServersPerIP", f.Service.ServersPerIP)

	c.Set("Service.Banned.Message", f.Service.Banned.Message)
	c.Set("Service.Banned.Networks", f.Service.Banned.Networks)

	c.Set("Poll.Enabled", f.Poll.Enabled)
	c.Set("Poll.Interval", f.Poll.Interval)
	c.Set("Poll.KnownMasters", f.Poll.KnownMasters)

	c.Set("HTTPD.Enabled", f.HTTPD.Enabled)
	c.Set("HTTPD.Listen.IP", f.HTTPD.Listen.IP)
	c.Set("HTTPD.Listen.Port", f.HTTPD.Listen.Port)
	c.Set("HTTPD.Admins", f.HTTPD.Admins)
	c.Set("HTTPD.Secrets.Authentication", f.HTTPD.Secrets.Authentication)
	c.Set("HTTPD.Secrets.Refresh", f.HTTPD.Secrets.Refresh)

	c.Set("Advanced.Verbose", f.Advanced.Verbose)
	c.Set("Advanced.Maintenance.Interval", f.Advanced.Maintenance.Interval)
	c.Set("Advanced.Network.ConnectionTimeout", f.Advanced.Network.ConnectionTimeout)
	c.Set("Advanced.Network.MaxPacketSize", f.Advanced.Network.MaxPacketSize)
	c.Set("Advanced.Network.MaxBufferSize", f.Advanced.Network.MaxBufferSize)
	c.Set("Advanced.Network.StunServers", f.Advanced.Network.StunServers)

	return c.WriteConfig()
}
