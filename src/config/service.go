package config

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/StarsiegePlayers/neos-thicc-master/src/log"
	"github.com/StarsiegePlayers/neos-thicc-master/src/service"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Service struct {
	Values           *Configuration
	Startup          time.Time
	BuildInfo        *service.BuildInfo
	ParsedBannedNets []*net.IPNet

	Callback struct {
		Rehash            func()
		StartStopServices func()
		Shutdown          func()
		Restart           func()
	}

	logService     *log.Service
	viper          *viper.Viper
	rehashMutex    sync.Mutex
	status         service.LifeCycle
	rehashSentinel bool

	*log.Log
	service.Interface
	service.Rehashable
	service.Getable
}

const (
	DefaultConfigFileName  = "mstrsvr.yaml"
	EnvPrefix              = "mstrsvr"
	EggURL                 = "https://youtu.be/pY725Ya74VU"
	MinimumSecureKeyLength = 64
)

func (s *Service) Init(services *map[service.ID]service.Interface) error {
	s.status = service.Static
	s.Startup = time.Now()
	s.logService = (*services)[service.Log].(*log.Service)
	s.Log = s.logService.NewLogger(service.Config)

	s.viper = viper.New()

	s.viper.AddConfigPath(".")
	s.viper.SetConfigName(EnvPrefix)

	s.viper.SetEnvPrefix(EnvPrefix)
	s.viper.AllowEmptyEnv(true)

	s.SetDefaults()

	s.Rehash()
	s.logService.SetColors(s.Values.Log.ConsoleColors)
	s.logService.SetLogables(s.Values.Log.Components)

	err := s.logService.SetLogFile(s.Values.Log.File)
	if err != nil {
		s.LogAlertf("error opening log file %s [%w]", s.Values.Log.File, err)
	}

	return nil
}

func (s *Service) Status() service.LifeCycle {
	return s.status
}

func (s *Service) SetBuildInfo(info *service.BuildInfo) {
	s.BuildInfo = info
}

func (s *Service) Get(string) (out string) {
	outB, _ := json.Marshal(s.Values)
	out = string(outB)

	return
}

func (s *Service) Rehash() {
	p := s.status
	s.status = service.Rehashing
	initialRewrite := false

	err := s.viper.ReadInConfig()
	if err != nil {
		configFileNotFoundErr := viper.ConfigFileNotFoundError{}
		if errors.As(err, &configFileNotFoundErr) {
			s.LogAlertf("file not found, creating...")
			err := s.viper.WriteConfigAs(DefaultConfigFileName)

			if err != nil {
				s.LogAlertf("unable to create config! [%s]", err)
				os.Exit(1)
			}
		} else {
			s.LogAlertf("error while reading config file [%s]", err)
		}
	}

	// replace the in-memory config with a new one
	s.Values = new(Configuration)
	s.Values.Lock()
	err = s.viper.Unmarshal(&s.Values, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			StringToCustomDurationHookFunc(),
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	))

	if err != nil {
		s.LogAlertf("error unmarshalling config [%w]", err)
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
			s.LogAlertf("unable to parse BannedNetwork %s, %w", x, err)
			continue
		}

		s.ParsedBannedNets = append(s.ParsedBannedNets, network)
	}

	s.Values.Unlock()

	// ensure we have secure httpd secrets
	if s.Values.HTTPD.Secrets.Authentication == "" || len(s.Values.HTTPD.Secrets.Authentication) < MinimumSecureKeyLength {
		s.LogAlertf("invalid http authentication secret detected, generating...")

		s.Values.HTTPD.Secrets.Authentication, err = s.GenerateSecureRandomASCIIString(MinimumSecureKeyLength)
		if err != nil {
			s.LogAlertf("error creating http authentication secret [%s]", err)
		}

		initialRewrite = true
	}

	if s.Values.HTTPD.Secrets.Refresh == "" || len(s.Values.HTTPD.Secrets.Refresh) < MinimumSecureKeyLength {
		s.LogAlertf("invalid http authentication refresh secret detected, generating...")

		s.Values.HTTPD.Secrets.Refresh, err = s.GenerateSecureRandomASCIIString(MinimumSecureKeyLength)
		if err != nil {
			s.LogAlertf("error creating http authentication refresh secret [%s]", err)
		}

		initialRewrite = true
	}

	if initialRewrite {
		err = s.Write()
		if err != nil {
			s.LogAlertf("error writing config [%s]", err)
		}
	}

	// execute the main rehash callback function
	if s.Callback.Rehash != nil && !s.rehashSentinel {
		s.rehashMutex.Lock()
		s.rehashSentinel = true
		s.Callback.Rehash()
		s.rehashSentinel = false
		s.rehashMutex.Unlock()
	}

	// update the running services
	if s.Callback.StartStopServices != nil && !s.rehashSentinel {
		s.rehashMutex.Lock()
		s.rehashSentinel = true
		s.Callback.StartStopServices()
		s.rehashSentinel = false
		s.rehashMutex.Unlock()
	}

	s.status = p
}

// GenerateSecureRandomASCIIString generates a secure random ASCII string
// and is adapted from the following source:
// https://gist.github.com/denisbrodbeck/635a644089868a51eccd6ae22b2eb800#gistcomment-3719619
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
