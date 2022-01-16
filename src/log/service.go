package log

import (
	"log"
	"os"
	"sync"

	"github.com/StarsiegePlayers/neos-thicc-master/src/service"
	"github.com/StarsiegePlayers/neos-thicc-master/src/service/file"

	"github.com/logrusorgru/aurora"
	"github.com/mattn/go-colorable"
)

type Service struct {
	sync.Mutex

	au              aurora.Aurora
	componentColors map[service.ID]aurora.Color
	log             struct {
		categories sync.Map
		fileName   string
		file       *log.Logger
		handle     *os.File
	}

	service.Interface
}

func (s *Service) Init(*map[service.ID]service.Interface) error {
	log.SetOutput(colorable.NewColorableStdout())

	aurora.NewAurora(false)

	s.componentColors = make(map[service.ID]aurora.Color)

	s.componentColors[service.Default] = aurora.WhiteFg

	s.componentColors[service.Main] = aurora.MagentaFg
	s.componentColors[service.Startup] = aurora.MagentaFg
	s.componentColors[service.Rehash] = aurora.MagentaFg
	s.componentColors[service.Shutdown] = aurora.MagentaFg

	s.componentColors[service.Config] = aurora.BrightFg | aurora.YellowFg
	s.componentColors[service.Log] = aurora.BrightFg | aurora.YellowFg

	s.componentColors[service.Master] = aurora.BlueFg
	s.componentColors[service.Poll] = aurora.YellowFg

	s.componentColors[service.Maintenance] = aurora.BrightFg | aurora.GreenFg
	s.componentColors[service.DailyMaintenance] = aurora.BrightFg | aurora.GreenFg
	s.componentColors[service.STUN] = aurora.BrightFg | aurora.GreenFg
	s.componentColors[service.Template] = aurora.BrightFg | aurora.GreenFg

	s.componentColors[service.HTTPD] = aurora.CyanFg
	s.componentColors[service.HTTPDRouter] = aurora.CyanFg

	return nil
}

func (s *Service) Status() service.LifeCycle {
	return service.Static
}

func (s *Service) NewLogger(id service.ID) *Log {
	return &Log{
		ID:         id,
		logService: s,
	}
}

func (s *Service) SetColors(enableColors bool) {
	s.au = aurora.NewAurora(enableColors)
}

func (s *Service) SetLogables(components []string) {
	s.Lock()

	if len(components) > 1 && components[0] != "*" {
		for _, v := range components {
			s.log.categories.Store(service.TagToID(v), true)
		}
	} else {
		for k := range service.List {
			s.log.categories.Store(k, true)
		}
	}

	s.Unlock()
}

func (s *Service) SetLogFile(logFileName string) (err error) {
	s.Mutex.Lock()

	if logFileName != "" && logFileName != s.log.fileName {
		s.log.handle, err = os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, file.UserReadWrite|file.GroupRead|file.OtherRead)
		if err != nil {
			return
		}

		s.log.file = log.New(s.log.handle, "", log.Ldate|log.Ltime)
	}

	s.log.fileName = logFileName

	s.Mutex.Unlock()

	return
}
