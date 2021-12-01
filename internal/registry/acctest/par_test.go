package acctest

import (
	"bytes"
	"context"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/packer/command"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

const (
	channel = "acc"
)

func TestAcc_PAR_service_create_and_datasource(t *testing.T) {
	runUUID := "11118ff1-a860-44a3-a669-a4f1b4a12688"
	os.Setenv("PACKER_RUN_UUID", runUUID)

	checkEnvVars(t)
	const tmpBucket = "pkr-acctest-temp-1"

	// allow other tests to go.
	t.Parallel()

	cfg, err := NewTestConfig(t)
	if err != nil {
		t.Fatalf("NewParConfig: %v", err)
	}

	_, _ = cfg.DeleteBucket(context.Background(), tmpBucket)
	defer func() {
		_, err := cfg.DeleteBucket(context.Background(), tmpBucket) // Hopefully everything else is deleted too
		if err != nil {
			t.Fatalf("failed to delete bucket: %v", err)
		}
	}()

	// create our bucket, this should fail if the bucket already exists
	_, err = cfg.CreateBucket(context.Background(), tmpBucket, "", nil)
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
		iterationFingerprint := strconv.Itoa(i)
		iterationID := cfg.UpsertIteration(tmpBucket, iterationFingerprint)

		for cloud, builds := range builds {
			for region, imageIDs := range builds {
				cfg.UpsertBuild(
					tmpBucket,
					iterationFingerprint,
					runUUID,
					iterationID,
					cloud,
					cloud,
					region,
					imageIDs,
				)
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

func TestAcc_PAR_pkr_build(t *testing.T) {

	runUUID := "11118ff1-a860-44a3-a669-a4f1b4a12688"
	os.Setenv("PACKER_RUN_UUID", runUUID)

	checkEnvVars(t)
	const tmpBucket = "pkr-acctest-temp-2"

	// allow other tests to go.
	t.Parallel()

	cfg, err := NewTestConfig(t)
	if err != nil {
		t.Fatalf("NewParConfig: %v", err)
	}

	_, _ = cfg.DeleteBucket(context.Background(), tmpBucket)
	defer func() {
		_, err := cfg.DeleteBucket(context.Background(), tmpBucket) // Hopefully everything else is deleted too
		if err != nil {
			t.Fatalf("failed to delete bucket: %v", err)
		}
	}()
	// create our bucket, this should fail if the bucket already exists
	_, err = cfg.CreateBucket(context.Background(), tmpBucket, "", nil)
	if err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}

	// now, let's try building something

	meta := command.TestMetaFile(t)
	ui := meta.Ui.(*packersdk.BasicUi)
	c := &command.BuildCommand{
		Meta: meta,
	}

	code := c.Run([]string{"./test-fixtures/build.pkr.hcl"})
	outW := ui.Writer.(*bytes.Buffer)
	errW := ui.ErrorWriter.(*bytes.Buffer)
	if code != 0 {
		t.Fatalf(
			"Bad exit code.\n\nStdout:\n\n%s\n\nStderr:\n\n%s",
			outW.String(),
			errW.String())
	}

	// extract id from output
	re := regexp.MustCompile(`(?m)^.*packer/` + tmpBucket + `/iterations/(\S*)`)
	match := re.FindStringSubmatch(outW.String())
	if len(match) != 2 {
		t.Fatalf("Could not extract ID from output.\n\nStdout:\n\n%s\n\nStderr:\n\n%s",
			outW.String(),
			errW.String())
	}
	iterationID := match[1]

	if foundID := cfg.GetIterationByID(tmpBucket, iterationID); foundID != iterationID {
		t.Fatalf("could not find iteration back in service: expected %q, found %q", iterationID, foundID)
	}
}
