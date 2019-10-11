package client

import (
	"github.com/hashicorp/go-oracle-terraform/opc"
)

// DebugLogString logs a string if debug logs are on
func (c *Client) DebugLogString(str string) {
	if c.loglevel != opc.LogDebug {
		return
	}
	c.logger.Log(str)
}
