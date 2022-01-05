package httpd

import (
	"encoding/json"
	"net"
	"sort"
	"time"

	"github.com/StarsiegePlayers/neos-thicc-master/src/service"
	"github.com/StarsiegePlayers/neos-thicc-master/src/stun"

	"github.com/StarsiegePlayers/darkstar-query-go/v2/query"
)

type MasterQuery struct {
	*query.MasterQuery
	ServerCount int
}

func (s *Service) maintenanceMultiplayerServersCache() (cacheData *CacheResponse) {
	CacheResponses := make(map[string]*CacheResponse)
	rawGames := make([]*query.PingInfoQuery, 0)
	localizedGames := rawGames
	errors := make([]string, 0)
	masters := make([]*MasterQuery, 0)

	// MasterService should never be nil, but just in case
	if s.MasterService != nil {
		s.MasterService.Lock()

		for _, v := range s.MasterService.ServerList {
			rawGames = append(rawGames, v.PingInfoQuery)
		}

		s.MasterService.Unlock()
	}

	// skip if poll service isn't running
	if s.PollService != nil {
		s.PollService.Lock()
		for _, v := range s.PollService.PollMasterInfo.Errors {
			errors = append(errors, v.Error())
		}

		for _, v := range s.PollService.PollMasterInfo.Masters {
			masters = append(masters, &MasterQuery{
				MasterQuery: v,
				ServerCount: len(v.Servers),
			})
		}
		s.PollService.Unlock()
	}

	// skip if STUN service isn't running
	if s.STUNService != nil {
		for _, v := range s.STUNService.(*stun.Service).LocalAddresses {
			localizedGames = make([]*query.PingInfoQuery, 0)

			for _, game := range rawGames {
				ipString, portString, _ := net.SplitHostPort(game.Address)
				ip := net.ParseIP(ipString)

				if ip != nil && v.Contains(ip) {
					game.Address = v.IP.String() + ":" + portString
				}

				localizedGames = append(localizedGames, game)
			}

			// update the cache
			data := &ServerListData{
				RequestTime: time.Now(),
				Masters:     masters,
				Games:       localizedGames,
				Errors:      errors,
			}

			sort.Sort(data.Masters)
			sort.Sort(data.Games)

			jsonOut, err := json.Marshal(data)
			if err != nil {
				continue
			}

			CacheResponses[v.String()] = &CacheResponse{
				Response: jsonOut,
				Time:     data.RequestTime,
			}
		}

		localizedGames = make([]*query.PingInfoQuery, 0)

		for _, game := range rawGames {
			addressString, portString, _ := net.SplitHostPort(game.Address)

			if addressString == service.LocalhostAddress {
				game.Address = s.STUNService.Get() + ":" + portString
			}

			localizedGames = append(localizedGames, game)
		}
	}

	// update the cache
	data := &ServerListData{
		RequestTime: time.Now(),
		Masters:     masters,
		Games:       localizedGames,
		Errors:      errors,
	}

	sort.Sort(data.Masters)
	sort.Sort(data.Games)

	jsonOut, err := json.Marshal(data)
	if err != nil {
		s.Logs.HTTPD.LogAlertf("error marshalling api server list %s", err)
		return
	}

	CacheResponses[""] = &CacheResponse{
		Response: jsonOut,
		Time:     data.RequestTime,
	}

	s.Lock()
	s.cache[cacheMultiplayer] = CacheResponses
	s.Unlock()

	cacheData = CacheResponses[""]

	return
}

func (s *Service) clearThrottleCache() {
	cache := s.cache[cacheThrottle].(map[string]int)
	if len(cache) >= 1 {
		s.Logs.HTTPD.Logf("[maintenance] resetting throttle cache for %d clients", len(cache))
		s.cache[cacheThrottle] = make(map[string]int)
	}
}
