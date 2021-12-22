package main

import (
	"fmt"
	"log"

	"github.com/logrusorgru/aurora"
	"github.com/mattn/go-colorable"
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
	componentColors["startup"] = aurora.MagentaFg
	componentColors["shutdown"] = aurora.MagentaFg
	componentColors["maintenance"] = aurora.BrightFg | aurora.GreenFg
	componentColors["daily-maintenance"] = aurora.BrightFg | aurora.GreenFg

	componentColors["master"] = aurora.BlueFg
	componentColors["config"] = aurora.BrightFg | aurora.YellowFg

	componentColors["httpd"] = aurora.CyanFg
	componentColors["httpd-router"] = aurora.CyanFg

	componentColors["poll"] = aurora.YellowFg

	componentColors["default"] = aurora.WhiteFg
}

type Logger struct {
	Name   string
	LogTag string
}

func (c *Logger) Log(format string, args ...interface{}) {
	color, ok := componentColors[c.LogTag]
	if !ok {
		color = componentColors["default"]
	}
	s := fmt.Sprintf("{%s}: %s\n", au.Colorize(c.LogTag, color), au.Colorize(format, color))
	log.Printf(s, args...)
}

func (c *Logger) LogAlert(format string, args ...interface{}) {
	color, ok := componentColors[c.LogTag]
	if !ok {
		color = componentColors["default"]
	}
	s := fmt.Sprintf("{%s}: %s %s\n", au.Colorize(c.LogTag, color), au.Red("!"), au.Yellow(format))
	log.Printf(s, args...)
}

func (c *Logger) serverColor(input string) uint8 {
	o := byte(0)
	for _, c := range input {
		o += byte(c)
	}
	return (((o % 36) * 36) + (o % 6) + 16) % 255
}

func (c *Logger) ServerLog(server string, format string, args ...interface{}) {
	color := c.serverColor(server)
	s := fmt.Sprintf("[%s]: %s\n", au.Index(color, server), au.Index(color, format))
	log.Printf(s, args...)
}

func (c *Logger) ServerAlert(server string, format string, args ...interface{}) {
	color := c.serverColor(server)
	s := fmt.Sprintf("[%s]: %s %s\n", au.Index(color, server), au.Red("!"), au.Yellow(format))
	log.Printf(s, args...)
}
