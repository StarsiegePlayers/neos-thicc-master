package main

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"net"
	"os"
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
		Hostname     string
		MOTD         string
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
		Admins map[string]string
	}

	Advanced struct {
		Network struct {
			MaxPacketSize uint16
			MaxBufferSize uint16
			StunServers   []string
		}
		Maintenance struct {
			Interval time.Duration
		}
	}

	parsedBannedNets []*net.IPNet
	externalIP       string
}

const (
	DefaultConfigFileName = "mstrsvr.yaml"
	EnvPrefix             = "mstrsvr"
)

func configInit() (config *Configuration) {
	component := Component{
		Name:   "config",
		LogTag: "config",
	}
	v := viper.New()
	v.AddConfigPath(".")
	v.SetConfigName(EnvPrefix)

	v.SetEnvPrefix(EnvPrefix)
	v.AllowEmptyEnv(true)

	v.SetDefault("Log.ConsoleColors", true)
	v.SetDefault("Log.File", "")
	v.SetDefault("Log.Components", []string{"*"})

	v.SetDefault("Service.Listen.IP", "")
	v.SetDefault("Service.Listen.Port", 29000)

	v.SetDefault("Service.ServerTTL", 5*time.Minute)
	v.SetDefault("Service.Hostname", "")
	v.SetDefault("Service.MOTD", "")
	v.SetDefault("Service.ID", 01)
	v.SetDefault("Service.ServersPerIP", 15)

	v.SetDefault("Service.Banned.Message", "You've been banned!")
	v.SetDefault("Service.Banned.Networks", []string{"224.0.0.0/4"})

	v.SetDefault("Poll.Enabled", true)
	v.SetDefault("Poll.Interval", 5*time.Minute)
	v.SetDefault("Poll.KnownMasters", []string{"master1.starsiegeplayers.com", "master2.starsiegeplayers.com", "master3.starsiegeplayers.com"})

	v.SetDefault("HTTPD.Enabled", true)
	v.SetDefault("HTTPD.Listen.IP", "")
	v.SetDefault("HTTPD.Listen.Port", "")
	v.SetDefault("HTTPD.Admins", map[string]string{})

	v.SetDefault("Advanced.Maintenance.Interval", 60*time.Second)
	v.SetDefault("Advanced.Network.MaxPacketSize", 512)
	v.SetDefault("Advanced.Network.MaxBufferSize", 32768)
	v.SetDefault("Advanced.Network.StunServers", []string{"stun.l.google.com:19302", "stun1.l.google.com:19302", "stun2.l.google.com:19302", "stun3.l.google.com:19302", "stun4.l.google.com:19302"})

	v.OnConfigChange(func(in fsnotify.Event) {
		component.Log("configuration change detected, updating...")
		config = rehashConfig(v, component)
	})
	v.WatchConfig()

	config = rehashConfig(v, component)

	loggerInit(config.Log.ConsoleColors)

	return config
}

func rehashConfig(v *viper.Viper, component Component) (config *Configuration) {
	err := v.ReadInConfig()
	if _, configFileNotFound := err.(viper.ConfigFileNotFoundError); err != nil && configFileNotFound {
		component.LogAlert("file not found, creating...")
		err := v.WriteConfigAs(DefaultConfigFileName)
		if err != nil {
			component.LogAlert("unable to create config! [%s]", err)
			os.Exit(1)
		}
	} else if err != nil {
		component.LogAlert("error while reading config file [%s]", err)
	}

	config = new(Configuration)
	config.Lock()
	defer config.Unlock()
	err = v.Unmarshal(&config)
	if err != nil {
		component.LogAlert("error unmarshalling config [%s]", err)
	}

	config.parsedBannedNets = make([]*net.IPNet, 0)
	for _, v := range config.Service.Banned.Networks {
		_, network, err := net.ParseCIDR(v)
		if err != nil {
			component.LogAlert("unable to parse BannedNetwork %s, %s", v, err)
			os.Exit(1)
		}
		config.parsedBannedNets = append(config.parsedBannedNets, network)
	}

	config.externalIP = getExternalIP(config.Advanced.Network.StunServers)

	return
}
