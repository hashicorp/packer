package qemu

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"

	"github.com/digitalocean/go-qemu/qmp"
)

// This step waits for the guest address to become available in the network
// bridge, then it sets the guestAddress state property.
type stepWaitGuestAddress struct {
	CommunicatorType string
	NetBridge        string

	timeout time.Duration
}

func (s *stepWaitGuestAddress) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	if s.CommunicatorType == "none" {
		ui.Message("No communicator is configured -- skipping StepWaitGuestAddress")
		return multistep.ActionContinue
	}
	if s.NetBridge == "" {
		ui.Message("Not using a NetBridge -- skipping StepWaitGuestAddress")
		return multistep.ActionContinue
	}

	qmpMonitor := state.Get("qmp_monitor").(*qmp.SocketMonitor)
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	ui.Say(fmt.Sprintf("Waiting for the guest address to become available in the %s network bridge...", s.NetBridge))
	for {
		guestAddress := getGuestAddress(qmpMonitor, s.NetBridge, "user.0")
		if guestAddress != "" {
			log.Printf("Found guest address %s", guestAddress)
			state.Put("guestAddress", guestAddress)
			return multistep.ActionContinue
		}
		select {
		case <-time.After(10 * time.Second):
			continue
		case <-ctx.Done():
			return multistep.ActionHalt
		}
	}
}

func (s *stepWaitGuestAddress) Cleanup(state multistep.StateBag) {
}

func getGuestAddress(qmpMonitor *qmp.SocketMonitor, bridgeName string, deviceName string) string {
	devices, err := getNetDevices(qmpMonitor)
	if err != nil {
		log.Printf("Could not retrieve QEMU QMP network device list: %v", err)
		return ""
	}

	for _, device := range devices {
		if device.Name == deviceName {
			ipAddress, _ := getDeviceIPAddress(bridgeName, device.MacAddress)
			return ipAddress
		}
	}

	log.Printf("QEMU QMP network device %s was not found", deviceName)
	return ""
}

func getDeviceIPAddress(device string, macAddress string) (string, error) {
	// this parses /proc/net/arp to retrieve the given device IP address.
	//
	// /proc/net/arp is normally someting alike:
	//
	// 		IP address       HW type     Flags       HW address            Mask     Device
	// 		192.168.121.111  0x1         0x2         52:54:00:12:34:56     *        virbr0
	//

	const (
		IPAddressIndex int = iota
		HWTypeIndex
		FlagsIndex
		HWAddressIndex
		MaskIndex
		DeviceIndex
	)

	// see ARP flags at https://github.com/torvalds/linux/blob/v5.4/include/uapi/linux/if_arp.h#L132
	const (
		AtfCom int = 0x02 // ATF_COM (complete)
	)

	f, err := os.Open("/proc/net/arp")
	if err != nil {
		return "", fmt.Errorf("failed to open /proc/net/arp: %w", err)
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	s.Scan()

	for s.Scan() {
		fields := strings.Fields(s.Text())

		if device != "" && fields[DeviceIndex] != device {
			continue
		}

		if fields[HWAddressIndex] != macAddress {
			continue
		}

		flags, err := strconv.ParseInt(fields[FlagsIndex], 0, 32)
		if err != nil {
			return "", fmt.Errorf("failed to parse /proc/net/arp flags field %s: %w", fields[FlagsIndex], err)
		}

		if int(flags)&AtfCom == AtfCom {
			return fields[IPAddressIndex], nil
		}
	}

	return "", fmt.Errorf("could not find %s", macAddress)
}
