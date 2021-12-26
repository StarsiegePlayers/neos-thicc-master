package main

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"net"
	"os"
)

const (
	DefaultConfigFileName = "mstrsvr.yaml"
	EnvPrefix             = "mstrsvr"
	EggURL                = "https://youtu.be/pY725Ya74VU"
)

func configInit() (config *Configuration) {
	logger := Logger{
		Name: "config",
		ID:   ConfigServiceID,
	}
	v := viper.New()
	v.AddConfigPath(".")
	v.SetConfigName(EnvPrefix)

	v.SetEnvPrefix(EnvPrefix)
	v.AllowEmptyEnv(true)

	config.SetDefaults(v)

	v.OnConfigChange(func(in fsnotify.Event) {
		logger.Log("configuration change detected, updating...")
		config = rehashConfig(v, logger)
	})

	config = rehashConfig(v, logger)
	config.serviceRunning = true

	loggerInit(config.Log.ConsoleColors)

	var err error
	if config.HTTPD.Secrets.Authentication == "" || len(config.HTTPD.Secrets.Authentication) < 64 {
		logger.LogAlert("invalid http authentication secret detected, generating...")
		config.HTTPD.Secrets.Authentication, err = GenerateSecureRandomASCIIString(64)
		if err != nil {
			logger.LogAlert("error creating http authentication secret [%s]", err)
		}
	}

	if config.HTTPD.Secrets.Refresh == "" || len(config.HTTPD.Secrets.Refresh) < 64 {
		logger.LogAlert("invalid http authentication refresh secret detected, generating...")
		config.HTTPD.Secrets.Refresh, err = GenerateSecureRandomASCIIString(64)
		if err != nil {
			logger.LogAlert("error creating http authentication refresh secret [%s]", err)
		}
	}

	v.WatchConfig()

	return config
}

func rehashConfig(v *viper.Viper, component Logger) (config *Configuration) {
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

	if config.callbackFn != nil {
		go config.callbackFn()
	}

	return
}

func GenerateSecureRandomASCIIString(length int) (string, error) {
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
