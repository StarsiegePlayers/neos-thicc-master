package main

import (
	"fmt"
)

var (
	ErrorInvalidArgument = fmt.Errorf("invalid argument")
)

type ServiceID int

const (
	DefaultService = ServiceID(iota)
	StartupServiceID
	ShutdownServiceID
	ConfigServiceID
	MasterServiceID
	MaintenanceServiceID
	DailyMaintenanceServiceID
	TemplateServiceID
	HTTPServiceID
	HTTPDRouterID
	PollServiceID
	STUNClientID
)

func (s ServiceID) String() string {
	names := []string{
		"default",
		"startup",
		"shutdown",
		"config",
		"master",
		"maintenance",
		"daily-maintenance",
		"template",
		"httpd",
		"httpd-router",
		"poll",
		"stun-client",
	}
	return names[s]
}

type InitArg int

const (
	InitArgConfig = InitArg(iota)
	InitArgServices
)

type Service interface {
	Init(args map[InitArg]interface{}) error
	Run()
	Rehash()
	Shutdown()
}
