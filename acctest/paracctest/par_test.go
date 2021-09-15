package paracctest

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/packer/acctest"
	"github.com/hashicorp/packer/command"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

const (
	tmpBucket = "pkr-acctest-temp"
	bucket    = "pkr-acctest-do-not-delete"
	channel   = "acc"
)

func TestAcc_PAR_service_create_and_datasource(t *testing.T) {

	// allow other tests to go.
	t.Parallel()

	cfg, err := NewParConfig(t)
	if err != nil {
		t.Fatalf("NewParConfig: %v", err)
	}

	defer cfg.DeleteBucket(tmpBucket) // Hopefully everything else is deleted too
	// create our bucket, this should fail if the bucket already exists
	_, err = cfg.CreateBucket(tmpBucket)
	if err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}

	// insert tons of iteration
	iterations := []map[string]map[string][]string{
		{
			"aws":        {"us-west-1": {"ami:something_1", "ami:other_thing"}},
			"gcp":        {"us-west-1": {"gcp_something_1", "gcpp"}},
			"azure":      {"us-west-1": {"azure_something_1", "azaz"}},
			"vmware-iso": {"": {"iso_something_1", "vmvmvm"}},
		},
		{
			"aws":        {"us-west-1": {"ami:something_2", "ami:other_thing"}},
			"gcp":        {"us-west-1": {"gcp_something_2", "gcpp"}},
			"azure":      {"us-west-1": {"azure_something_2", "azaz"}},
			"vmware-iso": {"": {"iso_something_2", "vmvmvm"}},
		},
		{
			"aws":        {"us-west-1": {"ami:something_3"}},
			"gcp":        {"us-west-1": {"gcp_something_3", "gcpp"}},
			"azure":      {"us-west-1": {"azure_something_3", "azaz"}},
			"vmware-iso": {"": {"iso_something_3", "vmvmvm"}},
		},
	}

	lastIterationID := ""
	for i, builds := range iterations {
		iterationsFingerprint := strconv.Itoa(i)
		iterationID := cfg.UpsertIteration(tmpBucket, iterationsFingerprint)

		for cloud, builds := range builds {
			for region, imageID := range builds {
				cfg.UpsertBuild(tmpBucket, iterationsFingerprint, iterationID, cloud, region, imageID)
			}
		}
		time.Sleep(time.Millisecond)
		lastIterationID = iterationID
	}

	if cfg.GetIterationByID(tmpBucket, lastIterationID) != lastIterationID {
		t.Fatal("that should not be possible")
	}

	cfg.UpsertChannel(tmpBucket, channel, lastIterationID)

	// now, let's try getting the datasource

	meta := command.TestMetaFile(t)
	ui := meta.Ui.(*packersdk.BasicUi)
	c := &command.BuildCommand{
		Meta: meta,
	}

	code := c.Run([]string{"./test-fixtures/ds.pkr.hcl"})
	outW := ui.Writer.(*bytes.Buffer)
	errW := ui.ErrorWriter.(*bytes.Buffer)
	if code != 0 {
		t.Fatalf(
			"Bad exit code.\n\nStdout:\n\n%s\n\nStderr:\n\n%s",
			outW.String(),
			errW.String())
	}

	if !strings.Contains(outW.String(), "the artifact id is: ami:something_3, yup yup") {
		t.Fatal("data source not found ?")
	}
}
