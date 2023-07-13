// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logging

import (
	"fmt"
	"io"
	stdlog "log"
	"os"
	"strings"

	isatty "github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sean-/conswriter"
)

var acceptedLogLevels = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
var acceptedLogFormat = []string{"HUMAN", "JSON"}

// SetupLogger sets the log level and outout format.
// Accepted levels are panic, fatal, error, warn, info and debug.
// Accepted formats are human or json.
func SetupLogger(level, format string) (err error) {

	if err = setLogFormat(strings.ToUpper(format)); err != nil {
		return err
	}

	if err = setLogLevel(strings.ToUpper(level)); err != nil {
		return err
	}

	return nil
}

func setLogLevel(level string) error {
	switch level {
	case "DEBUG":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "INFO":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "WARN":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "ERROR":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "FATAL":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	default:
		return fmt.Errorf("unsupported log level: %q (supported levels: %s)", level,
			strings.Join(acceptedLogLevels, " "))
	}
	return nil
}

func setLogFormat(format string) error {

	var logWriter io.Writer
	var zLog zerolog.Logger

	if isatty.IsTerminal(os.Stdout.Fd()) ||
		isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		logWriter = conswriter.GetTerminal()
	} else {
		logWriter = os.Stderr
	}

	switch format {
	case "HUMAN":
		w := zerolog.ConsoleWriter{
			Out:     logWriter,
			NoColor: true,
		}
		zLog = zerolog.New(w).With().Timestamp().Logger()
	case "JSON":
		zLog = zerolog.New(logWriter).With().Timestamp().Logger()
	default:
		return fmt.Errorf("unsupported log format: %q (supported formats: %s)", format,
			strings.Join(acceptedLogFormat, " "))
	}

	log.Logger = zLog
	stdlog.SetFlags(0)
	stdlog.SetOutput(zLog)

	return nil
}
