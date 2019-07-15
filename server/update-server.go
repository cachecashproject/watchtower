package server

import (
	"context"
	"database/sql"

	"github.com/cachecashproject/watchtower/database/models"
	"github.com/cachecashproject/watchtower/grpcmsg"
	"github.com/sirupsen/logrus"
)

// UpdateServer handles requests from watchtower clients
type UpdateServer struct {
	l  *logrus.Logger
	db *sql.DB
}

// NewUpdateServer creates a new update server state
func NewUpdateServer(l *logrus.Logger, db *sql.DB) (*UpdateServer, error) {
	return &UpdateServer{
		l:  l,
		db: db,
	}, nil
}

// CheckForUpdates handles a request from
func (u *UpdateServer) CheckForUpdates(ctx context.Context, req *grpcmsg.UpdateCheckRequest) (*grpcmsg.UpdateCheckResponse, error) {
	u.l.Infof("Got update request from %v", req.Pubkey)
	expected := map[string]string{}

	for _, image := range req.CurrentImages {
		version, err := models.FindVersion(ctx, u.db, image.Name)
		if err != nil {
			if err != sql.ErrNoRows {
				u.l.Error("Failed to query database: ", err)
			} else {
				u.l.Debugf("No expected version for image: %v (%v)", image.Name, image.Version)
			}
			continue
		}

		if version.Version == image.Version {
			u.l.Infof("No updates pending for image: %v (%v)", image.Name, image.Version)
		} else {
			u.l.Infof("Sending update to client: %v (%v -> %v)", image.Name, image.Version, version.Version)
			expected[image.Name] = version.Version
		}
	}
	u.l.Infof("Sending response to client, %d updates for %d images", len(expected), len(req.CurrentImages))

	return &grpcmsg.UpdateCheckResponse{
		// "cachecash/go-cachecash:dev": "sha256:b488870aeadcb51bd8719122ceeb7e09a40d4745d471c8fef22aeb8800040fb9",
		ExpectedImages: expected,
	}, nil
}
