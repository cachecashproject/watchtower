package main // import "github.com/cachecashproject/watchtower"

import (
	"flag"
	"log"
	"os"

	"github.com/cachecashproject/watchtower/common"
	"github.com/cachecashproject/watchtower/server"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	logLevelStr = flag.String("logLevel", "info", "Verbosity of log output")
	logCaller   = flag.Bool("logCaller", false, "Enable method name logging")
	logFile     = flag.String("logFile", "", "Path where file should be logged")
	configPath  = flag.String("config", "update-server.config.json", "Path to configuration file")
	traceAPI    = flag.String("trace", "", "Jaeger API for tracing")
)

func loadConfigFile(l *logrus.Logger, path string) (*server.ConfigFile, error) {
	conf := server.ConfigFile{}
	p := common.NewConfigParser(l, "update_server")
	err := p.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf.GrpcAddr = p.GetString("grpc_addr", ":4000")

	return &conf, nil
}

func main() {
	if err := mainC(); err != nil {
		if _, err := os.Stderr.WriteString(err.Error() + "\n"); err != nil {
			panic(err)
		}
		os.Exit(1)
	}
}

func mainC() error {
	flag.Parse()
	log.SetFlags(0)

	l := logrus.New()
	/*
		if err := common.ConfigureLogger(l, &common.LoggerConfig{
			LogLevelStr: *logLevelStr,
			LogCaller:   *logCaller,
			LogFile:     *logFile,
			Json:        true,
		}); err != nil {
			return errors.Wrap(err, "failed to configure logger")
		}
	*/
	l.Info("Starting watchtower update server version:unknown")

	cf, err := loadConfigFile(l, *configPath)
	if err != nil {
		return errors.Wrap(err, "failed to load configuration file")
	}

	u, err := server.NewUpdateServer(l)
	if err != nil {
		return nil
	}

	app, err := server.NewApplication(l, u, cf)
	if err != nil {
		return errors.Wrap(err, "failed to create update server application")
	}

	if err := common.RunStarterShutdowner(l, app); err != nil {
		return err
	}
	return nil
}
