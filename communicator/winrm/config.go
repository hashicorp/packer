package winrm

import (
	"net/http"
	"time"
)

// Config is used to configure the WinRM connection
type Config struct {
	Host               string
	Port               int
	Username           string
	Password           string
	Timeout            time.Duration
	Https              bool
	Insecure           bool
	TransportDecorator func(*http.Transport) http.RoundTripper
}
