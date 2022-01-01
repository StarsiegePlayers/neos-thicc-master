package log

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/StarsiegePlayers/neos-thicc-master/src/service"
)

const LoggingTextPadLength = 23

type Log struct {
	ID service.ID

	logService *Service
}

func (l *Log) serverColor(input string) uint8 {
	o := byte(0)
	for _, c := range input {
		o += byte(c)
	}

	// todo: see if this can be modified to eliminate dark, hard to see colors
	return (((o % 36) * 36) + (o % 6) + 16) % 255 //nolint:gomnd
}

func (l *Log) Logf(format string, args ...interface{}) {
	if key, ok := l.logService.logables.Load(l.ID); !ok || !key.(bool) {
		return
	}

	color, ok := l.logService.componentColors[l.ID]
	if !ok {
		color = l.logService.componentColors[service.Default]
	}

	lpad := strings.Repeat(" ", LoggingTextPadLength-(len(l.ID.String())))
	tag := fmt.Sprintf("%s%s |", lpad, l.logService.au.Colorize(l.ID.String(), color))
	s := fmt.Sprintf("%35s %s\n", tag, l.logService.au.Colorize(format, color))
	log.Printf(s, args...)

	if l.logService.logFile != nil {
		format = l.ID.String() + " |" + format
		l.logService.logFile.Printf(format, args...)
	}
}

func (l *Log) LogAlertf(format string, args ...interface{}) {
	if key, ok := l.logService.logables.Load(l.ID); !ok || !key.(bool) {
		return
	}

	color, ok := l.logService.componentColors[l.ID]
	if !ok {
		color = l.logService.componentColors[service.Default]
	}

	lpad := strings.Repeat(" ", LoggingTextPadLength-(len(l.ID.String())))
	tag := fmt.Sprintf("%s%s %s", lpad, l.logService.au.Colorize(l.ID.String(), color), l.logService.au.Red("!"))
	s := fmt.Sprintf("%44s %s\n", tag, l.logService.au.Yellow(format))
	log.Printf(s, args...)

	if l.logService.logFile != nil {
		format = l.ID.String() + " !" + format
		l.logService.logFile.Printf(format, args...)
	}
}

func (l *Log) ServerLogf(server string, format string, args ...interface{}) {
	if key, ok := l.logService.logables.Load(l.ID); !ok || !key.(bool) {
		return
	}

	color := l.serverColor(server)
	lpad := strings.Repeat(" ", LoggingTextPadLength-(len(server)+1))
	tag := fmt.Sprintf("%s[%s] |", lpad, l.logService.au.Index(color, server))
	s := fmt.Sprintf("%s {%s} %s\n", tag, l.logService.au.Index(color, l.ID.String()), l.logService.au.Index(color, format))
	log.Printf(s, args...)

	if l.logService.logFile != nil {
		format = "[" + server + "] | {" + l.ID.String() + "} " + format
		l.logService.logFile.Printf(format, args...)
	}
}

func (l *Log) ServerAlertf(server string, format string, args ...interface{}) {
	if key, ok := l.logService.logables.Load(l.ID); !ok || !key.(bool) {
		return
	}

	color := l.serverColor(server)
	lpad := strings.Repeat(" ", LoggingTextPadLength-(len(server)+1))
	tag := fmt.Sprintf("%s[%s] %s", lpad, l.logService.au.Index(color, server), l.logService.au.Red("!"))
	s := fmt.Sprintf("%44s {%s} %s\n", tag, l.logService.au.Index(color, l.ID.String()), l.logService.au.Index(color, format))
	log.Printf(s, args...)

	if l.logService.logFile != nil {
		format = "[" + server + "] ! {" + l.ID.String() + "} " + format
		l.logService.logFile.Printf(format, args...)
	}
}

func (l *Log) NCenter(width int, s string) string {
	var b bytes.Buffer

	const half, space = 2, "\u0020"

	n := (width - utf8.RuneCountInString(s)) / half
	if n < 0 {
		n = 0
	}

	_, _ = fmt.Fprintf(&b, "%s%s", strings.Repeat(space, n), s)

	return b.String()
}
