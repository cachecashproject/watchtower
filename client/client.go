package client

import (
	"context"

	"github.com/cachecashproject/watchtower/common"
	"github.com/cachecashproject/watchtower/container"
	"github.com/cachecashproject/watchtower/grpcmsg"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Client for our grpc watchtower update server
type Client struct {
	l          *logrus.Logger
	grpcClient grpcmsg.NodeUpdateClient
}

// NewUpdateClient creates an Client to query for updates
func NewUpdateClient(l *logrus.Logger, addr string, enableInsecureTransport bool) (*Client, error) {
	// XXX: Should not create a new connection for each attempt.
	l.Info("dialing update service: ", addr)

	var conn *grpc.ClientConn
	var err error

	if enableInsecureTransport {
		conn, err = common.GRPCDialInsecureTransport(addr)
	} else {
		conn, err = common.GRPCDial(addr)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial update service")
	}

	grpcClient := grpcmsg.NewNodeUpdateClient(conn)

	return &Client{
		l:          l,
		grpcClient: grpcClient,
	}, nil
}

// CheckForUpdates submits a list of running containers to check for pending updates
func (cl *Client) CheckForUpdates(containers []container.Container, pubkey string) (map[string]string, error) {
	ctx := context.Background()

	images := []*grpcmsg.ContainerImage{}

	for _, c := range containers {
		images = append(images, &grpcmsg.ContainerImage{
			Name:    c.ImageName(),
			Version: c.ImageID(),
		})
	}

	resp, err := cl.grpcClient.CheckForUpdates(ctx, &grpcmsg.UpdateCheckRequest{
		Pubkey:        pubkey,
		CurrentImages: images,
	})
	if err != nil {
		return nil, err
	}

	return resp.ExpectedImages, nil
}
