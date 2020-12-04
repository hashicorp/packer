package docker

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	builderT "github.com/hashicorp/packer/packer-plugin-sdk/acctest"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// RenderConfig helps create dynamic packer template configs for parsing by
// builderT without having to write the config to a file.
func RenderConfig(builderConfig map[string]interface{}, provisionerConfig []map[string]interface{}) string {
	// set up basic build template
	t := map[string][]map[string]interface{}{
		"builders": {
			// Setup basic docker config
			map[string]interface{}{
				"type":    "test",
				"image":   "ubuntu",
				"discard": true,
			},
		},
		"provisioners": []map[string]interface{}{},
	}
	// apply special builder overrides
	for k, v := range builderConfig {
		t["builders"][0][k] = v
	}
	// Apply special provisioner overrides
	t["provisioners"] = append(t["provisioners"], provisionerConfig...)

	j, _ := json.Marshal(t)
	return string(j)
}

// TestUploadDownload verifies that basic upload / download functionality works
func TestUploadDownload(t *testing.T) {
	if os.Getenv("PACKER_ACC") == "" {
		t.Skip("This test is only run with PACKER_ACC=1")
	}

	dockerBuilderExtraConfig := map[string]interface{}{
		"run_command": []string{"-d", "-i", "-t", "{{.Image}}", "/bin/sh"},
	}

	dockerProvisionerConfig := []map[string]interface{}{
		{
			"type":        "file",
			"source":      "test-fixtures/onecakes/strawberry",
			"destination": "/strawberry-cake",
		},
		{
			"type":        "file",
			"source":      "/strawberry-cake",
			"destination": "my-strawberry-cake",
			"direction":   "download",
		},
	}

	configString := RenderConfig(dockerBuilderExtraConfig, dockerProvisionerConfig)

	// this should be a precheck
	cmd := exec.Command("docker", "-v")
	err := cmd.Run()
	if err != nil {
		t.Error("docker command not found; please make sure docker is installed")
	}

	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: configString,
		Check: func(a []packersdk.Artifact) error {
			// Verify that the thing we downloaded is the same thing we sent up.
			// Complain loudly if it isn't.
			inputFile, err := ioutil.ReadFile("test-fixtures/onecakes/strawberry")
			if err != nil {
				return fmt.Errorf("Unable to read input file: %s", err)
			}
			outputFile, err := ioutil.ReadFile("my-strawberry-cake")
			if err != nil {
				return fmt.Errorf("Unable to read output file: %s", err)
			}
			if sha256.Sum256(inputFile) != sha256.Sum256(outputFile) {
				return fmt.Errorf("Input and output files do not match\n"+
					"Input:\n%s\nOutput:\n%s\n", inputFile, outputFile)
			}
			return nil
		},
		Teardown: func() error {
			// Cleanup. Honestly I don't know why you would want to get rid
			// of my strawberry cake. It's so tasty! Do you not like cake? Are you a
			// cake-hater? Or are you keeping all the cake all for yourself? So selfish!
			os.Remove("my-strawberry-cake")
			return nil
		},
	})
}

// TestLargeDownload verifies that files are the appropriate size after being
// downloaded. This is to identify and fix the race condition in #2793. You may
// need to use github.com/cbednarski/rerun to verify since this problem occurs
// only intermittently.
func TestLargeDownload(t *testing.T) {
	if os.Getenv("PACKER_ACC") == "" {
		t.Skip("This test is only run with PACKER_ACC=1")
	}

	dockerProvisionerConfig := []map[string]interface{}{
		{
			"type": "shell",
			"inline": []string{
				"dd if=/dev/urandom of=/tmp/cupcake bs=1M count=2",
				"dd if=/dev/urandom of=/tmp/bigcake bs=1M count=100",
				"sync",
				"md5sum /tmp/cupcake /tmp/bigcake",
			},
		},
		{
			"type":        "file",
			"source":      "/tmp/cupcake",
			"destination": "cupcake",
			"direction":   "download",
		},
		{
			"type":        "file",
			"source":      "/tmp/bigcake",
			"destination": "bigcake",
			"direction":   "download",
		},
	}

	configString := RenderConfig(map[string]interface{}{}, dockerProvisionerConfig)

	// this should be a precheck
	cmd := exec.Command("docker", "-v")
	err := cmd.Run()
	if err != nil {
		t.Error("docker command not found; please make sure docker is installed")
	}

	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: configString,
		Check: func(a []packersdk.Artifact) error {
			// Verify that the things we downloaded are the right size. Complain loudly
			// if they are not.
			//
			// cupcake should be 2097152 bytes
			// bigcake should be 104857600 bytes
			cupcake, err := os.Stat("cupcake")
			if err != nil {
				t.Fatalf("Unable to stat cupcake file: %s", err)
			}
			cupcakeExpected := int64(2097152)
			if cupcake.Size() != cupcakeExpected {
				t.Errorf("Expected cupcake to be %d bytes; found %d", cupcakeExpected, cupcake.Size())
			}

			bigcake, err := os.Stat("bigcake")
			if err != nil {
				t.Fatalf("Unable to stat bigcake file: %s", err)
			}
			bigcakeExpected := int64(104857600)
			if bigcake.Size() != bigcakeExpected {
				t.Errorf("Expected bigcake to be %d bytes; found %d", bigcakeExpected, bigcake.Size())
			}

			// TODO if we can, calculate a sha inside the container and compare to the
			// one we get after we pull it down. We will probably have to parse the log
			// or ui output to do this because we use /dev/urandom to create the file.

			// if sha256.Sum256(inputFile) != sha256.Sum256(outputFile) {
			//	t.Fatalf("Input and output files do not match\n"+
			//		"Input:\n%s\nOutput:\n%s\n", inputFile, outputFile)
			// }
			return nil
		},
		Teardown: func() error {
			os.Remove("cupcake")
			os.Remove("bigcake")
			return nil
		},
	})

}

// TestFixUploadOwner verifies that owner of uploaded files is the user the container is running as.
func TestFixUploadOwner(t *testing.T) {
	if os.Getenv("PACKER_ACC") == "" {
		t.Skip("This test is only run with PACKER_ACC=1")
	}

	cmd := exec.Command("docker", "-v")
	err := cmd.Run()
	if err != nil {
		t.Error("docker command not found; please make sure docker is installed")
	}

	dockerBuilderExtraConfig := map[string]interface{}{
		"run_command": []string{"-d", "-i", "-t", "-u", "42", "{{.Image}}", "/bin/sh"},
	}

	testFixUploadOwnerProvisionersTemplate := []map[string]interface{}{
		{
			"type":        "file",
			"source":      "test-fixtures/onecakes/strawberry",
			"destination": "/tmp/strawberry-cake",
		},
		{
			"type":        "file",
			"source":      "test-fixtures/manycakes",
			"destination": "/tmp/",
		},
		{
			"type":   "shell",
			"inline": "touch /tmp/testUploadOwner",
		},
		{
			"type": "shell",
			"inline": []string{
				"[ $(stat -c %u /tmp/strawberry-cake) -eq 42 ] || (echo 'Invalid owner of /tmp/strawberry-cake' && exit 1)",
				"[ $(stat -c %u /tmp/testUploadOwner) -eq 42 ] || (echo 'Invalid owner of /tmp/testUploadOwner' && exit 1)",
				"find /tmp/manycakes | xargs -n1 -IFILE /bin/sh -c '[ $(stat -c %u FILE) -eq 42 ] || (echo \"Invalid owner of FILE\" && exit 1)'",
			},
		},
	}

	configString := RenderConfig(dockerBuilderExtraConfig, testFixUploadOwnerProvisionersTemplate)
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: configString,
	})
}
