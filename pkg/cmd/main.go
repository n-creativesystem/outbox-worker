package cmd

import (
	"log/slog"
	"os"

	"github.com/n-creativesystem/outbox-worker/pkg/internal/logging"
	"github.com/spf13/pflag"
)

func Execute() {
	exitCode := 0
	cmd := rootCommand()
	if err := cmd.Execute(); err != nil {
		slog.With(logging.WithStack(err)).Error(err.Error())
		exitCode = 1
	}
	os.Exit(exitCode)
}

type mainArgs struct {
	configFilePath  string
	healthCheckAddr string
}

func (args *mainArgs) setpflag(flag *pflag.FlagSet) {
	flag.StringVarP(&args.configFilePath, "config", "c", "config.yaml", "configuration file path")
	flag.StringVar(&args.healthCheckAddr, "health-check-addr", ":8080", "health check address")
}
