package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unicode/utf8"
)

var serviceRunning = true

//go:generate go-winres make --arch "amd64,386,arm,arm64"

var (
	VERSION = "0.1.0"
	DATE    = "2021/12/21"
	TIME    = "69:69:69"
	DEBUG   = "DEBUG"
)

func main() {
	var err error
	startup := Logger{
		Name:   "main",
		LogTag: "startup",
	}

	loggerInit(false)
	config := configInit()

	startup.Log(strings.Repeat("-", 50))
	startup.Log(NCenter(50, "Neo's Dummy Thicc Master Server"))
	startup.Log(NCenter(50, fmt.Sprintf("Version %s %s", VERSION, DEBUG)))
	startup.Log(NCenter(50, "https://youtu.be/pY725Ya74VU"))
	startup.Log(NCenter(50, fmt.Sprintf("Built on [%s@%s]", DATE, TIME)))
	startup.Log(strings.Repeat("-", 50))
	startup.Log("Hostname:  %s", config.Service.Hostname)
	startup.Log("MOTD:      %s", config.Service.MOTD)
	startup.Log("Server ID: %d", config.Service.ID)

	services := make(map[string]Service)

	services["master"] = new(MasterService)
	services["maintenance"] = new(MaintenanceService)
	services["daily-maintenance"] = new(DailyMaintenanceService)

	if config.HTTPD.Enabled {
		services["httpd"] = new(HTTPDService)
	}

	if config.Poll.Enabled {
		services["poll"] = new(PollService)
	}

	args := make(map[string]interface{})
	args["services"] = &services
	args["config"] = config
	for k, v := range services {
		err = v.Init(args)
		if err != nil {
			startup.LogAlert("service {%s} failed to initialize, removing from threads list - [%s]", k, err)
			delete(services, k)
		}
		go v.Run()
	}

	// setup kill / rehash hooks
	shutdown := Logger{
		Name:   "main",
		LogTag: "shutdown",
	}
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for serviceRunning == true {
			sig := <-c
			shutdown.Log("received [%s]", sig.String())
			switch sig {
			case os.Interrupt:
				fallthrough
			case os.Kill:
				fallthrough
			case syscall.SIGTERM:
				shutdown.Log("shutdown initiated...")
				serviceRunning = false
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
	_, _ = fmt.Fprintf(&b, "%s%s", strings.Repeat(space, n), s)
	return b.String()
}
