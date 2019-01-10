package plugin

import (
	"log"

	"github.com/hashicorp/packer/packer"
)

type cmdPreProcessor struct {
	p      packer.PreProcessor
	client *Client
}

func (c *cmdPreProcessor) Configure(config ...interface{}) error {
	defer func() {
		r := recover()
		c.checkExit(r, nil)
	}()

	return c.p.Configure(config...)
}

func (c *cmdPreProcessor) PreProcess(ui packer.Ui) error {
	defer func() {
		r := recover()
		c.checkExit(r, nil)
	}()

	return c.p.PreProcess(ui)
}

func (c *cmdPreProcessor) checkExit(p interface{}, cb func()) {
	if c.client.Exited() && cb != nil {
		cb()
	} else if p != nil && !Killed {
		log.Panic(p)
	}
}
