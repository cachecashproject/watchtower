package client

import (
	"context"

	"github.com/cachecashproject/watchtower/common"
	"github.com/cachecashproject/watchtower/container"
	"github.com/cachecashproject/watchtower/grpcmsg"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Client for our grpc watchtower update server
type Client struct {
	l          *logrus.Logger
	grpcClient grpcmsg.NodeUpdateClient
}

// NewUpdateClient creates an Client to query for updates
func NewUpdateClient(l *logrus.Logger, addr string) (*Client, error) {
	// XXX: Should not create a new connection for each attempt.
	l.Info("dialing bootstrap service: ", addr)
	conn, err := common.GRPCDial(addr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial bootstrap service")
	}

	grpcClient := grpcmsg.NewNodeUpdateClient(conn)

	return &Client{
		l:          l,
		grpcClient: grpcClient,
	}, nil
}

// CheckForUpdates submits a list of running containers to check for pending updates
func (cl *Client) CheckForUpdates(containers []container.Container) (map[string]string, error) {
	ctx := context.Background()

	images := []*grpcmsg.ContainerImage{}

	for _, c := range containers {
		images = append(images, &grpcmsg.ContainerImage{
			Name:    c.ImageName(),
			Version: c.ImageID(),
		})
	}

	resp, err := cl.grpcClient.CheckForUpdates(ctx, &grpcmsg.UpdateCheckRequest{
		Pubkey:        "TODO-publickey",
		CurrentImages: images,
	})
	if err != nil {
		return nil, err
	}

	return resp.ExpectedImages, nil
}
