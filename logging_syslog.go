//go:build unix

package main

import (
	"fmt"
	"log/syslog"

	log "github.com/sirupsen/logrus"
	logsyslog "github.com/sirupsen/logrus/hooks/syslog"
)

var syslogFacilities = map[string]syslog.Priority{
	"0": syslog.LOG_LOCAL0,
	"1": syslog.LOG_LOCAL1,
	"2": syslog.LOG_LOCAL2,
	"3": syslog.LOG_LOCAL3,
	"4": syslog.LOG_LOCAL4,
	"5": syslog.LOG_LOCAL5,
	"6": syslog.LOG_LOCAL6,
	"7": syslog.LOG_LOCAL7,
}

func init() {
	addSyslogHook = func(appname, local string) error {
		if appname == "" {
			return fmt.Errorf("missing appname parameter")
		}
		if local == "" {
			return fmt.Errorf("missing local parameter")
		}
		facility, ok := syslogFacilities[local]
		if !ok {
			return fmt.Errorf("invalid local(%s) for syslog", local)
		}

		// Empty network and address connects to the local syslog daemon
		hook, err := logsyslog.NewSyslogHook("", "", facility, appname)
		if err != nil {
			return fmt.Errorf("can't connect logger to syslog: %v", err)
		}
		log.AddHook(hook)
		return nil
	}
}
