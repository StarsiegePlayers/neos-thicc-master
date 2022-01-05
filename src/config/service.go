package config

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net"
	"os"
	"strings"

	"github.com/StarsiegePlayers/neos-thicc-master/src/log"
	"github.com/StarsiegePlayers/neos-thicc-master/src/service"

	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Service struct {
	Values                  *Configuration
	RehashFn                func()
	UpdateRunningServicesFn func()
	ParsedBannedNets        []*net.IPNet
	BuildInfo               *service.BuildInfo

	logService *log.Service
	viper      *viper.Viper

	*log.Log
	service.Interface
}

const (
	DefaultConfigFileName  = "mstrsvr.yaml"
	EnvPrefix              = "mstrsvr"
	EggURL                 = "https://youtu.be/pY725Ya74VU"
	MinimumSecureKeyLength = 64
)

func (s *Service) Init(services *map[service.ID]service.Interface) error {
	s.logService = (*services)[service.Log].(*log.Service)
	s.Log = s.logService.NewLogger(service.Config)

	s.viper = viper.New()

	s.viper.AddConfigPath(".")
	s.viper.SetConfigName(EnvPrefix)

	s.viper.SetEnvPrefix(EnvPrefix)
	s.viper.AllowEmptyEnv(true)

	s.SetDefaults()

	s.viper.OnConfigChange(func(in fsnotify.Event) {
		s.Logf("configuration change detected, updating...")
		s.Rehash()
	})

	s.Rehash()
	s.logService.SetColors(s.Values.Log.ConsoleColors)
	s.logService.SetLogables(s.Values.Log.Components)

	err := s.logService.SetLogFile(s.Values.Log.File)
	if err != nil {
		s.LogAlertf("error opening log file %s [%s]", s.Values.Log.File, err)
	}

	if s.Values.HTTPD.Secrets.Authentication == "" || len(s.Values.HTTPD.Secrets.Authentication) < MinimumSecureKeyLength {
		s.LogAlertf("invalid http authentication secret detected, generating...")
		s.Values.HTTPD.Secrets.Authentication, err = s.GenerateSecureRandomASCIIString(MinimumSecureKeyLength)

		if err != nil {
			s.LogAlertf("error creating http authentication secret [%s]", err)
		}
	}

	if s.Values.HTTPD.Secrets.Refresh == "" || len(s.Values.HTTPD.Secrets.Refresh) < MinimumSecureKeyLength {
		s.LogAlertf("invalid http authentication refresh secret detected, generating...")
		s.Values.HTTPD.Secrets.Refresh, err = s.GenerateSecureRandomASCIIString(MinimumSecureKeyLength)

		if err != nil {
			s.LogAlertf("error creating http authentication refresh secret [%s]", err)
		}
	}

	s.viper.WatchConfig()

	return nil
}

func (s *Service) SetBuildInfo(info *service.BuildInfo) {
	s.BuildInfo = info
}

func (s *Service) Run() {
	// noop
}

func (s *Service) Shutdown() {
	// noop
}

func (s *Service) Get() (out string) {
	outB, _ := json.Marshal(s.Values)
	out = string(outB)

	return
}

func (s *Service) Rehash() {
	err := s.viper.ReadInConfig()
	if _, configFileNotFound := err.(viper.ConfigFileNotFoundError); err != nil && configFileNotFound {
		s.LogAlertf("file not found, creating...")
		err := s.viper.WriteConfigAs(DefaultConfigFileName)

		if err != nil {
			s.LogAlertf("unable to create config! [%s]", err)
			os.Exit(1)
		}
	} else if err != nil {
		s.LogAlertf("error while reading config file [%s]", err)
	}

	// replace the in-memory config with a new one
	s.Values = new(Configuration)
	s.Values.Lock()
	defer s.Values.Unlock()
	err = s.viper.Unmarshal(&s.Values, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			StringToCustomDurationHookFunc(),
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	))

	if err != nil {
		s.LogAlertf("error unmarshalling config [%s]", err)
	}

	// normalize all admin-usernames to be lowercase
	for user, password := range s.Values.HTTPD.Admins {
		lowerUser := strings.ToLower(user)
		if user != lowerUser {
			s.Values.HTTPD.Admins[lowerUser] = password
			delete(s.Values.HTTPD.Admins, user)
		}
	}

	// pre-parse the banned networks
	s.ParsedBannedNets = make([]*net.IPNet, 0)
	for _, x := range s.Values.Service.Banned.Networks {
		_, network, err := net.ParseCIDR(x)
		if err != nil {
			s.LogAlertf("unable to parse BannedNetwork %s, %s", x, err)
			os.Exit(1)
		}

		s.ParsedBannedNets = append(s.ParsedBannedNets, network)
	}

	// execute the main rehash callback function
	if s.RehashFn != nil {
		go s.RehashFn()
	}

	// update the running services
	if s.UpdateRunningServicesFn != nil {
		go s.UpdateRunningServicesFn()
	}
}

func (s *Service) GenerateSecureRandomASCIIString(length int) (string, error) {
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
