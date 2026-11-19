package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/svc/eventlog"
)

func init() {
	addEventlogHook = func(name string, debugAsInfo bool) error {
		if name == "" {
			return fmt.Errorf("missing name parameter")
		}

		eventLog, err := eventlog.Open(name)
		if err != nil {
			return fmt.Errorf("can't connect logger to eventlog: %v", err)
		}
		log.AddHook(&eventlogHook{log: eventLog, debugAsInfo: debugAsInfo})
		return nil
	}
}

// eventlogHook writes log entries to the Windows event log, in addition to the regular log output
type eventlogHook struct {
	log         *eventlog.Log
	debugAsInfo bool
}

func (h *eventlogHook) Levels() []log.Level {
	return log.AllLevels
}

func (h *eventlogHook) Fire(e *log.Entry) error {
	switch e.Level {
	case log.PanicLevel, log.FatalLevel, log.ErrorLevel:
		return h.log.Error(102, e.Message)
	case log.WarnLevel:
		return h.log.Warning(101, e.Message)
	case log.DebugLevel, log.TraceLevel:
		if h.debugAsInfo {
			return h.log.Info(100, e.Message)
		}
		return nil
	default:
		return h.log.Info(100, e.Message)
	}
}
