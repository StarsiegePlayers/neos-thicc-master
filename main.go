package main

import (
	"embed"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/StarsiegePlayers/neos-thicc-master/src"
	"github.com/StarsiegePlayers/neos-thicc-master/src/config"
	"github.com/StarsiegePlayers/neos-thicc-master/src/log"
	"github.com/StarsiegePlayers/neos-thicc-master/src/service"
)

//go:generate go-winres make --arch "amd64,386,arm,arm64" --file-version=git-tag --product-version=git-tag

const startupTextWidth = 50

var (
	buildVersion string
	buildDate    string
	buildTime    string
	buildCommit  string
	buildRelease string

	//go:embed www-build/*
	wwwFS embed.FS
)

func main() {
	if !strings.Contains(strings.ToLower(buildRelease), "true") {
		buildVersion += "-debug"
	} else {
		buildVersion += " Release"
	}

	server := new(src.Server)
	_ = server.Init(&server.Services)

	configService := server.Services[service.Config].(*config.Service)
	configService.SetBuildInfo(&service.BuildInfo{
		Version: buildVersion,
		Date:    buildDate,
		Time:    buildTime,
		Commit:  buildCommit,
		Release: buildRelease,
		EmbedFS: &wwwFS,
	})

	// handle command line options, if we need to, exit early
	if exit := processCommandLine(configService); exit {
		return
	}

	// create our log instance and send our session information
	mainLog := server.Services[service.Log].(*log.Service).NewLogger(service.Main)
	showHeader(mainLog)

	// setup kill / rehash hooks
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go signalHandler(c, server, mainLog)

	// exit early if nothing is running
	if len(server.Services) == 0 {
		mainLog.LogAlertf("no services detected running! shutting down...")
		server.Shutdown()
	}

	// block until complete
	server.Run()

	// clean exit message
	mainLog.Logf("main thread terminating")
}

func showHeader(mainLog *log.Log) {
	mainLog.Logf(strings.Repeat("-", startupTextWidth))
	mainLog.Logf(mainLog.NCenter(startupTextWidth, "Neo's Dummy Thicc Master Server"))
	mainLog.Logf(mainLog.NCenter(startupTextWidth, fmt.Sprintf("Version %s", buildVersion)))
	mainLog.Logf(mainLog.NCenter(startupTextWidth, config.EggURL))
	mainLog.Logf(mainLog.NCenter(startupTextWidth, fmt.Sprintf("Built on [%s@%s]", buildDate, buildTime)))
	mainLog.Logf(strings.Repeat("-", startupTextWidth))
}

func signalHandler(c chan os.Signal, server *src.Server, mainLog *log.Log) {
	for server.IsRunning {
		sig := <-c
		mainLog.Logf("received [%s]", sig.String())

		switch sig {
		case os.Interrupt:
			fallthrough

		case syscall.SIGTERM:
			mainLog.Logf("shutdown initiated...")
			server.Shutdown()

		case syscall.SIGHUP:
			server.Rehash()
		}
	}
}
