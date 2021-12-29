package main

import (
	"time"

	"github.com/StarsiegePlayers/darkstar-query-go/v2/query"
)

type CacheResponse struct {
	Response []byte
	Time     time.Time
}

type ServerListData struct {
	RequestTime time.Time
	Masters     MastersByPing
	Games       PingInfoQueryByPing
	Errors      []string
}

type MastersByPing []*MasterQuery

func (m MastersByPing) Len() int           { return len(m) }
func (m MastersByPing) Less(i, j int) bool { return m[i].Ping < m[j].Ping }
func (m MastersByPing) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

type PingInfoQueryByPing []*query.PingInfoQuery

func (p PingInfoQueryByPing) Len() int           { return len(p) }
func (p PingInfoQueryByPing) Less(i, j int) bool { return p[i].Ping < p[j].Ping }
func (p PingInfoQueryByPing) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
