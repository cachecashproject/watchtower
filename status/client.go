package status

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type statusPage struct {
	PublicKey string
}

// FetchPublicKeyIdentity fetches a public key from an endpoint if configured.
// If no endpoint is configured default to "anonymous".
func FetchPublicKeyIdentity(endpoint string) (string, error) {
	if endpoint == "" {
		return "anonymous", nil
	}

	log.Infof("Fetching indentity from %v", endpoint)
	resp, err := http.Get(endpoint)
	if err != nil {
		return "", errors.Wrap(err, "failed to fetch status page")
	}

	var status statusPage
	err = json.NewDecoder(resp.Body).Decode(&status)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode json response")
	}

	return status.PublicKey, nil
}
