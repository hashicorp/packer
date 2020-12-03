//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type CustomizeConfig,LinuxOptions,NetworkInterfaces,NetworkInterface,GlobalDnsSettings,GlobalRoutingSettings
package clone

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"

	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/vmware/govmomi/vim25/types"
)

// A cloned virtual machine can be [customized](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.vm_admin.doc/GUID-58E346FF-83AE-42B8-BE58-253641D257BC.html)
// to configure host, network, or licensing settings.
//
// To perform virtual machine customization as a part of the clone process, specify the customize block with the
// respective customization options. Windows guests are customized using Sysprep, which will result in the machine SID being reset.
// Before using customization, check that your source VM meets the [requirements](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.vm_admin.doc/GUID-E63B6FAA-8D35-428D-B40C-744769845906.html)
// for guest OS customization on vSphere.
// See the [customization example](#customization-example) for a usage synopsis.
//
// The settings for customize are as follows:
type CustomizeConfig struct {
	// Settings to Linux guest OS customization. See [Linux customization settings](#linux-customization-settings).
	LinuxOptions *LinuxOptions `mapstructure:"linux_options"`
	// Supply your own sysprep.xml file to allow full control of the customization process out-of-band of vSphere.
	WindowsSysPrepFile string `mapstructure:"windows_sysprep_file"`
	// Configure network interfaces on a per-interface basis that should matched up to the network adapters present in the VM.
	// To use DHCP, declare an empty network_interface for each adapter being configured. This field is required.
	// See [Network interface settings](#network-interface-settings).
	NetworkInterfaces     NetworkInterfaces `mapstructure:"network_interface"`
	GlobalRoutingSettings `mapstructure:",squash"`
	GlobalDnsSettings     `mapstructure:",squash"`
}

type LinuxOptions struct {
	// The domain name for this machine. This, along with [host_name](#host_name), make up the FQDN of this virtual machine.
	Domain string `mapstructure:"domain"`
	// The host name for this machine. This, along with [domain](#domain), make up the FQDN of this virtual machine.
	Hostname string `mapstructure:"host_name"`
	// Tells the operating system that the hardware clock is set to UTC. Default: true.
	HWClockUTC config.Trilean `mapstructure:"hw_clock_utc"`
	// Sets the time zone. The default is UTC.
	Timezone string `mapstructure:"time_zone"`
}

type NetworkInterface struct {
	// Network interface-specific DNS server settings for Windows operating systems.
	// Ignored on Linux and possibly other operating systems - for those systems, please see the [global DNS settings](#global-dns-settings) section.
	DnsServerList []string `mapstructure:"dns_server_list"`
	// Network interface-specific DNS search domain for Windows operating systems.
	// Ignored on Linux and possibly other operating systems - for those systems, please see the [global DNS settings](#global-dns-settings) section.
	DnsDomain string `mapstructure:"dns_domain"`
	// The IPv4 address assigned to this network adapter. If left blank or not included, DHCP is used.
	Ipv4Address string `mapstructure:"ipv4_address"`
	// The IPv4 subnet mask, in bits (example: 24 for 255.255.255.0).
	Ipv4NetMask int `mapstructure:"ipv4_netmask"`
	// The IPv6 address assigned to this network adapter. If left blank or not included, auto-configuration is used.
	Ipv6Address string `mapstructure:"ipv6_address"`
	// The IPv6 subnet mask, in bits (example: 32).
	Ipv6NetMask int `mapstructure:"ipv6_netmask"`
}

type NetworkInterfaces []NetworkInterface

// The settings here must match the IP/mask of at least one network_interface supplied to customization.
type GlobalRoutingSettings struct {
	// The IPv4 default gateway when using network_interface customization on the virtual machine.
	Ipv4Gateway string `mapstructure:"ipv4_gateway"`
	// The IPv6 default gateway when using network_interface customization on the virtual machine.
	Ipv6Gateway string `mapstructure:"ipv6_gateway"`
}

// The following settings configure DNS globally, generally for Linux systems. For Windows systems,
// this is done per-interface, see [network interface](#network_interface) settings.
type GlobalDnsSettings struct {
	// The list of DNS servers to configure on a virtual machine.
	DnsServerList []string `mapstructure:"dns_server_list"`
	// A list of DNS search domains to add to the DNS configuration on the virtual machine.
	DnsSuffixList []string `mapstructure:"dns_suffix_list"`
}

type StepCustomize struct {
	Config *CustomizeConfig
}

func (c *CustomizeConfig) Prepare() []error {
	var errs []error

	if c.LinuxOptions == nil && c.WindowsSysPrepFile == "" {
		errs = append(errs, fmt.Errorf("customize is empty"))
	}
	if c.LinuxOptions != nil && c.WindowsSysPrepFile != "" {
		errs = append(errs, fmt.Errorf("`linux_options` and `windows_sysprep_text` both set - one must not be included if the other is specified"))
	}

	if c.LinuxOptions != nil {
		if c.LinuxOptions.Hostname == "" {
			errs = append(errs, fmt.Errorf("linux options `host_name` is empty"))
		}
		if c.LinuxOptions.Domain == "" {
			errs = append(errs, fmt.Errorf("linux options `domain` is empty"))
		}

		if c.LinuxOptions.HWClockUTC == config.TriUnset {
			c.LinuxOptions.HWClockUTC = config.TriTrue
		}
		if c.LinuxOptions.Timezone == "" {
			c.LinuxOptions.Timezone = "UTC"
		}
	}

	if len(c.NetworkInterfaces) == 0 {
		errs = append(errs, fmt.Errorf("one or more `network_interface` must be provided"))
	}

	return errs
}

func (s *StepCustomize) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	vm := state.Get("vm").(*driver.VirtualMachineDriver)
	ui := state.Get("ui").(packersdk.Ui)

	identity, err := s.identitySettings()
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	nicSettingsMap := s.nicSettingsMap()
	globalIpSettings := s.globalIpSettings()

	spec := types.CustomizationSpec{
		Identity:         identity,
		NicSettingMap:    nicSettingsMap,
		GlobalIPSettings: globalIpSettings,
	}
	ui.Say("Customizing VM...")
	err = vm.Customize(spec)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepCustomize) identitySettings() (types.BaseCustomizationIdentitySettings, error) {
	if s.Config.LinuxOptions != nil {
		return &types.CustomizationLinuxPrep{
			HostName: &types.CustomizationFixedName{
				Name: s.Config.LinuxOptions.Hostname,
			},
			Domain:     s.Config.LinuxOptions.Domain,
			TimeZone:   s.Config.LinuxOptions.Timezone,
			HwClockUTC: s.Config.LinuxOptions.HWClockUTC.ToBoolPointer(),
		}, nil
	}

	if s.Config.WindowsSysPrepFile != "" {
		sysPrep, err := ioutil.ReadFile(s.Config.WindowsSysPrepFile)
		if err != nil {
			return nil, fmt.Errorf("error on reading %s: %s", s.Config.WindowsSysPrepFile, err)
		}
		return &types.CustomizationSysprepText{
			Value: string(sysPrep),
		}, nil
	}

	return nil, fmt.Errorf("no customization identity found")
}

func (s *StepCustomize) nicSettingsMap() []types.CustomizationAdapterMapping {
	result := make([]types.CustomizationAdapterMapping, len(s.Config.NetworkInterfaces))
	var ipv4gwFound, ipv6gwFound bool
	for i := range s.Config.NetworkInterfaces {
		var adapter types.CustomizationIPSettings
		adapter, ipv4gwFound, ipv6gwFound = s.ipSettings(i, !ipv4gwFound, !ipv6gwFound)
		obj := types.CustomizationAdapterMapping{
			Adapter: adapter,
		}
		result[i] = obj
	}
	return result
}

func (s *StepCustomize) ipSettings(n int, ipv4gwAdd bool, ipv6gwAdd bool) (types.CustomizationIPSettings, bool, bool) {
	var v4gwFound, v6gwFound bool
	var obj types.CustomizationIPSettings

	ipv4Address := s.Config.NetworkInterfaces[n].Ipv4Address
	if ipv4Address != "" {
		ipv4mask := s.Config.NetworkInterfaces[n].Ipv4NetMask
		ipv4Gateway := s.Config.Ipv4Gateway
		obj.Ip = &types.CustomizationFixedIp{
			IpAddress: ipv4Address,
		}
		obj.SubnetMask = v4CIDRMaskToDotted(ipv4mask)
		// Check for the gateway
		if ipv4gwAdd && ipv4Gateway != "" && matchGateway(ipv4Address, ipv4mask, ipv4Gateway) {
			obj.Gateway = []string{ipv4Gateway}
			v4gwFound = true
		}
	} else {
		obj.Ip = &types.CustomizationDhcpIpGenerator{}
	}

	obj.DnsServerList = s.Config.NetworkInterfaces[n].DnsServerList
	obj.DnsDomain = s.Config.NetworkInterfaces[n].DnsDomain
	obj.IpV6Spec, v6gwFound = s.IPSettingsIPV6Address(n, ipv6gwAdd)

	return obj, v4gwFound, v6gwFound
}

func v4CIDRMaskToDotted(mask int) string {
	m := net.CIDRMask(mask, 32)
	a := int(m[0])
	b := int(m[1])
	c := int(m[2])
	d := int(m[3])
	return fmt.Sprintf("%d.%d.%d.%d", a, b, c, d)
}

func (s *StepCustomize) IPSettingsIPV6Address(n int, gwAdd bool) (*types.CustomizationIPSettingsIpV6AddressSpec, bool) {
	addr := s.Config.NetworkInterfaces[n].Ipv6Address
	var gwFound bool
	if addr == "" {
		return nil, gwFound
	}
	mask := s.Config.NetworkInterfaces[n].Ipv6NetMask
	gw := s.Config.Ipv6Gateway
	obj := &types.CustomizationIPSettingsIpV6AddressSpec{
		Ip: []types.BaseCustomizationIpV6Generator{
			&types.CustomizationFixedIpV6{
				IpAddress:  addr,
				SubnetMask: int32(mask),
			},
		},
	}
	if gwAdd && gw != "" && matchGateway(addr, mask, gw) {
		obj.Gateway = []string{gw}
		gwFound = true
	}
	return obj, gwFound
}

// matchGateway take an IP, mask, and gateway, and checks to see if the gateway
// is reachable from the IP address.
func matchGateway(a string, m int, g string) bool {
	ip := net.ParseIP(a)
	gw := net.ParseIP(g)
	var mask net.IPMask
	if ip.To4() != nil {
		mask = net.CIDRMask(m, 32)
	} else {
		mask = net.CIDRMask(m, 128)
	}
	if ip.Mask(mask).Equal(gw.Mask(mask)) {
		return true
	}
	return false
}

func (s *StepCustomize) globalIpSettings() types.CustomizationGlobalIPSettings {
	return types.CustomizationGlobalIPSettings{
		DnsServerList: s.Config.DnsServerList,
		DnsSuffixList: s.Config.DnsSuffixList,
	}
}

func (s *StepCustomize) Cleanup(_ multistep.StateBag) {}
