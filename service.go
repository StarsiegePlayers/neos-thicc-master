package main

import (
	"fmt"
	"log"
)

var (
	ErrorInvalidArgument = fmt.Errorf("invalid argument")
)

type Service interface {
	Init(args map[string]interface{}) error
	Run()
	Rehash()
	Shutdown()
}

type Component struct {
	Name   string
	LogTag string
}

func (c *Component) Log(format string, args ...interface{}) {
	color, ok := componentColors[c.LogTag]
	if !ok {
		color = componentColors["default"]
	}
	s := fmt.Sprintf("{%s}: %s\n", au.Colorize(c.LogTag, color), au.Colorize(format, color))
	log.Printf(s, args...)
}

func (c *Component) LogAlert(format string, args ...interface{}) {
	color, ok := componentColors[c.LogTag]
	if !ok {
		color = componentColors["default"]
	}
	s := fmt.Sprintf("{%s}: %s %s\n", au.Colorize(c.LogTag, color), au.Red("!"), au.Yellow(format))
	log.Printf(s, args...)
}

func (c *Component) serverColor(input string) uint8 {
	o := byte(0)
	for _, c := range input {
		o += byte(c)
	}
	return (((o % 36) * 36) + (o % 6) + 16) % 255
}

func (c *Component) ServerLog(server string, format string, args ...interface{}) {
	color := c.serverColor(server)
	s := fmt.Sprintf("[%s]: %s\n", au.Index(color, server), au.Index(color, format))
	log.Printf(s, args...)
}

func (c *Component) ServerAlert(server string, format string, args ...interface{}) {
	color := c.serverColor(server)
	s := fmt.Sprintf("[%s]: %s %s\n", au.Index(color, server), au.Red("!"), au.Yellow(format))
	log.Printf(s, args...)
}
