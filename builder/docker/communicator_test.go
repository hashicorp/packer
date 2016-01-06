package docker

import (
	"crypto/sha256"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/provisioner/file"
	"github.com/mitchellh/packer/provisioner/shell"
	"github.com/mitchellh/packer/template"
)

func TestCommunicator_impl(t *testing.T) {
	var _ packer.Communicator = new(Communicator)
}

// TestUploadDownload verifies that basic upload / download functionality works
func TestUploadDownload(t *testing.T) {
	ui := packer.TestUi(t)
	cache := &packer.FileCache{CacheDir: os.TempDir()}

	tpl, err := template.Parse(strings.NewReader(dockerBuilderConfig))
	if err != nil {
		t.Fatalf("Unable to parse config: %s", err)
	}

	if os.Getenv("PACKER_ACC") == "" {
		t.Skip("This test is only run with PACKER_ACC=1")
	}
	cmd := exec.Command("docker", "-v")
	cmd.Run()
	if !cmd.ProcessState.Success() {
		t.Error("docker command not found; please make sure docker is installed")
	}

	// Setup the builder
	builder := &Builder{}
	warnings, err := builder.Prepare(tpl.Builders["docker"].Config)
	if err != nil {
		t.Fatalf("Error preparing configuration %s", err)
	}
	if len(warnings) > 0 {
		t.Fatal("Encountered configuration warnings; aborting")
	}

	// Setup the provisioners
	upload := &file.Provisioner{}
	err = upload.Prepare(tpl.Provisioners[0].Config)
	if err != nil {
		t.Fatalf("Error preparing upload: %s", err)
	}
	download := &file.Provisioner{}
	err = download.Prepare(tpl.Provisioners[1].Config)
	if err != nil {
		t.Fatalf("Error preparing download: %s", err)
	}
	// Preemptive cleanup. Honestly I don't know why you would want to get rid
	// of my strawberry cake. It's so tasty! Do you not like cake? Are you a
	// cake-hater? Or are you keeping all the cake all for yourself? So selfish!
	defer os.Remove("my-strawberry-cake")

	// Add hooks so the provisioners run during the build
	hooks := map[string][]packer.Hook{}
	hooks[packer.HookProvision] = []packer.Hook{
		&packer.ProvisionHook{
			Provisioners: []packer.Provisioner{
				upload,
				download,
			},
		},
	}
	hook := &packer.DispatchHook{Mapping: hooks}

	// Run things
	artifact, err := builder.Run(ui, hook, cache)
	if err != nil {
		t.Fatalf("Error running build %s", err)
	}
	// Preemptive cleanup
	defer artifact.Destroy()

	// Verify that the thing we downloaded is the same thing we sent up.
	// Complain loudly if it isn't.
	inputFile, err := ioutil.ReadFile("test-fixtures/onecakes/strawberry")
	if err != nil {
		t.Fatalf("Unable to read input file: %s", err)
	}
	outputFile, err := ioutil.ReadFile("my-strawberry-cake")
	if err != nil {
		t.Fatalf("Unable to read output file: %s", err)
	}
	if sha256.Sum256(inputFile) != sha256.Sum256(outputFile) {
		t.Fatalf("Input and output files do not match\n"+
			"Input:\n%s\nOutput:\n%s\n", inputFile, outputFile)
	}
}

// TestLargeDownload verifies that files are the apporpriate size after being
// downloaded. This is to identify and fix the race condition in #2793. You may
// need to use github.com/cbednarski/rerun to verify since this problem occurs
// only intermittently.
func TestLargeDownload(t *testing.T) {
	ui := packer.TestUi(t)
	cache := &packer.FileCache{CacheDir: os.TempDir()}

	tpl, err := template.Parse(strings.NewReader(dockerLargeBuilderConfig))
	if err != nil {
		t.Fatalf("Unable to parse config: %s", err)
	}

	if os.Getenv("PACKER_ACC") == "" {
		t.Skip("This test is only run with PACKER_ACC=1")
	}
	cmd := exec.Command("docker", "-v")
	cmd.Run()
	if !cmd.ProcessState.Success() {
		t.Error("docker command not found; please make sure docker is installed")
	}

	// Setup the builder
	builder := &Builder{}
	warnings, err := builder.Prepare(tpl.Builders["docker"].Config)
	if err != nil {
		t.Fatalf("Error preparing configuration %s", err)
	}
	if len(warnings) > 0 {
		t.Fatal("Encountered configuration warnings; aborting")
	}

	// Setup the provisioners
	shell := &shell.Provisioner{}
	err = shell.Prepare(tpl.Provisioners[0].Config)
	if err != nil {
		t.Fatalf("Error preparing shell provisioner: %s", err)
	}
	downloadCupcake := &file.Provisioner{}
	err = downloadCupcake.Prepare(tpl.Provisioners[1].Config)
	if err != nil {
		t.Fatalf("Error preparing downloadCupcake: %s", err)
	}
	downloadBigcake := &file.Provisioner{}
	err = downloadBigcake.Prepare(tpl.Provisioners[2].Config)
	if err != nil {
		t.Fatalf("Error preparing downloadBigcake: %s", err)
	}

	// Preemptive cleanup.
	defer os.Remove("cupcake")
	defer os.Remove("bigcake")

	// Add hooks so the provisioners run during the build
	hooks := map[string][]packer.Hook{}
	hooks[packer.HookProvision] = []packer.Hook{
		&packer.ProvisionHook{
			Provisioners: []packer.Provisioner{
				shell,
				downloadCupcake,
				downloadBigcake,
			},
		},
	}
	hook := &packer.DispatchHook{Mapping: hooks}

	// Run things
	artifact, err := builder.Run(ui, hook, cache)
	if err != nil {
		t.Fatalf("Error running build %s", err)
	}
	// Preemptive cleanup
	defer artifact.Destroy()

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
	// 	t.Fatalf("Input and output files do not match\n"+
	// 		"Input:\n%s\nOutput:\n%s\n", inputFile, outputFile)
	// }

}

const dockerBuilderConfig = `
{
  "builders": [
    {
      "type": "docker",
      "image": "ubuntu",
      "discard": true,
      "run_command": ["-d", "-i", "-t", "{{.Image}}", "/bin/sh"]
    }
  ],
  "provisioners": [
    {
      "type": "file",
      "source": "test-fixtures/onecakes/strawberry",
      "destination": "/strawberry-cake"
    },
    {
      "type": "file",
      "source": "/strawberry-cake",
      "destination": "my-strawberry-cake",
      "direction": "download"
    }
  ]
}
`

const dockerLargeBuilderConfig = `
{
  "builders": [
    {
      "type": "docker",
      "image": "ubuntu",
      "discard": true
    }
  ],
  "provisioners": [
    {
      "type": "shell",
      "inline": [
        "dd if=/dev/urandom of=/tmp/cupcake bs=1M count=2",
        "dd if=/dev/urandom of=/tmp/bigcake bs=1M count=100",
        "sync",
        "md5sum /tmp/cupcake /tmp/bigcake"
      ]
    },
    {
      "type": "file",
      "source": "/tmp/cupcake",
      "destination": "cupcake",
      "direction": "download"
    },
    {
      "type": "file",
      "source": "/tmp/bigcake",
      "destination": "bigcake",
      "direction": "download"
    }
  ]
}
`
