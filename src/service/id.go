package service

type ID int

type IDs []ID

func (m IDs) Len() int           { return len(m) }
func (m IDs) Less(i, j int) bool { return m[i] < m[j] }
func (m IDs) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

type serviceInfo struct {
	ID
	tag         string
	description string
}

const (
	Default = ID(iota)
	Main
	Logger
	Startup
	Shutdown
	Rehash
	Config
	STUN
	Template
	Master
	Poll
	Maintenance
	DailyMaintenance
	HTTPD
	HTTPDRouter
	HeartbeatLog
	ServerRegistrationLog
	BannedTrafficLog
)

var (
	List = map[ID]serviceInfo{
		Default:               {Default, "default", "Default Service"},
		Main:                  {Main, "main", "Main Application"},
		Logger:                {Logger, "logger", "Logging Service"},
		Startup:               {Startup, "startup", "Server Startup"},
		Shutdown:              {Shutdown, "shutdown", "Server Shutdown"},
		Rehash:                {Rehash, "rehash", "Rehashing Messages"},
		Config:                {Config, "config", "Configuration Service"},
		STUN:                  {STUN, "stun-client", "STUN Client"},
		Template:              {Template, "template", "Template Strings Service"},
		Master:                {Master, "master", "Master Service"},
		Poll:                  {Poll, "poll", "Peer Master Polling Service"},
		Maintenance:           {Maintenance, "maintenance", "Server Maintenance Service"},
		DailyMaintenance:      {DailyMaintenance, "daily-maintenance", "Daily Server Maintenance Service"},
		HTTPD:                 {HTTPD, "httpd", "HTTPD Service"},
		HTTPDRouter:           {HTTPDRouter, "httpd-router", "HTTPD Routing"},
		HeartbeatLog:          {HeartbeatLog, "heartbeat", "Server Heartbeats"},
		ServerRegistrationLog: {ServerRegistrationLog, "registration", "Server Registrations"},
		BannedTrafficLog:      {BannedTrafficLog, "banned", "Banned Client/Server traffic"},
	}

	ListByTag = map[string]serviceInfo{}
)

func init() {
	for _, v := range List {
		ListByTag[v.tag] = v
	}
}

func TagToID(tagIn string) ID {
	i, ok := ListByTag[tagIn]
	if !ok {
		return Default
	}

	return i.ID
}

func (s ID) String() string {
	e, ok := List[s]
	if !ok {
		return List[Default].tag
	}

	return e.tag
}

func (s ID) Description() string {
	e, ok := List[s]
	if !ok {
		return List[Default].description
	}

	return e.description
}
