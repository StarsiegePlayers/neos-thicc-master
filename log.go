package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/mattn/go-colorable"
)

var (
	au              aurora.Aurora
	componentColors = make(map[ServiceID]aurora.Color)
)

func init() {
	log.SetOutput(colorable.NewColorableStdout())
}

func loggerInit(colorLogs bool) {
	au = aurora.NewAurora(colorLogs)
	componentColors[StartupServiceID] = aurora.MagentaFg
	componentColors[ShutdownServiceID] = aurora.MagentaFg

	componentColors[MaintenanceServiceID] = aurora.BrightFg | aurora.GreenFg
	componentColors[DailyMaintenanceServiceID] = aurora.BrightFg | aurora.GreenFg

	componentColors[MasterServiceID] = aurora.BlueFg
	componentColors[ConfigServiceID] = aurora.BrightFg | aurora.YellowFg

	componentColors[HTTPServiceID] = aurora.CyanFg
	componentColors[HTTPDRouterID] = aurora.CyanFg

	componentColors[PollServiceID] = aurora.YellowFg

	componentColors[DefaultService] = aurora.WhiteFg
}

type Logger struct {
	Name string
	ID   ServiceID
}

const LoggingTextPadLength = 23

func (c *Logger) Log(format string, args ...interface{}) {
	color, ok := componentColors[c.ID]
	if !ok {
		color = componentColors[DefaultService]
	}
	lpad := strings.Repeat(" ", LoggingTextPadLength-(len(c.Name)))
	tag := fmt.Sprintf("%s%s |", lpad, au.Colorize(c.Name, color))
	s := fmt.Sprintf("%35s %s\n", tag, au.Colorize(format, color))
	log.Printf(s, args...)
}

func (c *Logger) LogAlert(format string, args ...interface{}) {
	color, ok := componentColors[c.ID]
	if !ok {
		color = componentColors[DefaultService]
	}
	lpad := strings.Repeat(" ", LoggingTextPadLength-(len(c.Name)))
	tag := fmt.Sprintf("%s%s %s", lpad, au.Colorize(c.Name, color), au.Red("!"))
	s := fmt.Sprintf("%44s %s\n", tag, au.Yellow(format))
	log.Printf(s, args...)
}

func (c *Logger) serverColor(input string) uint8 {
	o := byte(0)
	for _, c := range input {
		o += byte(c)
	}
	// todo: see if this can be modified to eliminate dark, hard to see colors
	return (((o % 36) * 36) + (o % 6) + 16) % 255
}

func (c *Logger) ServerLog(server string, format string, args ...interface{}) {
	color := c.serverColor(server)
	lpad := strings.Repeat(" ", LoggingTextPadLength-(len(server)+1))
	tag := fmt.Sprintf("%s[%s] |", lpad, au.Index(color, server))
	s := fmt.Sprintf("%s %s\n", tag, au.Index(color, format))
	log.Printf(s, args...)
}

func (c *Logger) ServerAlert(server string, format string, args ...interface{}) {
	color := c.serverColor(server)
	lpad := strings.Repeat(" ", LoggingTextPadLength-(len(server)+1))
	tag := fmt.Sprintf("%s[%s] %s", lpad, au.Index(color, server), au.Red("!"))
	s := fmt.Sprintf("%44s %s\n", tag, au.Index(color, format))
	log.Printf(s, args...)
}
