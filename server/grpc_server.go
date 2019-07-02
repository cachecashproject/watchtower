package server

import (
	"context"

	"github.com/cachecashproject/watchtower/grpcmsg"
)

type grpcUpdateServer struct {
	update *UpdateServer
}

var _ grpcmsg.NodeUpdateServer = (*grpcUpdateServer)(nil)

func (s *grpcUpdateServer) CheckForUpdates(ctx context.Context, req *grpcmsg.UpdateCheckRequest) (*grpcmsg.UpdateCheckResponse, error) {
	return s.update.CheckForUpdates(ctx, req)
}
