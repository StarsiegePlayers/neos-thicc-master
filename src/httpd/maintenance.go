package httpd

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/StarsiegePlayers/darkstar-query-go/v2/query"
)

type MasterQuery struct {
	*query.MasterQuery
	ServerCount int
}

func (s *Service) maintenanceMultiplayerServersCache() (cacheData *CacheResponse, localCacheData *CacheResponse) {
	games := make([]*query.PingInfoQuery, 0)
	localGames := make([]*query.PingInfoQuery, 0)
	errors := make([]string, 0)
	masters := make([]*MasterQuery, 0)

	// MasterService should never be nil, but just in case
	if s.MasterService != nil {
		s.MasterService.Lock()

		for _, v := range s.MasterService.ServerList.ServerList {
			games = append(games, v.PingInfoQuery)
		}

		for _, v := range s.MasterService.ServerList.LocalServerList {
			localGames = append(localGames, v.PingInfoQuery)
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

	// update the cache
	data := &ServerListData{
		RequestTime: time.Now(),
		Masters:     masters,
		Games:       games,
		Errors:      errors,
	}

	localData := &ServerListData{
		RequestTime: data.RequestTime,
		Masters:     data.Masters,
		Games:       localGames,
		Errors:      data.Errors,
	}

	sort.Sort(data.Masters)
	sort.Sort(data.Games)
	sort.Sort(localData.Games)

	jsonOut, err := json.Marshal(data)
	if err != nil {
		s.Logs.HTTPD.LogAlertf("error marshalling api server list %s", err)
		return
	}

	localJSONOut, err := json.Marshal(localData)
	if err != nil {
		s.Logs.HTTPD.LogAlertf("error marshalling api local server list %s", err)
		return
	}

	cacheData = &CacheResponse{
		Response: jsonOut,
		Time:     data.RequestTime,
	}

	localCacheData = &CacheResponse{
		Response: localJSONOut,
		Time:     data.RequestTime,
	}

	s.Lock()
	s.cache[Multiplayer] = cacheData
	s.cache[LocalMultiplayer] = localCacheData
	s.Unlock()

	return
}

func (s *Service) clearThrottleCache() {
	cache := s.cache[Throttle].(map[string]int)
	if len(cache) >= 1 {
		s.Logs.HTTPD.Logf("[maintenance] resetting throttle cache for %d clients", len(cache))
		s.cache[Throttle] = make(map[string]int)
	}
}
