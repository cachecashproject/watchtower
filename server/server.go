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
	GrpcAddr string `json:"grpc_addr"`
	Database string `json:"database"`
}

type application struct {
	l            *logrus.Logger
	updateServer *updateServer
}

var _ Application = (*application)(nil)

// NewApplication creates a new grpc service
func NewApplication(l *logrus.Logger, u *UpdateServer, conf *ConfigFile) (Application, error) {
	updateServer, err := newUpdateServer(l, u, conf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create bootstrap server")
	}

	return &application{
		l:            l,
		updateServer: updateServer,
	}, nil
}

func (a *application) Start() error {
	if err := a.updateServer.Start(); err != nil {
		return errors.Wrap(err, "failed to start bootstrap server")
	}
	return nil
}

func (a *application) Shutdown(ctx context.Context) error {
	if err := a.updateServer.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to shut down bootstrap server")
	}
	return nil
}

type updateServer struct {
	l          *logrus.Logger
	conf       *ConfigFile
	update     *UpdateServer
	grpcServer *grpc.Server
}

var _ common.StarterShutdowner = (*updateServer)(nil)

func newUpdateServer(l *logrus.Logger, u *UpdateServer, conf *ConfigFile) (*updateServer, error) {
	grpcServer := common.NewGRPCServer()
	grpcmsg.RegisterNodeUpdateServer(grpcServer, &grpcUpdateServer{update: u})

	return &updateServer{
		l:          l,
		conf:       conf,
		update:     u,
		grpcServer: grpcServer,
	}, nil
}

func (s *updateServer) Start() error {
	s.l.Info("updateServer - Start - enter")

	grpcLis, err := net.Listen("tcp", s.conf.GrpcAddr)
	if err != nil {
		return errors.Wrap(err, "failed to bind listener")
	}

	go func() {
		// This will block until we call `Stop`.
		if err := s.grpcServer.Serve(grpcLis); err != nil {
			s.l.WithError(err).Error("failed to serve updateServer(grpc)")
		}
	}()

	s.l.Info("updateServer - Start - exit")
	return nil
}

func (s *updateServer) Shutdown(ctx context.Context) error {
	// TODO: Should use `GracefulStop` until context expires, and then fall back on `Stop`.
	s.grpcServer.Stop()
	return nil
}
