package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unicode/utf8"
)

//go:generate go-winres make --arch "amd64,386,arm,arm64"

var (
	buildVersion = ""
	buildDate    = ""
	buildTime    = ""
	buildCommit  = ""
	buildRelease = ""
)

func init() {
	if !strings.Contains(buildRelease, "true") {
		buildVersion = buildVersion + "-debug"
	}
}

func main() {
	var err error
	startup := Logger{
		Name: "startup",
		ID:   StartupServiceID,
	}

	loggerInit(false)
	config := configInit()

	startup.Log(strings.Repeat("-", 50))
	startup.Log(NCenter(50, "Neo's Dummy Thicc Master Server"))
	startup.Log(NCenter(50, fmt.Sprintf("Version %s", buildVersion)))
	startup.Log(NCenter(50, EggURL))
	startup.Log(NCenter(50, fmt.Sprintf("Built on [%s@%s]", buildDate, buildTime)))
	startup.Log(strings.Repeat("-", 50))

	x, _ := json.Marshal(config)
	startup.Log(string(x))

	services := make(map[ServiceID]Service)

	services[MasterServiceID] = new(MasterService)
	services[MaintenanceServiceID] = new(MaintenanceService)
	services[DailyMaintenanceServiceID] = new(DailyMaintenanceService)
	services[TemplateServiceID] = new(TemplateService)

	if config.HTTPD.Enabled {
		services[HTTPServiceID] = new(HTTPDService)
	}

	if config.Poll.Enabled {
		services[PollServiceID] = new(PollService)
	}

	args := make(map[InitArg]interface{})
	args[InitArgServices] = &services
	args[InitArgConfig] = config
	for k, v := range services {
		err = v.Init(args)
		if err != nil {
			startup.LogAlert("service {%s} failed to initialize, removing from threads list - [%s]", k, err)
			delete(services, k)
			continue
		}
		go v.Run()
	}

	startup.Log("startup finished")

	// setup kill / rehash hooks
	shutdown := Logger{
		Name: "shutdown",
		ID:   ShutdownServiceID,
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	// exit early if nothing is running
	if len(services) <= 0 {
		startup.LogAlert("no services detected running! shutting down...")
		config.serviceRunning = false
		cancel()
	}

	rehash := Logger{
		Name: "rehash",
		ID:   DefaultService,
	}
	rehashCB := func() {
		rehash.Log("Reloading services")
		for _, v := range services {
			v.Rehash()
		}
	}
	config.callbackFn = rehashCB

	// os signal (control+c / sigkill / sigterm) watcher
	go func() {
		for config.serviceRunning == true {
			sig := <-c
			shutdown.Log("received [%s]", sig.String())
			switch sig {
			case os.Interrupt:
				fallthrough
			case os.Kill:
				fallthrough
			case syscall.SIGTERM:
				shutdown.Log("shutdown initiated...")
				config.Lock()
				config.serviceRunning = false
				config.Unlock()
				for _, v := range services {
					v.Shutdown()
				}
				break
			}
		}
		cancel()
	}()

	// wait for everything to finish before exiting main
	<-ctx.Done()
	shutdown.Log("process complete")
}

func NCenter(width int, s string) string {
	const half, space = 2, "\u0020"
	var b bytes.Buffer
	n := (width - utf8.RuneCountInString(s)) / half
	if n < 0 {
		n = 0
	}
	_, _ = fmt.Fprintf(&b, "%s%s", strings.Repeat(space, n), s)
	return b.String()
}
