# conswriter

Console Writer is intended to be used by CLI apps as a writer for terminals and
automatically pipes output to a pager if desired.

## Sample Usage

In `main.go` (this sample uses [`spf13/cobra`](https://github.com/spf13/cobra)):

```go
package main

import "github.com/sean-/conswriter"

func main() {
	defer func() {
		p := conswriter.GetTerminal()
		p.Wait()
	}()

	if err := cmd.Execute(); err != nil {
		log.Error().Err(err).Msg("unable to run")
		os.Exit(1)
	}
}
```

In the rest of your program, logger, wherever you need to write to stdout/stderr:

```go
var Cmd = &cobra.Command{
  Use:    "test",
  Short:  "spew stuff to stdout",

  RunE: func(cmd *cobra.Command, args []string) error {
    cons := conswriter.GetTerminal()

    fmt.Fprintf(cons, "first line\n")
    for i := 0; i < 100; i++ {
      fmt.Fprintf(cons, "output\n")
      log.Info().Msg("output")
    }
    fmt.Fprintf(cons, "last line\n")
  },
}
```

Or wherever you initialize your logger,
assuming [`rs/zerolog`](https://github.com/rs/zerolog):

```go
const (
	// Use a log format that resembles time.RFC3339Nano but includes all trailing
	// zeros so that we get fixed-width logging.
	logTimeFormat = "2006-01-02T15:04:05.000000000Z07:00"
)

// Perform the initial configuration of zerolog
func init() {
  w := zerolog.ConsoleWriter{
    Out:     os.Stderr,
    NoColor: true,
  }
  zlog := zerolog.New(zerolog.SyncWriter(w)).With().Timestamp().Logger()

  zerolog.DurationFieldUnit = time.Microsecond
  zerolog.DurationFieldInteger = true
  zerolog.TimeFieldFormat = logTimeFormat
  zerolog.SetGlobalLevel(zerolog.InfoLevel)

  log.Logger = zlog

  stdlog.SetFlags(0)
  stdlog.SetOutput(zlog)

}
```

and when you actually configure your logger with your CLI options:

```go
func setupLoggerAfterCLIArgParsing() error {
    logLevel := zerolog.InfoLevel
    // Set your log level from CLI flags.
    zerolog.SetGlobalLevel(logLevel)

	var logWriter io.Writer
	if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		logWriter = conswriter.GetTerminal()
	} else {
		logWriter = os.Stderr
	}

	logFmt := "auto"
	if logFmt == "auto" {
      if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
        logFmt = "human"
	  } else {
        logFmt = "zerolog"
      }
	}

	var zlog zerolog.Logger
	switch logFmt {
	case "zerolog":
		zlog = zerolog.New(logWriter).With().Timestamp().Logger()
	case "human":
		useColor := viper.GetBool(config.KeyLogTermColor)
		w := zerolog.ConsoleWriter{
			Out:     logWriter,
			NoColor: !useColor,
		}
		zlog = zerolog.New(w).With().Timestamp().Logger()
	default:
		return fmt.Errorf("unsupported log format: %q", logFmt)
	}

	log.Logger = zlog

	stdlog.SetFlags(0)
	stdlog.SetOutput(zlog)
	stdLogger = &stdlog.Logger{}

	// In order to prevent random libraries from hooking the standard logger and
	// filling the logger with garbage, discard all log entries.  At debug level,
	// however, let it all through.
	if logLevel != zerolog.DebugLevel {
		stdLogger.SetOutput(ioutil.Discard)
	} else {
		stdLogger.SetOutput(zlog)
	}

	return nil
}
```

Configure accordingly depending on taste.
