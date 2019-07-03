package server

import (
	"context"

	"github.com/cachecashproject/watchtower/grpcmsg"
	"github.com/sirupsen/logrus"
)

// UpdateServer handles requests from watchtower clients
type UpdateServer struct {
	l *logrus.Logger
	// db *sql.DB
}

// func NewUpdateServer(l *logrus.Logger, db *sql.DB) (*UpdateServer, error) {

// NewUpdateServer creates a new update server state
func NewUpdateServer(l *logrus.Logger) (*UpdateServer, error) {
	return &UpdateServer{
		l: l,
		// db: db,
	}, nil
}

// CheckForUpdates handles a request from
func (u *UpdateServer) CheckForUpdates(ctx context.Context, req *grpcmsg.UpdateCheckRequest) (*grpcmsg.UpdateCheckResponse, error) {
	u.l.Infof("Got update request from %v", req.Pubkey)

	total := len(req.CurrentImages)
	for i, image := range req.CurrentImages {
		u.l.Infof("[%d/%d] Found current image: %s, %s", i+1, total, image.Name, image.Version)
	}

	return &grpcmsg.UpdateCheckResponse{
		ExpectedImages: map[string]string{
			"cachecash/go-cachecash:dev": "sha256:b488870aeadcb51bd8719122ceeb7e09a40d4745d471c8fef22aeb8800040fb9",
		},
	}, nil
}
