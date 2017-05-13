package common

import (
	"os"
	"testing"

	"github.com/hashicorp/packer/common"
)

func TestDriverConfigPrepare_WithoutConfig(t *testing.T) {
	var c *DriverConfig

	c = new(DriverConfig)
	//	c.Format = "ovf"
	pc := &common.PackerConfig{PackerBuildName: "foo"}
	errs := c.Prepare(testConfigTemplate(t), pc)
	if len(errs) == 0 {
		t.Fatalf("Should fail with error")
	}
}

func TestDriverConfigPrepare_EnvVariables(t *testing.T) {
	var c *DriverConfig

	c = new(DriverConfig)
	c.RemoteHost = "foo"
	pc := &common.PackerConfig{PackerBuildName: "foo"}
	os.Setenv("GOVC_USERNAME", "foo")
	os.Setenv("GOVC_PASSWORD", "bar")
	errs := c.Prepare(testConfigTemplate(t), pc)
	if len(errs) > 0 {
		t.Fatalf("bad: %#v", errs)
	}

	if c.RemoteUser != "foo" {
		t.Fatalf("User default value incorrect: %s", c.RemoteUser)
	}

	if c.RemotePassword != "bar" {
		t.Fatalf("Password default value incorrect: %s", c.RemotePassword)
	}
}

func TestDriverConfigPrepare_Default(t *testing.T) {
	var c *DriverConfig

	c = new(DriverConfig)
	c.RemoteHost = "foo"
	pc := &common.PackerConfig{PackerBuildName: "foo"}
	os.Unsetenv("GOVC_USERNAME")
	os.Unsetenv("GOVC_PASSWORD")
	errs := c.Prepare(testConfigTemplate(t), pc)
	if len(errs) > 0 {
		t.Fatalf("bad: %#v", errs)
	}

	if c.VMName != "packer-foo" {
		t.Fatalf("VMName default value incorrect: %s", c.VMName)
	}

	if c.Vcenter != c.RemoteHost {
		t.Fatalf("Vcenter should default to Host: %s", c.Vcenter)
	}

	if c.RemoteUser != "root" {
		t.Fatalf("User default value incorrect: %s", c.RemoteUser)
	}

	if c.RemotePassword != "" {
		t.Fatalf("Password default value incorrect: %s", c.RemotePassword)
	}

	if c.RemoteDatacenter != "ha-datacenter" {
		t.Fatalf("Datacenter default value incorrect: %s", c.RemoteDatacenter)
	}

	if c.RemoteDatastore != "datastore1" {
		t.Fatalf("Datastore default value incorrect: %s", c.RemoteDatastore)
	}

	if c.RemoteCacheDatastore != c.RemoteDatastore {
		t.Fatalf("Cache Datastore should default value to Datastore: %s", c.RemoteCacheDatastore)
	}

	if c.RemoteCacheDirectory != "packer_cache" {
		t.Fatalf("Cache directory default value incorrect: %s", c.RemoteCacheDirectory)
	}
}

func TestDriverConfigPrepare_DefaultCacheDatastore(t *testing.T) {
	var c *DriverConfig

	c = new(DriverConfig)
	c.RemoteHost = "foo"
	pc := &common.PackerConfig{PackerBuildName: "foo"}
	errs := c.Prepare(testConfigTemplate(t), pc)
	if len(errs) != 0 {
		t.Fatalf("bad: %#v", errs)
	}
	if c.RemoteCacheDatastore != c.RemoteDatastore {
		t.Fatalf("Cache Datastore should default value to Datastore: %s", c.RemoteCacheDatastore)
	}

}
func TestDriverConfigPrepare_ClusterWithoutVcenter(t *testing.T) {
	var c *DriverConfig

	c = new(DriverConfig)
	c.RemoteCluster = "foo"
	pc := &common.PackerConfig{PackerBuildName: "foo"}
	errs := c.Prepare(testConfigTemplate(t), pc)
	if len(errs) == 0 {
		t.Fatalf("should fail with error")
	}

}

func TestDriverConfigPrepare_ClusterVcenter(t *testing.T) {
	var c *DriverConfig

	c = new(DriverConfig)
	c.RemoteCluster = "foo"
	c.Vcenter = "bar"
	pc := &common.PackerConfig{PackerBuildName: "foo"}
	errs := c.Prepare(testConfigTemplate(t), pc)
	if len(errs) > 0 {
		t.Fatalf("bad: %#v", errs)
	}

}
func TestDriverConfigPrepare_HostVcenter(t *testing.T) {
	var c *DriverConfig

	c = new(DriverConfig)
	c.RemoteHost = "foo"
	c.Vcenter = "bar"
	pc := &common.PackerConfig{PackerBuildName: "foo"}
	errs := c.Prepare(testConfigTemplate(t), pc)
	if len(errs) > 0 {
		t.Fatalf("bad: %#v", errs)
	}

	if c.Vcenter == c.RemoteHost {
		t.Fatalf("Vcenter should not be overriden by Host")
	}

}
