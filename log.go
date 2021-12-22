package main

import (
	"github.com/logrusorgru/aurora"
	"github.com/mattn/go-colorable"
	"log"
)

var (
	au              aurora.Aurora
	componentColors = make(map[string]aurora.Color)
)

func init() {
	log.SetOutput(colorable.NewColorableStdout())
}

func loggerInit(colorLogs bool) {
	au = aurora.NewAurora(colorLogs)
	componentColors["startup"] = aurora.BrightFg | aurora.MagentaFg
	componentColors["shutdown"] = aurora.BrightFg | aurora.MagentaFg
	componentColors["master"] = aurora.BrightFg | aurora.CyanFg
	componentColors["config"] = aurora.BrightFg | aurora.YellowFg
	componentColors["maintenance"] = aurora.BrightFg | aurora.GreenFg
	componentColors["daily-maintenance"] = aurora.BrightFg | aurora.GreenFg
	componentColors["httpd"] = aurora.BrightFg | aurora.BlueFg
	componentColors["default"] = aurora.BrightFg | aurora.WhiteFg
}
