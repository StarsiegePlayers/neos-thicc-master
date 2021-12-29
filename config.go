package main

import (
	"crypto/rand"
	"encoding/base64"
	"net"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type ConfigurationService struct {
	*viper.Viper
	Values *Configuration

	Logger
	Service
}

const (
	DefaultConfigFileName = "mstrsvr.yaml"
	EnvPrefix             = "mstrsvr"
	EggURL                = "https://youtu.be/pY725Ya74VU"
)

func (c *ConfigurationService) Init(args map[InitArg]interface{}) error {
	c.Logger = Logger{
		Name: "config",
		ID:   ConfigServiceID,
	}

	c.Viper = viper.New()

	c.AddConfigPath(".")
	c.SetConfigName(EnvPrefix)

	c.SetEnvPrefix(EnvPrefix)
	c.AllowEmptyEnv(true)

	c.SetDefaults()

	c.OnConfigChange(func(in fsnotify.Event) {
		c.Log("configuration change detected, updating...")
		c.Rehash()
	})

	c.Rehash()
	c.Values.serviceRunning = true

	loggerInit(c.Values.Log.ConsoleColors)

	var err error
	if c.Values.HTTPD.Secrets.Authentication == "" || len(c.Values.HTTPD.Secrets.Authentication) < 64 {
		c.LogAlert("invalid http authentication secret detected, generating...")
		c.Values.HTTPD.Secrets.Authentication, err = c.GenerateSecureRandomASCIIString(64)
		if err != nil {
			c.LogAlert("error creating http authentication secret [%s]", err)
		}
	}

	if c.Values.HTTPD.Secrets.Refresh == "" || len(c.Values.HTTPD.Secrets.Refresh) < 64 {
		c.LogAlert("invalid http authentication refresh secret detected, generating...")
		c.Values.HTTPD.Secrets.Refresh, err = c.GenerateSecureRandomASCIIString(64)
		if err != nil {
			c.LogAlert("error creating http authentication refresh secret [%s]", err)
		}
	}

	c.WatchConfig()

	//exit early if we're just processing command line arguments
	if c.processCommandLine() {
		os.Exit(0)
	}
	return nil
}

func (c *ConfigurationService) Run() {
	//noop
}

func (c *ConfigurationService) Shutdown() {
	//noop
}

func (c *ConfigurationService) Rehash() {
	// load up the config file
	err := c.ReadInConfig()
	if _, configFileNotFound := err.(viper.ConfigFileNotFoundError); err != nil && configFileNotFound {
		c.LogAlert("file not found, creating...")
		err := c.WriteConfigAs(DefaultConfigFileName)
		if err != nil {
			c.LogAlert("unable to create config! [%s]", err)
			os.Exit(1)
		}
	} else if err != nil {
		c.LogAlert("error while reading config file [%s]", err)
	}

	// replace the in-memory config with a new one
	c.Values = new(Configuration)
	c.Values.Lock()
	defer c.Values.Unlock()
	err = c.Unmarshal(&c.Values, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			StringToCustomDurationHookFunc(),
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	))
	if err != nil {
		c.LogAlert("error unmarshalling config [%s]", err)
	}

	// normalize all admin-usernames to be lowercase
	for user, password := range c.Values.HTTPD.Admins {
		lowerUser := strings.ToLower(user)
		if user != lowerUser {
			c.Values.HTTPD.Admins[lowerUser] = password
			delete(c.Values.HTTPD.Admins, user)
		}
	}

	// pre-parse the banned networks
	c.Values.parsedBannedNets = make([]*net.IPNet, 0)
	for _, x := range c.Values.Service.Banned.Networks {
		_, network, err := net.ParseCIDR(x)
		if err != nil {
			c.LogAlert("unable to parse BannedNetwork %s, %s", x, err)
			os.Exit(1)
		}
		c.Values.parsedBannedNets = append(c.Values.parsedBannedNets, network)
	}

	// cache the external ip
	c.Values.externalIP = getExternalIP(c.Values.Advanced.Network.StunServers)

	// execute the main rehash callback function
	if c.Values.callbackFn != nil {
		go c.Values.callbackFn()
	}

	return
}

func (c *ConfigurationService) GenerateSecureRandomASCIIString(length int) (string, error) {
	result := make([]byte, length)
	_, err := rand.Read(result)
	if err != nil {
		return "", err
	}
	for i := 0; i < length; i++ {
		result[i] &= 0x7F
		for result[i] < 32 || result[i] == 127 {
			_, err = rand.Read(result[i : i+1])
			if err != nil {
				return "", err
			}
			result[i] &= 0x7F
		}
	}
	return base64.RawURLEncoding.EncodeToString(result), nil
}
