package proxmox

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Telmate/proxmox-api-go/proxmox"
)

const defaultTaskTimeout = 30 * time.Second

type authenticatedTransport struct {
	rt    http.RoundTripper
	user  string
	token string
}

func newAuthenticatedTransport(rt http.RoundTripper, user, token string) *authenticatedTransport {
	return &authenticatedTransport{rt, user, token}
}

func (t *authenticatedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	auth := fmt.Sprintf("PVEAPIToken=%s=%s", t.user, t.token)
	req.Header.Set("Authorization", auth)
	return t.rt.RoundTrip(req)
}

func NewProxmoxClient(config Config) (*proxmox.Client, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.SkipCertValidation,
	}

	client, err := proxmox.NewClient(config.proxmoxURL.String(), nil, tlsConfig, int(defaultTaskTimeout.Seconds()))
	if err != nil {
		return nil, err
	}

	if config.Token != "" {
		// configure token auth
		log.Print("using token auth")
		client.SetAPIToken(config.Username, config.Token)
	} else {
		// fallback to login if not using tokens
		log.Print("using password auth")
		err = client.Login(config.Username, config.Password, "")
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}
