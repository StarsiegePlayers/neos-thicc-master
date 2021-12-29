package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
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
	configService := new(ConfigurationService)
	_ = configService.Init(nil)

	startup.Log(strings.Repeat("-", 50))
	startup.Log(startup.NCenter(50, "Neo's Dummy Thicc Master Server"))
	startup.Log(startup.NCenter(50, fmt.Sprintf("Version %s", buildVersion)))
	startup.Log(startup.NCenter(50, EggURL))
	startup.Log(startup.NCenter(50, fmt.Sprintf("Built on [%s@%s]", buildDate, buildTime)))
	startup.Log(strings.Repeat("-", 50))

	//x, _ := json.Marshal(configService.Values)
	//startup.Log(string(x))

	services := make(map[ServiceID]Service)

	services[MasterServiceID] = new(MasterService)
	services[MaintenanceServiceID] = new(MaintenanceService)
	services[DailyMaintenanceServiceID] = new(DailyMaintenanceService)
	services[TemplateServiceID] = new(TemplateService)

	if configService.Values.HTTPD.Enabled {
		services[HTTPServiceID] = new(HTTPDService)
	}

	if configService.Values.Poll.Enabled {
		services[PollServiceID] = new(PollService)
	}

	args := make(map[InitArg]interface{})
	args[InitArgServices] = &services
	args[InitArgConfig] = configService
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
		configService.Values.serviceRunning = false
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
	configService.Values.callbackFn = rehashCB

	// os signal (control+c / sigkill / sigterm) watcher
	go func() {
		for configService.Values.serviceRunning == true {
			sig := <-c
			shutdown.Log("received [%s]", sig.String())
			switch sig {
			case os.Interrupt:
				fallthrough
			case os.Kill:
				fallthrough
			case syscall.SIGTERM:
				shutdown.Log("shutdown initiated...")
				configService.Values.Lock()
				configService.Values.serviceRunning = false
				configService.Values.Unlock()
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
