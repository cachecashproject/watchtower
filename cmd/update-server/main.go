package main // import "github.com/cachecashproject/watchtower"

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"time"

	"github.com/cachecashproject/watchtower/common"
	"github.com/cachecashproject/watchtower/database/migrations"
	"github.com/cachecashproject/watchtower/server"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
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
	conf.Database = p.GetString("database", "host=127.0.0.1 port=5432 user=postgres dbname=updates sslmode=disable")

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
	if err := common.ConfigureLogger(l, &common.LoggerConfig{
		LogLevelStr: *logLevelStr,
		LogCaller:   *logCaller,
		LogFile:     *logFile,
		Json:        true,
	}); err != nil {
		return errors.Wrap(err, "failed to configure logger")
	}
	l.Info("Starting watchtower update server version:unknown")

	cf, err := loadConfigFile(l, *configPath)
	if err != nil {
		return errors.Wrap(err, "failed to load configuration file")
	}

	if cf.Database == "" {
		return errors.New("database connection isn't configured")
	}

	db, err := sql.Open("postgres", cf.Database)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}

	deadline := time.Now().Add(5 * time.Minute)
	for {
		err = db.Ping()

		if err == nil {
			// connected successfully
			break
		} else if time.Now().Before(deadline) {
			// connection failed, try again
			l.Info("Connection failed, trying again shortly")
			time.Sleep(250 * time.Millisecond)
		} else {
			// connection failed too many times, giving up
			return errors.Wrap(err, "database ping failed")
		}
	}
	l.Info("connected to database")

	l.Info("applying migrations")
	n, err := migrate.Exec(db, "postgres", migrations.Migrations, migrate.Up)
	if err != nil {
		return errors.Wrap(err, "failed to apply migrations")
	}
	l.Infof("applied %d migrations", n)

	u, err := server.NewUpdateServer(l, db)
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
