package plugin

import (
	"os/exec"
)

type Client struct {
	cmd *exec.Cmd
}

func NewClient(cmd *exec.Cmd) *Client {
	return &Client{cmd}
}
