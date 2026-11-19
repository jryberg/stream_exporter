package main

import (
	"os"
	"runtime"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
)

func resetLogging(t *testing.T) {
	t.Cleanup(func() {
		log.SetLevel(log.InfoLevel)
		log.SetOutput(os.Stderr)
		log.SetFormatter(&log.TextFormatter{})
		log.StandardLogger().ReplaceHooks(make(log.LevelHooks))
	})
}

func TestConfigureLogging(t *testing.T) {
	cases := []struct {
		name    string
		level   string
		format  string
		wantErr string
	}{
		{name: "defaults", level: "info", format: "logger:stderr"},
		{name: "debug to stdout", level: "debug", format: "logger:stdout"},
		{name: "invalid level", level: "verbose", format: "logger:stderr", wantErr: "not a valid logrus Level"},
		{name: "invalid scheme", level: "info", format: "file:stderr", wantErr: "invalid scheme file"},
		{name: "unknown target", level: "info", format: "logger:journald", wantErr: `unsupported logger "journald"`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resetLogging(t)

			err := configureLogging(c.level, c.format)
			if c.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("Expected error containing %q, got %v", c.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got := log.GetLevel().String(); got != c.level {
				t.Fatalf("Mismatch! Expected level %q, got %q", c.level, got)
			}
		})
	}
}

func TestConfigureLoggingJSON(t *testing.T) {
	resetLogging(t)

	if err := configureLogging("info", "logger:stdout?json=true"); err != nil {
		t.Fatal(err)
	}
	if _, ok := log.StandardLogger().Formatter.(*log.JSONFormatter); !ok {
		t.Fatalf("Expected JSON formatter, got %T", log.StandardLogger().Formatter)
	}
}

func TestConfigureLoggingPlatformTargets(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("syslog and eventlog availability is inverted on Windows")
	}

	// None of these reach the point of connecting to a syslog daemon
	cases := []struct {
		name    string
		format  string
		wantErr string
	}{
		{name: "eventlog unsupported", format: "logger:eventlog?name=stream_exporter", wantErr: "system does not support eventlog"},
		{name: "syslog without appname", format: "logger:syslog?local=7", wantErr: "missing appname parameter"},
		{name: "syslog without local", format: "logger:syslog?appname=stream_exporter", wantErr: "missing local parameter"},
		{name: "syslog with invalid local", format: "logger:syslog?appname=stream_exporter&local=9", wantErr: "invalid local(9) for syslog"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resetLogging(t)

			err := configureLogging("info", c.format)
			if err == nil || !strings.Contains(err.Error(), c.wantErr) {
				t.Fatalf("Expected error containing %q, got %v", c.wantErr, err)
			}
		})
	}
}
