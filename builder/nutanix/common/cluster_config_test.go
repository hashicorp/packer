package common

import (
	"os"
	"testing"

	"github.com/hashicorp/packer/template/interpolate"
)

func TestCommon_Prepare_Cluster_Config(t *testing.T) {
	t.Log("Clearing environment variables")
	os.Clearenv()
	ctx := interpolate.Context{}
	c := &NutanixCluster{}

	_, errs := c.Prepare(&ctx)
	// Exactly 4 required fields - username, password, endpoint, port
	if len(errs) != 4 {
		//t.Fatalf("bad cluster config - %d errors", len(errs))
		t.Fatalf("bad cluster config - errors: %v", errs)
	}

	u := "test_username"
	c.ClusterUsername = &u
	p := "p@ssword!"
	c.ClusterPassword = &p
	e := "https://127.0.0.1"
	c.ClusterEndpoint = &e
	r := 9200
	c.ClusterPort = &r
	i := true
	c.ClusterInsecure = &i

	warns, errs := c.Prepare(&ctx)
	// Exactly 4 required fields - username, password, endpoint, port
	if len(errs) > 0 || len(warns) > 0 {
		//t.Fatalf("bad cluster config - %d errors", len(errs))
		t.Fatalf("errors or warnings returned from valid config - errors: %v, warnings: %v", errs, warns)
	}
}
