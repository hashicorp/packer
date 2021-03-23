package proxmox

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/stretchr/testify/require"
)

func TestTokenAuth(t *testing.T) {
	mockAPI := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorization") != "PVEAPIToken=dummy@vmhost!test-token=ac5293bf-15e2-477f-b04c-a6dfa7a46b80" {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}
	}))
	defer mockAPI.Close()

	pmURL, _ := url.Parse(mockAPI.URL)
	config := Config{
		proxmoxURL:         pmURL,
		SkipCertValidation: false,
		Username:           "dummy@vmhost!test-token",
		Password:           "not-used",
		Token:              "ac5293bf-15e2-477f-b04c-a6dfa7a46b80",
	}

	client, err := newProxmoxClient(config)
	require.NoError(t, err)

	ref := proxmox.NewVmRef(110)
	ref.SetNode("node1")
	ref.SetVmType("qemu")
	err = client.Sendkey(ref, "ping")
	require.NoError(t, err)
}

func TestLogin(t *testing.T) {
	mockAPI := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// mock ticketing api
		if req.Method == http.MethodPost && req.URL.Path == "/access/ticket" {
			body, _ := ioutil.ReadAll(req.Body)
			values, _ := url.ParseQuery(string(body))
			user := values.Get("username")
			pass := values.Get("password")
			if user != "dummy@vmhost" || pass != "correct-horse-battery-staple" {
				rw.WriteHeader(http.StatusUnauthorized)
				return
			}

			_ = json.NewEncoder(rw).Encode(map[string]interface{}{
				"data": map[string]string{
					"username":            user,
					"ticket":              "dummy-ticket",
					"CSRFPreventionToken": "random-token",
				},
			})
			return
		}

		// validate ticket
		if val, err := req.Cookie("PVEAuthCookie"); err != nil || val.Value != "dummy-ticket" {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}
	}))
	defer mockAPI.Close()

	pmURL, _ := url.Parse(mockAPI.URL)
	config := Config{
		proxmoxURL:         pmURL,
		SkipCertValidation: false,
		Username:           "dummy@vmhost",
		Password:           "correct-horse-battery-staple",
		Token:              "",
	}

	client, err := newProxmoxClient(config)
	require.NoError(t, err)

	ref := proxmox.NewVmRef(110)
	ref.SetNode("node1")
	ref.SetVmType("qemu")
	err = client.Sendkey(ref, "ping")
	require.NoError(t, err)
}
