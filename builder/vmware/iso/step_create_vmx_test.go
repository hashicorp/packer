package iso

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"testing"

	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template"
	"github.com/hashicorp/packer/provisioner/shell"
)

var vmxTestBuilderConfig = map[string]string{
	"type":                        `"vmware-iso"`,
	"iso_url":                     `"https://archive.org/download/ut-ttylinux-i686-12.6/ut-ttylinux-i686-12.6.iso"`,
	"iso_checksum":                `"md5:43c1feeae55a44c6ef694b8eb18408a6"`,
	"ssh_username":                `"root"`,
	"ssh_password":                `"password"`,
	"ssh_wait_timeout":            `"45s"`,
	"boot_command":                `["<enter><wait5><wait10>","root<enter><wait>password<enter><wait>","udhcpc<enter><wait>"]`,
	"shutdown_command":            `"/sbin/shutdown -h; exit 0"`,
	"ssh_key_exchange_algorithms": `["diffie-hellman-group1-sha1"]`,
}

var vmxTestProvisionerConfig = map[string]string{
	"type":   `"shell"`,
	"inline": `["echo hola mundo"]`,
}

const vmxTestTemplate string = `{"builders":[{%s}],"provisioners":[{%s}]}`

func tmpnam(prefix string) string {
	var path string
	var err error

	const length = 16

	dir := os.TempDir()
	max := int(math.Pow(2, float64(length)))

	// FIXME use ioutil.TempFile() or at least mimic implementation, this could loop forever
	n, err := rand.Intn(max), nil
	for path = filepath.Join(dir, prefix+strconv.Itoa(n)); err == nil; _, err = os.Stat(path) {
		n = rand.Intn(max)
		path = filepath.Join(dir, prefix+strconv.Itoa(n))
	}
	return path
}

func createFloppyOutput(prefix string) (string, string, error) {
	output := tmpnam(prefix)
	f, err := os.Create(output)
	if err != nil {
		return "", "", fmt.Errorf("Unable to create empty %s: %s", output, err)
	}
	f.Close()

	vmxData := []string{
		`"floppy0.present":"TRUE"`,
		`"floppy0.fileType":"file"`,
		`"floppy0.clientDevice":"FALSE"`,
		`"floppy0.fileName":"%s"`,
		`"floppy0.startConnected":"TRUE"`,
	}

	outputFile := strings.Replace(output, "\\", "\\\\", -1)
	vmxString := fmt.Sprintf("{"+strings.Join(vmxData, ",")+"}", outputFile)
	return output, vmxString, nil
}

func readFloppyOutput(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("Unable to open file %s", path)
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("Unable to read file: %s", err)
	}
	if len(data) == 0 {
		return "", nil
	}
	return string(data[:bytes.IndexByte(data, 0)]), nil
}

func setupVMwareBuild(t *testing.T, builderConfig map[string]string, provisionerConfig map[string]string) error {
	ui := packer.TestUi(t)

	// create builder config and update with user-supplied options
	cfgBuilder := map[string]string{}
	for k, v := range vmxTestBuilderConfig {
		cfgBuilder[k] = v
	}
	for k, v := range builderConfig {
		cfgBuilder[k] = v
	}

	// convert our builder config into a single sprintfable string
	builderLines := []string{}
	for k, v := range cfgBuilder {
		builderLines = append(builderLines, fmt.Sprintf(`"%s":%s`, k, v))
	}

	// create provisioner config and update with user-supplied options
	cfgProvisioner := map[string]string{}
	for k, v := range vmxTestProvisionerConfig {
		cfgProvisioner[k] = v
	}
	for k, v := range provisionerConfig {
		cfgProvisioner[k] = v
	}

	// convert our provisioner config into a single sprintfable string
	provisionerLines := []string{}
	for k, v := range cfgProvisioner {
		provisionerLines = append(provisionerLines, fmt.Sprintf(`"%s":%s`, k, v))
	}

	// and now parse them into a template
	configString := fmt.Sprintf(vmxTestTemplate, strings.Join(builderLines, `,`), strings.Join(provisionerLines, `,`))

	tpl, err := template.Parse(strings.NewReader(configString))
	if err != nil {
		t.Fatalf("Unable to parse test config: %s", err)
	}

	// create our config to test the vmware-iso builder
	components := packer.ComponentFinder{
		BuilderStore: packer.MapOfBuilder{
			"vmware-iso": func() (packer.Builder, error) { return &Builder{}, nil },
		},
		Hook: func(n string) (packersdk.Hook, error) {
			return &packer.DispatchHook{}, nil
		},
		ProvisionerStore: packer.MapOfProvisioner{
			"shell": func() (packer.Provisioner, error) { return &shell.Provisioner{}, nil },
		},
		PostProcessorStore: packer.MapOfPostProcessor{
			"something": func() (packer.PostProcessor, error) { return &packer.MockPostProcessor{}, nil },
		},
	}
	config := packer.CoreConfig{
		Template:   tpl,
		Components: components,
	}

	// create a core using our template
	core := packer.NewCore(&config)
	err = core.Initialize()
	if err != nil {
		t.Fatalf("Unable to create core: %s", err)
	}

	// now we can prepare our build
	b, err := core.Build("vmware-iso")
	if err != nil {
		t.Fatalf("Unable to create build: %s", err)
	}

	warn, err := b.Prepare()
	if err != nil {
		t.Fatalf("error preparing build: %v", err)
	}
	if len(warn) > 0 {
		for _, w := range warn {
			t.Logf("Configuration warning: %s", w)
		}
	}

	// and then finally build it
	artifacts, err := b.Run(context.Background(), ui)
	if err != nil {
		t.Fatalf("Failed to build artifact: %s", err)
	}

	// check to see that we only got one artifact back
	if len(artifacts) == 1 {
		return artifacts[0].Destroy()
	}

	// otherwise some number of errors happened
	t.Logf("Unexpected number of artifacts returned: %d", len(artifacts))
	errors := make([]error, 0)
	for _, artifact := range artifacts {
		if err := artifact.Destroy(); err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		t.Errorf("%d Errors returned while trying to destroy artifacts", len(errors))
		return fmt.Errorf("Error while trying to destroy artifacts: %v", errors)
	}
	return nil
}

func TestStepCreateVmx_SerialFile(t *testing.T) {
	if os.Getenv("PACKER_ACC") == "" {
		t.Skip("This test is only run with PACKER_ACC=1 due to the requirement of access to the VMware binaries.")
	}

	tmpfile := tmpnam("SerialFileInput.")

	serialConfig := map[string]string{
		"serial": fmt.Sprintf(`"file:%s"`, filepath.ToSlash(tmpfile)),
	}

	error := setupVMwareBuild(t, serialConfig, map[string]string{})
	if error != nil {
		t.Errorf("Unable to read file: %s", error)
	}

	f, err := os.Stat(tmpfile)
	if err != nil {
		t.Errorf("VMware builder did not create a file for serial port: %s", err)
	}

	if f != nil {
		if err := os.Remove(tmpfile); err != nil {
			t.Fatalf("Unable to remove file %s: %s", tmpfile, err)
		}
	}
}

func TestStepCreateVmx_SerialPort(t *testing.T) {
	if os.Getenv("PACKER_ACC") == "" {
		t.Skip("This test is only run with PACKER_ACC=1 due to the requirement of access to the VMware binaries.")
	}

	var defaultSerial string
	if runtime.GOOS == "windows" {
		defaultSerial = "COM1"
	} else {
		defaultSerial = "/dev/ttyS0"
	}

	config := map[string]string{
		"serial": fmt.Sprintf(`"device:%s"`, filepath.ToSlash(defaultSerial)),
	}
	provision := map[string]string{
		"inline": `"dmesg | egrep -o '^serial8250: ttyS1 at' > /dev/fd0"`,
	}

	// where to write output
	output, vmxData, err := createFloppyOutput("SerialPortOutput.")
	if err != nil {
		t.Fatalf("Error creating output: %s", err)
	}
	defer func() {
		if _, err := os.Stat(output); err == nil {
			os.Remove(output)
		}
	}()
	config["vmx_data"] = vmxData
	t.Logf("Preparing to write output to %s", output)

	// whee
	err = setupVMwareBuild(t, config, provision)
	if err != nil {
		t.Errorf("%s", err)
	}

	// check the output
	data, err := readFloppyOutput(output)
	if err != nil {
		t.Errorf("%s", err)
	}

	if data != "serial8250: ttyS1 at\n" {
		t.Errorf("Serial port not detected : %v", data)
	}
}

func TestStepCreateVmx_ParallelPort(t *testing.T) {
	if os.Getenv("PACKER_ACC") == "" {
		t.Skip("This test is only run with PACKER_ACC=1 due to the requirement of access to the VMware binaries.")
	}

	var defaultParallel string
	if runtime.GOOS == "windows" {
		defaultParallel = "LPT1"
	} else {
		defaultParallel = "/dev/lp0"
	}

	config := map[string]string{
		"parallel": fmt.Sprintf(`"device:%s,uni"`, filepath.ToSlash(defaultParallel)),
	}
	provision := map[string]string{
		"inline": `"cat /proc/modules | egrep -o '^parport ' > /dev/fd0"`,
	}

	// where to write output
	output, vmxData, err := createFloppyOutput("ParallelPortOutput.")
	if err != nil {
		t.Fatalf("Error creating output: %s", err)
	}
	defer func() {
		if _, err := os.Stat(output); err == nil {
			os.Remove(output)
		}
	}()
	config["vmx_data"] = vmxData
	t.Logf("Preparing to write output to %s", output)

	// whee
	error := setupVMwareBuild(t, config, provision)
	if error != nil {
		t.Errorf("%s", error)
	}

	// check the output
	data, err := readFloppyOutput(output)
	if err != nil {
		t.Errorf("%s", err)
	}

	if data != "parport \n" {
		t.Errorf("Parallel port not detected : %v", data)
	}
}

func TestStepCreateVmx_Usb(t *testing.T) {
	if os.Getenv("PACKER_ACC") == "" {
		t.Skip("This test is only run with PACKER_ACC=1 due to the requirement of access to the VMware binaries.")
	}

	config := map[string]string{
		"usb": `"TRUE"`,
	}
	provision := map[string]string{
		"inline": `"dmesg | egrep -m1 -o 'USB hub found$' > /dev/fd0"`,
	}

	// where to write output
	output, vmxData, err := createFloppyOutput("UsbOutput.")
	if err != nil {
		t.Fatalf("Error creating output: %s", err)
	}
	defer func() {
		if _, err := os.Stat(output); err == nil {
			os.Remove(output)
		}
	}()
	config["vmx_data"] = vmxData
	t.Logf("Preparing to write output to %s", output)

	// whee
	error := setupVMwareBuild(t, config, provision)
	if error != nil {
		t.Errorf("%s", error)
	}

	// check the output
	data, err := readFloppyOutput(output)
	if err != nil {
		t.Errorf("%s", err)
	}

	if data != "USB hub found\n" {
		t.Errorf("USB support not detected : %v", data)
	}
}

func TestStepCreateVmx_Sound(t *testing.T) {
	if os.Getenv("PACKER_ACC") == "" {
		t.Skip("This test is only run with PACKER_ACC=1 due to the requirement of access to the VMware binaries.")
	}

	config := map[string]string{
		"sound": `"TRUE"`,
	}
	provision := map[string]string{
		"inline": `"cat /proc/modules | egrep -o '^soundcore' > /dev/fd0"`,
	}

	// where to write output
	output, vmxData, err := createFloppyOutput("SoundOutput.")
	if err != nil {
		t.Fatalf("Error creating output: %s", err)
	}
	defer func() {
		if _, err := os.Stat(output); err == nil {
			os.Remove(output)
		}
	}()
	config["vmx_data"] = vmxData
	t.Logf("Preparing to write output to %s", output)

	// whee
	error := setupVMwareBuild(t, config, provision)
	if error != nil {
		t.Errorf("Unable to read file: %s", error)
	}

	// check the output
	data, err := readFloppyOutput(output)
	if err != nil {
		t.Errorf("%s", err)
	}

	if data != "soundcore\n" {
		t.Errorf("Soundcard not detected : %v", data)
	}
}
