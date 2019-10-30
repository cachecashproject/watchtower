package server

import (
	"context"
	"database/sql"
	"strings"

	"github.com/cachecashproject/watchtower/database/models"
	"github.com/cachecashproject/watchtower/grpcmsg"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/opencontainers/go-digest"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/boil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UpdateControl handles requests from watchtower clients
type UpdateControl struct {
	l  *logrus.Logger
	db *sql.DB
}

// NewUpdateControl creates a new update server state
func NewUpdateControl(l *logrus.Logger, db *sql.DB) (*UpdateControl, error) {
	return &UpdateControl{
		l:  l,
		db: db,
	}, nil
}

// SetLatestUpdate sets the latest update that should be reflected when clients reach out to the CheckForUpdates method.
func (u *UpdateControl) SetLatestUpdate(ctx context.Context, req *grpcmsg.ContainerImage) (*empty.Empty, error) {
	for key, val := range map[string]string{"name": req.Name, "version": req.Version} {
		if strings.TrimSpace(val) == "" {
			return nil, status.Errorf(codes.FailedPrecondition, "%v is empty", key)
		}
	}

	_, err := digest.Parse(req.Version)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	v := &models.Version{
		Image:   req.Name,
		Version: req.Version,
	}

	return &empty.Empty{}, v.Upsert(ctx, u.db, true, []string{"image"}, boil.Infer(), boil.Infer())
}
