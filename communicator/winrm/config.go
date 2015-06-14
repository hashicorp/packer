package winrm

import (
	"time"
)

// Config is used to configure the WinRM connection
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	Timeout  time.Duration
}
