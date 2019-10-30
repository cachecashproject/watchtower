package server

import (
	"context"
	"net"

	"github.com/cachecashproject/watchtower/common"
	"github.com/cachecashproject/watchtower/grpcmsg"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Application is a wraper around the grpc service
type Application interface {
	common.StarterShutdowner
}

// ConfigFile holds the configuration of our server
type ConfigFile struct {
	GrpcAddr    string `json:"grpc_addr"`
	ControlAddr string `json:"control_addr"`
	Database    string `json:"database"`
}

type application struct {
	conf          *ConfigFile
	l             *logrus.Logger
	updateServer  *grpc.Server
	controlServer *grpc.Server
}

var _ Application = (*application)(nil)

// NewApplication creates a new grpc service
func NewApplication(l *logrus.Logger, u *UpdateServer, c *UpdateControl, conf *ConfigFile) (Application, error) {
	updateServer := common.NewGRPCServer()
	grpcmsg.RegisterNodeUpdateServer(updateServer, u)

	controlServer := common.NewGRPCServer()
	grpcmsg.RegisterUpdateControlServer(controlServer, c)

	return &application{
		l:             l,
		conf:          conf,
		updateServer:  updateServer,
		controlServer: controlServer,
	}, nil
}

func (a *application) Start() error {
	if err := Start("updateServer", a.updateServer, a.l, a.conf.GrpcAddr); err != nil {
		return errors.Wrap(err, "failed to start update server")
	}

	if err := Start("controlServer", a.controlServer, a.l, a.conf.ControlAddr); err != nil {
		return errors.Wrap(err, "failed to start update server")
	}

	return nil
}

func (a *application) Shutdown(ctx context.Context) error {
	a.controlServer.GracefulStop()
	a.updateServer.GracefulStop()
	return nil
}

// Start starts a grpc.Server -- some logging is provided via the tag and
// logrus arguments, and the addr is used for dialing to.
func Start(tag string, s *grpc.Server, l *logrus.Logger, addr string) error {
	l.Infof("%s - Start - enter", tag)

	grpcLis, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.Wrap(err, "failed to bind listener")
	}

	go func() {
		// This will block until we call `Stop`.
		if err := s.Serve(grpcLis); err != nil {
			l.WithError(err).Errorf("failed to serve %s(grpc)", tag)
		}
	}()

	l.Infof("%s - Start - exit", tag)
	return nil
}
