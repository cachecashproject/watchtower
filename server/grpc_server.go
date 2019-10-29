package server

import (
	"context"

	"github.com/cachecashproject/watchtower/grpcmsg"
	"github.com/golang/protobuf/ptypes/empty"
)

type grpcControlServer struct {
	control *UpdateControl
}

var _ grpcmsg.UpdateControlServer = (*grpcControlServer)(nil)

func (s *grpcControlServer) SetLatestUpdate(ctx context.Context, req *grpcmsg.ContainerImage) (*empty.Empty, error) {
	return s.control.SetLatestUpdate(ctx, req)
}

type grpcUpdateServer struct {
	update *UpdateServer
}

var _ grpcmsg.NodeUpdateServer = (*grpcUpdateServer)(nil)

func (s *grpcUpdateServer) CheckForUpdates(ctx context.Context, req *grpcmsg.UpdateCheckRequest) (*grpcmsg.UpdateCheckResponse, error) {
	return s.update.CheckForUpdates(ctx, req)
}
