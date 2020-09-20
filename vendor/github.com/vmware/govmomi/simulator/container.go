/*
Copyright (c) 2018 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package simulator

import (
	"archive/tar"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/vmware/govmomi/vim25/types"
)

var (
	shell = "/bin/sh"
)

func init() {
	if sh, err := exec.LookPath("bash"); err != nil {
		shell = sh
	}
}

// container provides methods to manage a container within a simulator VM lifecycle.
type container struct {
	id   string
	name string
}

type networkSettings struct {
	Gateway     string
	IPAddress   string
	IPPrefixLen int
	MacAddress  string
}

// inspect applies container network settings to vm.Guest properties.
func (c *container) inspect(vm *VirtualMachine) error {
	if c.id == "" {
		return nil
	}

	var objects []struct {
		NetworkSettings struct {
			networkSettings
			Networks map[string]networkSettings
		}
	}

	cmd := exec.Command("docker", "inspect", c.id)
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	if err = json.NewDecoder(bytes.NewReader(out)).Decode(&objects); err != nil {
		return err
	}

	vm.Config.Annotation = strings.Join(cmd.Args, " ")
	vm.logPrintf("%s: %s", vm.Config.Annotation, string(out))

	for _, o := range objects {
		s := o.NetworkSettings.networkSettings

		for _, n := range o.NetworkSettings.Networks {
			s = n
			break
		}

		if s.IPAddress == "" {
			continue
		}

		vm.Guest.IpAddress = s.IPAddress
		vm.Summary.Guest.IpAddress = s.IPAddress

		if len(vm.Guest.Net) != 0 {
			net := &vm.Guest.Net[0]
			net.IpAddress = []string{s.IPAddress}
			net.MacAddress = s.MacAddress
		}
	}

	return nil
}

func (c *container) prepareGuestOperation(vm *VirtualMachine, auth types.BaseGuestAuthentication) types.BaseMethodFault {
	if c.id == "" {
		return new(types.GuestOperationsUnavailable)
	}
	if vm.Runtime.PowerState != types.VirtualMachinePowerStatePoweredOn {
		return &types.InvalidPowerState{
			RequestedState: types.VirtualMachinePowerStatePoweredOn,
			ExistingState:  vm.Runtime.PowerState,
		}
	}
	switch creds := auth.(type) {
	case *types.NamePasswordAuthentication:
		if creds.Username == "" || creds.Password == "" {
			return new(types.InvalidGuestLogin)
		}
	default:
		return new(types.InvalidGuestLogin)
	}
	return nil
}

// createDMI writes BIOS UUID DMI files to a container volume
func (c *container) createDMI(vm *VirtualMachine, name string) error {
	cmd := exec.Command("docker", "run", "--rm", "-i", "-v", name+":"+"/"+name, "busybox", "tar", "-C", "/"+name, "-xf", "-")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	tw := tar.NewWriter(stdin)

	dmi := []struct {
		name string
		val  func(uuid.UUID) string
	}{
		{"product_uuid", productUUID},
		{"product_serial", productSerial},
	}

	for _, file := range dmi {
		val := file.val(vm.uid)
		_ = tw.WriteHeader(&tar.Header{
			Name:    file.name,
			Size:    int64(len(val) + 1),
			Mode:    0444,
			ModTime: time.Now(),
		})
		_, _ = fmt.Fprintln(tw, val)
	}

	_ = tw.Close()
	_ = stdin.Close()

	return cmd.Wait()
}

// start runs the container if specified by the RUN.container extraConfig property.
func (c *container) start(vm *VirtualMachine) {
	if c.id != "" {
		start := "start"
		if vm.Runtime.PowerState == types.VirtualMachinePowerStateSuspended {
			start = "unpause"
		}
		cmd := exec.Command("docker", start, c.id)
		err := cmd.Run()
		if err != nil {
			log.Printf("%s %s: %s", vm.Name, cmd.Args, err)
		}
		return
	}

	var args []string
	var env []string

	for _, opt := range vm.Config.ExtraConfig {
		val := opt.GetOptionValue()
		if val.Key == "RUN.container" {
			run := val.Value.(string)
			err := json.Unmarshal([]byte(run), &args)
			if err != nil {
				args = []string{run}
			}

			continue
		}
		if strings.HasPrefix(val.Key, "guestinfo.") {
			key := strings.Replace(strings.ToUpper(val.Key), ".", "_", -1)
			env = append(env, "--env", fmt.Sprintf("VMX_%s=%s", key, val.Value.(string)))
		}
	}

	if len(args) == 0 {
		return
	}
	if len(env) != 0 {
		// Configure env as the data access method for cloud-init-vmware-guestinfo
		env = append(env, "--env", "VMX_GUESTINFO=true")
	}

	c.name = fmt.Sprintf("vcsim-%s-%s", vm.Name, vm.uid)
	run := append([]string{"docker", "run", "-d", "--name", c.name}, env...)

	if err := c.createDMI(vm, c.name); err != nil {
		log.Printf("%s: %s", vm.Name, err)
		return
	}
	run = append(run, "-v", fmt.Sprintf("%s:%s:ro", c.name, "/sys/class/dmi/id"))

	args = append(run, args...)
	cmd := exec.Command(shell, "-c", strings.Join(args, " "))
	out, err := cmd.Output()
	if err != nil {
		log.Printf("%s %s: %s", vm.Name, cmd.Args, err)
		return
	}

	c.id = strings.TrimSpace(string(out))
	vm.logPrintf("%s %s: %s", cmd.Path, cmd.Args, c.id)

	if err = c.inspect(vm); err != nil {
		log.Printf("%s inspect %s: %s", vm.Name, c.id, err)
	}
}

// stop the container (if any) for the given vm.
func (c *container) stop(vm *VirtualMachine) {
	if c.id == "" {
		return
	}

	cmd := exec.Command("docker", "stop", c.id)
	err := cmd.Run()
	if err != nil {
		log.Printf("%s %s: %s", vm.Name, cmd.Args, err)
	}
}

// pause the container (if any) for the given vm.
func (c *container) pause(vm *VirtualMachine) {
	if c.id == "" {
		return
	}

	cmd := exec.Command("docker", "pause", c.id)
	err := cmd.Run()
	if err != nil {
		log.Printf("%s %s: %s", vm.Name, cmd.Args, err)
	}
}

// remove the container (if any) for the given vm.
func (c *container) remove(vm *VirtualMachine) {
	if c.id == "" {
		return
	}

	args := [][]string{
		[]string{"rm", "-v", "-f", c.id},
		[]string{"volume", "rm", "-f", c.name},
	}

	for i := range args {
		cmd := exec.Command("docker", args[i]...)
		err := cmd.Run()
		if err != nil {
			log.Printf("%s %s: %s", vm.Name, cmd.Args, err)
		}
	}

	c.id = ""
}

func guestUpload(file string, r *http.Request) error {
	cmd := exec.Command("docker", "cp", "-", path.Dir(file))
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err = cmd.Start(); err != nil {
		return err
	}

	tw := tar.NewWriter(stdin)
	_ = tw.WriteHeader(&tar.Header{
		Name:    path.Base(file),
		Size:    r.ContentLength,
		Mode:    0444,
		ModTime: time.Now(),
	})

	_, _ = io.Copy(tw, r.Body)

	_ = tw.Close()
	_ = stdin.Close()
	_ = r.Body.Close()

	return cmd.Wait()
}

func guestDownload(file string, w http.ResponseWriter) error {
	cmd := exec.Command("docker", "cp", file, "-")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err = cmd.Start(); err != nil {
		return err
	}

	tr := tar.NewReader(stdout)
	header, err := tr.Next()
	if err != nil {
		return err
	}

	w.Header().Set("Content-Length", strconv.FormatInt(header.Size, 10))
	_, _ = io.Copy(w, tr)

	_ = stdout.Close()

	return cmd.Wait()
}

const guestPrefix = "/guestFile/"

// ServeGuest handles container guest file upload/download
func ServeGuest(w http.ResponseWriter, r *http.Request) {
	// Real vCenter form: /guestFile?id=139&token=...
	// vcsim form:        /guestFile/tmp/foo/bar?id=ebc8837b8cb6&token=...

	id := r.URL.Query().Get("id")
	file := id + ":" + strings.TrimPrefix(r.URL.Path, guestPrefix[:len(guestPrefix)-1])
	var err error

	switch r.Method {
	case http.MethodPut:
		err = guestUpload(file, r)
	case http.MethodGet:
		err = guestDownload(file, w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err != nil {
		log.Printf("%s %s: %s", r.Method, r.URL, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// productSerial returns the uuid in /sys/class/dmi/id/product_serial format
func productSerial(id uuid.UUID) string {
	var dst [len(id)*2 + len(id) - 1]byte

	j := 0
	for i := 0; i < len(id); i++ {
		hex.Encode(dst[j:j+2], id[i:i+1])
		j += 3
		if j < len(dst) {
			s := j - 1
			if s == len(dst)/2 {
				dst[s] = '-'
			} else {
				dst[s] = ' '
			}
		}
	}

	return fmt.Sprintf("VMware-%s", string(dst[:]))
}

// productUUID returns the uuid in /sys/class/dmi/id/product_uuid format
func productUUID(id uuid.UUID) string {
	var dst [36]byte

	hex.Encode(dst[0:2], id[3:4])
	hex.Encode(dst[2:4], id[2:3])
	hex.Encode(dst[4:6], id[1:2])
	hex.Encode(dst[6:8], id[0:1])
	dst[8] = '-'
	hex.Encode(dst[9:11], id[5:6])
	hex.Encode(dst[11:13], id[4:5])
	dst[13] = '-'
	hex.Encode(dst[14:16], id[7:8])
	hex.Encode(dst[16:18], id[6:7])
	dst[18] = '-'
	hex.Encode(dst[19:23], id[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:], id[10:])

	return strings.ToUpper(string(dst[:]))
}
