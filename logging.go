package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

var (
	logLevel  = flag.String("log.level", "info", "Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal]")
	logFormat = flag.String("log.format", "logger:stderr", `Set the log target and format. Example: "logger:syslog?appname=bob&local=7" or "logger:stdout?json=true"`)
)

// Platform specific log targets are registered from init() in their own files, and are nil where unsupported
var addSyslogHook func(appname, local string) error
var addEventlogHook func(name string, debugAsInfo bool) error

// configureLogging applies -log.level and -log.format, which use the syntax of the former prometheus/common/log package
func configureLogging(level, format string) error {
	l, err := log.ParseLevel(level)
	if err != nil {
		return err
	}
	log.SetLevel(l)

	u, err := url.Parse(format)
	if err != nil {
		return err
	}
	if u.Scheme != "logger" {
		return fmt.Errorf("invalid scheme %s", u.Scheme)
	}
	if u.Query().Get("json") == "true" {
		log.SetFormatter(&log.JSONFormatter{})
	}

	switch u.Opaque {
	case "syslog":
		if addSyslogHook == nil {
			return fmt.Errorf("system does not support syslog")
		}
		return addSyslogHook(u.Query().Get("appname"), u.Query().Get("local"))
	case "eventlog":
		if addEventlogHook == nil {
			return fmt.Errorf("system does not support eventlog")
		}
		debugAsInfo, _ := strconv.ParseBool(u.Query().Get("debugAsInfo"))
		return addEventlogHook(u.Query().Get("name"), debugAsInfo)
	case "stdout":
		log.SetOutput(os.Stdout)
	case "stderr":
		log.SetOutput(os.Stderr)
	default:
		return fmt.Errorf("unsupported logger %q", u.Opaque)
	}
	return nil
}
