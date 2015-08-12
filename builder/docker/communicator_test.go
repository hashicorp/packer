package docker

import (
	"crypto/sha256"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/provisioner/file"
	"github.com/mitchellh/packer/template"
)

func TestCommunicator_impl(t *testing.T) {
	var _ packer.Communicator = new(Communicator)
}

func TestUploadDownload(t *testing.T) {
	ui := packer.TestUi(t)
	cache := &packer.FileCache{CacheDir: os.TempDir()}

	tpl, err := template.Parse(strings.NewReader(dockerBuilderConfig))
	if err != nil {
		t.Fatalf("Unable to parse config: %s", err)
	}

	// Make sure we only run this on linux hosts
	if os.Getenv("PACKER_ACC") == "" {
		t.Skip("This test is only run with PACKER_ACC=1")
	}
	if runtime.GOOS != "linux" {
		t.Skip("This test is only supported on linux")
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

const dockerBuilderConfig = `
{
  "builders": [
    {
      "type": "docker",
      "image": "alpine",
      "export_path": "alpine.tar",
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
