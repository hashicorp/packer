package vsphere

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
)

func TestArgs(t *testing.T) {
	var p PostProcessor

	p.config.Username = "me"
	p.config.Password = "notpassword"
	p.config.Host = "myhost"
	p.config.Datacenter = "mydc"
	p.config.Cluster = "mycluster"
	p.config.VMName = "my vm"
	p.config.Datastore = "my datastore"
	p.config.Insecure = true
	p.config.DiskMode = "thin"
	p.config.VMFolder = "my folder"

	source := "something.vmx"
	ovftool_uri := fmt.Sprintf("vi://%s:%s@%s/%s/host/%s",
		url.QueryEscape(p.config.Username),
		url.QueryEscape(p.config.Password),
		p.config.Host,
		p.config.Datacenter,
		p.config.Cluster)

	if p.config.ResourcePool != "" {
		ovftool_uri += "/Resources/" + p.config.ResourcePool
	}

	args, err := p.BuildArgs(source, ovftool_uri)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	t.Logf("ovftool %s", strings.Join(args, " "))
}
