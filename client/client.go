package client

import (
	"context"
	"strings"

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
	grpcConn   *grpc.ClientConn
	grpcClient grpcmsg.NodeUpdateClient
}

// NewUpdateClient creates an Client to query for updates
func NewUpdateClient(l *logrus.Logger, addr string, enableInsecureTransport bool) (*Client, error) {
	// XXX: Should not create a new connection for each attempt.
	l.Info("Dialing update service: ", addr)

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
		grpcConn:   conn,
		grpcClient: grpcClient,
	}, nil
}

// CheckForUpdates submits a list of running containers to check for pending updates
func (cl *Client) CheckForUpdates(containers []container.Container, pubkey string) (map[string]string, error) {
	ctx := context.Background()

	images := []*grpcmsg.ContainerImage{}

	for _, c := range containers {
		// our fallback string, digest isn't guaranteed to be available
		version := "<unknown>"
		imageDigests := c.ImageDigests()

		if len(imageDigests) > 0 {
			parts := strings.Split(imageDigests[0], "@")
			version = parts[len(parts)-1]
		}

		images = append(images, &grpcmsg.ContainerImage{
			Name:    c.ImageName(),
			Version: version,
		})
	}

	resp, err := cl.grpcClient.CheckForUpdates(ctx, &grpcmsg.UpdateCheckRequest{
		Pubkey:        pubkey,
		CurrentImages: images,
	})
	if err != nil {
		return nil, err
	}
	cl.l.Infof("Received %d pending updates", len(resp.ExpectedImages))

	return resp.ExpectedImages, nil
}

// Close the grpc client
func (cl *Client) Close() error {
	return cl.grpcConn.Close()
}
