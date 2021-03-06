package actions

import (
	"math/rand"
	"strings"
	"time"

	"github.com/cachecashproject/watchtower/client"
	"github.com/cachecashproject/watchtower/container"
	"github.com/cachecashproject/watchtower/status"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

// UpdateParams contains all different options available to alter the behavior of the Update func
type UpdateParams struct {
	Filter                  container.Filter
	Cleanup                 bool
	NoRestart               bool
	Timeout                 time.Duration
	MonitorOnly             bool
	StatusEndpoint          string
	UpdateServer            string
	EnableInsecureTransport bool
}

// Update looks at the running Docker containers to see if any of the images
// used to start those containers have been updated. If a change is detected in
// any of the images, the associated containers are stopped and restarted with
// the new image.
func Update(cl container.Client, params UpdateParams) error {
	log.Debug("Checking containers for updated images")

	containers, err := cl.ListContainers(params.Filter)
	if err != nil {
		return err
	}

	updateClient, err := client.NewUpdateClient(log.New(), params.UpdateServer, params.EnableInsecureTransport)
	if err != nil {
		return err
	}
	defer updateClient.Close()

	identity, err := status.FetchPublicKeyIdentity(params.StatusEndpoint)
	if err != nil {
		return errors.Wrap(err, "Failed to get identity from status page")
	}
	log.Infof("Identifiying to update server as: %v", identity)

	updates, err := updateClient.CheckForUpdates(containers, identity)
	if err != nil {
		return err
	}

	for i, container := range containers {
		update, ok := updates[container.ImageName()]
		if ok {
			log.Infof("Update service flagged container outdated: %s. Pulling update for %s.", container.Name(), container.ImageName())

			s := strings.Split(container.ImageName(), ":")
			ref, tag := s[0], s[1]

			err = cl.PullImageBySha(ref, update, tag)
			if err != nil {
				log.Errorf("Unable to pull container, skipping.")
				containers[i].Stale = false
			} else {
				containers[i].Stale = true
			}
		}
	}

	containers, err = container.SortByDependencies(containers)
	if err != nil {
		return err
	}

	checkDependencies(containers)

	if params.MonitorOnly {
		return nil
	}

	// Stop stale containers in reverse order
	for i := len(containers) - 1; i >= 0; i-- {
		container := containers[i]

		if container.IsWatchtower() {
			log.Debugf("This is the watchtower container %s", containers[i].Name())
			continue
		}

		if container.Stale {
			if err := cl.StopContainer(container, params.Timeout); err != nil {
				log.Error(err)
			}
		}
	}

	// Restart stale containers in sorted order
	for _, container := range containers {
		if container.Stale {
			// Since we can't shutdown a watchtower container immediately, we need to
			// start the new one while the old one is still running. This prevents us
			// from re-using the same container name so we first rename the current
			// instance so that the new one can adopt the old name.
			if container.IsWatchtower() {
				if err := cl.RenameContainer(container, randName()); err != nil {
					log.Error(err)
					continue
				}
			}

			if !params.NoRestart {
				if err := cl.StartContainer(container); err != nil {
					log.Error(err)
				}
			}

			if params.Cleanup {
				cl.RemoveImage(container)
			}
		}
	}

	return nil
}

func checkDependencies(containers []container.Container) {

	for i, parent := range containers {
		if parent.Stale {
			continue
		}

	LinkLoop:
		for _, linkName := range parent.Links() {
			for _, child := range containers {
				if child.Name() == linkName && child.Stale {
					containers[i].Stale = true
					break LinkLoop
				}
			}
		}
	}
}

// Generates a random, 32-character, Docker-compatible container name.
func randName() string {
	b := make([]rune, 32)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}
