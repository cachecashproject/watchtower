package docker

//
// This code is adapted from the fetcher and pull modules in
// https://github.com/box-builder/box, which is apache licensed.
//

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/term"
)

// FetchImage pulls an image and reports the image ID pulled.
func FetchImage(context context.Context, client *client.Client, name string, pull bool) (string, error) {
	if !strings.Contains(name, ":") {
		// if we don't have a sub-tag, we need to add :latest to avoid pulling the whole repo.
		name += ":latest"
	}

	if pull {
		reader, err := client.ImagePull(context, name, types.ImagePullOptions{})
		if err != nil {
			return "", err
		}

		// this will not print anything if the tty is not enabled.
		_, err = NewProgress(term.IsTerminal(os.Stdin.Fd()), reader).Process()
		if err != nil {
			return "", err
		}

		select {
		case <-context.Done():
			if context.Err() != nil {
				return "", context.Err()
			}
		default:
		}
	}

	inspect, _, err := client.ImageInspectWithRaw(context, name)
	if err != nil {
		return "", err
	}

	select {
	case <-context.Done():
		if context.Err() != nil {
			return "", context.Err()
		}
	default:
	}

	if len(inspect.RepoDigests) == 0 {
		return "", errors.New("no digests, wtf")
	}

	var digestC int

	for i, tag := range inspect.RepoTags {
		if tag == name {
			digestC = i
			break
		}
	}

	return strings.SplitN(inspect.RepoDigests[digestC], "@", 2)[1], nil
}
