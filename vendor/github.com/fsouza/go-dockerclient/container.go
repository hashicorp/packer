// Copyright 2013 go-dockerclient authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/docker/go-units"
	"golang.org/x/net/context"
)

// ErrContainerAlreadyExists is the error returned by CreateContainer when the
// container already exists.
var ErrContainerAlreadyExists = errors.New("container already exists")

// ListContainersOptions specify parameters to the ListContainers function.
//
// See https://goo.gl/kaOHGw for more details.
type ListContainersOptions struct {
	All     bool
	Size    bool
	Limit   int
	Since   string
	Before  string
	Filters map[string][]string
	Context context.Context
}

// APIPort is a type that represents a port mapping returned by the Docker API
type APIPort struct {
	PrivatePort int64  `json:"PrivatePort,omitempty" yaml:"PrivatePort,omitempty"`
	PublicPort  int64  `json:"PublicPort,omitempty" yaml:"PublicPort,omitempty"`
	Type        string `json:"Type,omitempty" yaml:"Type,omitempty"`
	IP          string `json:"IP,omitempty" yaml:"IP,omitempty"`
}

// APIMount represents a mount point for a container.
type APIMount struct {
	Name        string `json:"Name,omitempty" yaml:"Name,omitempty"`
	Source      string `json:"Source,omitempty" yaml:"Source,omitempty"`
	Destination string `json:"Destination,omitempty" yaml:"Destination,omitempty"`
	Driver      string `json:"Driver,omitempty" yaml:"Driver,omitempty"`
	Mode        string `json:"Mode,omitempty" yaml:"Mode,omitempty"`
	RW          bool   `json:"RW,omitempty" yaml:"RW,omitempty"`
	Propogation string `json:"Propogation,omitempty" yaml:"Propogation,omitempty"`
}

// APIContainers represents each container in the list returned by
// ListContainers.
type APIContainers struct {
	ID         string            `json:"Id" yaml:"Id"`
	Image      string            `json:"Image,omitempty" yaml:"Image,omitempty"`
	Command    string            `json:"Command,omitempty" yaml:"Command,omitempty"`
	Created    int64             `json:"Created,omitempty" yaml:"Created,omitempty"`
	State      string            `json:"State,omitempty" yaml:"State,omitempty"`
	Status     string            `json:"Status,omitempty" yaml:"Status,omitempty"`
	Ports      []APIPort         `json:"Ports,omitempty" yaml:"Ports,omitempty"`
	SizeRw     int64             `json:"SizeRw,omitempty" yaml:"SizeRw,omitempty"`
	SizeRootFs int64             `json:"SizeRootFs,omitempty" yaml:"SizeRootFs,omitempty"`
	Names      []string          `json:"Names,omitempty" yaml:"Names,omitempty"`
	Labels     map[string]string `json:"Labels,omitempty" yaml:"Labels,omitempty"`
	Networks   NetworkList       `json:"NetworkSettings,omitempty" yaml:"NetworkSettings,omitempty"`
	Mounts     []APIMount        `json:"Mounts,omitempty" yaml:"Mounts,omitempty"`
}

// NetworkList encapsulates a map of networks, as returned by the Docker API in
// ListContainers.
type NetworkList struct {
	Networks map[string]ContainerNetwork `json:"Networks" yaml:"Networks,omitempty"`
}

// ListContainers returns a slice of containers matching the given criteria.
//
// See https://goo.gl/kaOHGw for more details.
func (c *Client) ListContainers(opts ListContainersOptions) ([]APIContainers, error) {
	path := "/containers/json?" + queryString(opts)
	resp, err := c.do("GET", path, doOptions{context: opts.Context})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var containers []APIContainers
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, err
	}
	return containers, nil
}

// Port represents the port number and the protocol, in the form
// <number>/<protocol>. For example: 80/tcp.
type Port string

// Port returns the number of the port.
func (p Port) Port() string {
	return strings.Split(string(p), "/")[0]
}

// Proto returns the name of the protocol.
func (p Port) Proto() string {
	parts := strings.Split(string(p), "/")
	if len(parts) == 1 {
		return "tcp"
	}
	return parts[1]
}

// HealthCheck represents one check of health.
type HealthCheck struct {
	Start    time.Time `json:"Start,omitempty" yaml:"Start,omitempty"`
	End      time.Time `json:"End,omitempty" yaml:"End,omitempty"`
	ExitCode int       `json:"ExitCode,omitempty" yaml:"ExitCode,omitempty"`
	Output   string    `json:"Output,omitempty" yaml:"Output,omitempty"`
}

// Health represents the health of a container.
type Health struct {
	Status        string        `json:"Status,omitempty" yaml:"Status,omitempty"`
	FailingStreak int           `json:"FailingStreak,omitempty" yaml:"FailingStreak,omitempty"`
	Log           []HealthCheck `json:"Log,omitempty" yaml:"Log,omitempty"`
}

// State represents the state of a container.
type State struct {
	Status            string    `json:"Status,omitempty" yaml:"Status,omitempty"`
	Running           bool      `json:"Running,omitempty" yaml:"Running,omitempty"`
	Paused            bool      `json:"Paused,omitempty" yaml:"Paused,omitempty"`
	Restarting        bool      `json:"Restarting,omitempty" yaml:"Restarting,omitempty"`
	OOMKilled         bool      `json:"OOMKilled,omitempty" yaml:"OOMKilled,omitempty"`
	RemovalInProgress bool      `json:"RemovalInProgress,omitempty" yaml:"RemovalInProgress,omitempty"`
	Dead              bool      `json:"Dead,omitempty" yaml:"Dead,omitempty"`
	Pid               int       `json:"Pid,omitempty" yaml:"Pid,omitempty"`
	ExitCode          int       `json:"ExitCode,omitempty" yaml:"ExitCode,omitempty"`
	Error             string    `json:"Error,omitempty" yaml:"Error,omitempty"`
	StartedAt         time.Time `json:"StartedAt,omitempty" yaml:"StartedAt,omitempty"`
	FinishedAt        time.Time `json:"FinishedAt,omitempty" yaml:"FinishedAt,omitempty"`
	Health            Health    `json:"Health,omitempty" yaml:"Health,omitempty"`
}

// String returns a human-readable description of the state
func (s *State) String() string {
	if s.Running {
		if s.Paused {
			return fmt.Sprintf("Up %s (Paused)", units.HumanDuration(time.Now().UTC().Sub(s.StartedAt)))
		}
		if s.Restarting {
			return fmt.Sprintf("Restarting (%d) %s ago", s.ExitCode, units.HumanDuration(time.Now().UTC().Sub(s.FinishedAt)))
		}

		return fmt.Sprintf("Up %s", units.HumanDuration(time.Now().UTC().Sub(s.StartedAt)))
	}

	if s.RemovalInProgress {
		return "Removal In Progress"
	}

	if s.Dead {
		return "Dead"
	}

	if s.StartedAt.IsZero() {
		return "Created"
	}

	if s.FinishedAt.IsZero() {
		return ""
	}

	return fmt.Sprintf("Exited (%d) %s ago", s.ExitCode, units.HumanDuration(time.Now().UTC().Sub(s.FinishedAt)))
}

// StateString returns a single string to describe state
func (s *State) StateString() string {
	if s.Running {
		if s.Paused {
			return "paused"
		}
		if s.Restarting {
			return "restarting"
		}
		return "running"
	}

	if s.Dead {
		return "dead"
	}

	if s.StartedAt.IsZero() {
		return "created"
	}

	return "exited"
}

// PortBinding represents the host/container port mapping as returned in the
// `docker inspect` json
type PortBinding struct {
	HostIP   string `json:"HostIP,omitempty" yaml:"HostIP,omitempty"`
	HostPort string `json:"HostPort,omitempty" yaml:"HostPort,omitempty"`
}

// PortMapping represents a deprecated field in the `docker inspect` output,
// and its value as found in NetworkSettings should always be nil
type PortMapping map[string]string

// ContainerNetwork represents the networking settings of a container per network.
type ContainerNetwork struct {
	MacAddress          string `json:"MacAddress,omitempty" yaml:"MacAddress,omitempty"`
	GlobalIPv6PrefixLen int    `json:"GlobalIPv6PrefixLen,omitempty" yaml:"GlobalIPv6PrefixLen,omitempty"`
	GlobalIPv6Address   string `json:"GlobalIPv6Address,omitempty" yaml:"GlobalIPv6Address,omitempty"`
	IPv6Gateway         string `json:"IPv6Gateway,omitempty" yaml:"IPv6Gateway,omitempty"`
	IPPrefixLen         int    `json:"IPPrefixLen,omitempty" yaml:"IPPrefixLen,omitempty"`
	IPAddress           string `json:"IPAddress,omitempty" yaml:"IPAddress,omitempty"`
	Gateway             string `json:"Gateway,omitempty" yaml:"Gateway,omitempty"`
	EndpointID          string `json:"EndpointID,omitempty" yaml:"EndpointID,omitempty"`
	NetworkID           string `json:"NetworkID,omitempty" yaml:"NetworkID,omitempty"`
}

// NetworkSettings contains network-related information about a container
type NetworkSettings struct {
	Networks               map[string]ContainerNetwork `json:"Networks,omitempty" yaml:"Networks,omitempty"`
	IPAddress              string                      `json:"IPAddress,omitempty" yaml:"IPAddress,omitempty"`
	IPPrefixLen            int                         `json:"IPPrefixLen,omitempty" yaml:"IPPrefixLen,omitempty"`
	MacAddress             string                      `json:"MacAddress,omitempty" yaml:"MacAddress,omitempty"`
	Gateway                string                      `json:"Gateway,omitempty" yaml:"Gateway,omitempty"`
	Bridge                 string                      `json:"Bridge,omitempty" yaml:"Bridge,omitempty"`
	PortMapping            map[string]PortMapping      `json:"PortMapping,omitempty" yaml:"PortMapping,omitempty"`
	Ports                  map[Port][]PortBinding      `json:"Ports,omitempty" yaml:"Ports,omitempty"`
	NetworkID              string                      `json:"NetworkID,omitempty" yaml:"NetworkID,omitempty"`
	EndpointID             string                      `json:"EndpointID,omitempty" yaml:"EndpointID,omitempty"`
	SandboxKey             string                      `json:"SandboxKey,omitempty" yaml:"SandboxKey,omitempty"`
	GlobalIPv6Address      string                      `json:"GlobalIPv6Address,omitempty" yaml:"GlobalIPv6Address,omitempty"`
	GlobalIPv6PrefixLen    int                         `json:"GlobalIPv6PrefixLen,omitempty" yaml:"GlobalIPv6PrefixLen,omitempty"`
	IPv6Gateway            string                      `json:"IPv6Gateway,omitempty" yaml:"IPv6Gateway,omitempty"`
	LinkLocalIPv6Address   string                      `json:"LinkLocalIPv6Address,omitempty" yaml:"LinkLocalIPv6Address,omitempty"`
	LinkLocalIPv6PrefixLen int                         `json:"LinkLocalIPv6PrefixLen,omitempty" yaml:"LinkLocalIPv6PrefixLen,omitempty"`
	SecondaryIPAddresses   []string                    `json:"SecondaryIPAddresses,omitempty" yaml:"SecondaryIPAddresses,omitempty"`
	SecondaryIPv6Addresses []string                    `json:"SecondaryIPv6Addresses,omitempty" yaml:"SecondaryIPv6Addresses,omitempty"`
}

// PortMappingAPI translates the port mappings as contained in NetworkSettings
// into the format in which they would appear when returned by the API
func (settings *NetworkSettings) PortMappingAPI() []APIPort {
	var mapping []APIPort
	for port, bindings := range settings.Ports {
		p, _ := parsePort(port.Port())
		if len(bindings) == 0 {
			mapping = append(mapping, APIPort{
				PrivatePort: int64(p),
				Type:        port.Proto(),
			})
			continue
		}
		for _, binding := range bindings {
			p, _ := parsePort(port.Port())
			h, _ := parsePort(binding.HostPort)
			mapping = append(mapping, APIPort{
				PrivatePort: int64(p),
				PublicPort:  int64(h),
				Type:        port.Proto(),
				IP:          binding.HostIP,
			})
		}
	}
	return mapping
}

func parsePort(rawPort string) (int, error) {
	port, err := strconv.ParseUint(rawPort, 10, 16)
	if err != nil {
		return 0, err
	}
	return int(port), nil
}

// Config is the list of configuration options used when creating a container.
// Config does not contain the options that are specific to starting a container on a
// given host.  Those are contained in HostConfig
type Config struct {
	Hostname          string              `json:"Hostname,omitempty" yaml:"Hostname,omitempty"`
	Domainname        string              `json:"Domainname,omitempty" yaml:"Domainname,omitempty"`
	User              string              `json:"User,omitempty" yaml:"User,omitempty"`
	Memory            int64               `json:"Memory,omitempty" yaml:"Memory,omitempty"`
	MemorySwap        int64               `json:"MemorySwap,omitempty" yaml:"MemorySwap,omitempty"`
	MemoryReservation int64               `json:"MemoryReservation,omitempty" yaml:"MemoryReservation,omitempty"`
	KernelMemory      int64               `json:"KernelMemory,omitempty" yaml:"KernelMemory,omitempty"`
	CPUShares         int64               `json:"CpuShares,omitempty" yaml:"CpuShares,omitempty"`
	CPUSet            string              `json:"Cpuset,omitempty" yaml:"Cpuset,omitempty"`
	PortSpecs         []string            `json:"PortSpecs,omitempty" yaml:"PortSpecs,omitempty"`
	ExposedPorts      map[Port]struct{}   `json:"ExposedPorts,omitempty" yaml:"ExposedPorts,omitempty"`
	PublishService    string              `json:"PublishService,omitempty" yaml:"PublishService,omitempty"`
	StopSignal        string              `json:"StopSignal,omitempty" yaml:"StopSignal,omitempty"`
	Env               []string            `json:"Env,omitempty" yaml:"Env,omitempty"`
	Cmd               []string            `json:"Cmd" yaml:"Cmd"`
	Healthcheck       *HealthConfig       `json:"Healthcheck,omitempty" yaml:"Healthcheck,omitempty"`
	DNS               []string            `json:"Dns,omitempty" yaml:"Dns,omitempty"` // For Docker API v1.9 and below only
	Image             string              `json:"Image,omitempty" yaml:"Image,omitempty"`
	Volumes           map[string]struct{} `json:"Volumes,omitempty" yaml:"Volumes,omitempty"`
	VolumeDriver      string              `json:"VolumeDriver,omitempty" yaml:"VolumeDriver,omitempty"`
	WorkingDir        string              `json:"WorkingDir,omitempty" yaml:"WorkingDir,omitempty"`
	MacAddress        string              `json:"MacAddress,omitempty" yaml:"MacAddress,omitempty"`
	Entrypoint        []string            `json:"Entrypoint" yaml:"Entrypoint"`
	SecurityOpts      []string            `json:"SecurityOpts,omitempty" yaml:"SecurityOpts,omitempty"`
	OnBuild           []string            `json:"OnBuild,omitempty" yaml:"OnBuild,omitempty"`
	Mounts            []Mount             `json:"Mounts,omitempty" yaml:"Mounts,omitempty"`
	Labels            map[string]string   `json:"Labels,omitempty" yaml:"Labels,omitempty"`
	AttachStdin       bool                `json:"AttachStdin,omitempty" yaml:"AttachStdin,omitempty"`
	AttachStdout      bool                `json:"AttachStdout,omitempty" yaml:"AttachStdout,omitempty"`
	AttachStderr      bool                `json:"AttachStderr,omitempty" yaml:"AttachStderr,omitempty"`
	ArgsEscaped       bool                `json:"ArgsEscaped,omitempty" yaml:"ArgsEscaped,omitempty"`
	Tty               bool                `json:"Tty,omitempty" yaml:"Tty,omitempty"`
	OpenStdin         bool                `json:"OpenStdin,omitempty" yaml:"OpenStdin,omitempty"`
	StdinOnce         bool                `json:"StdinOnce,omitempty" yaml:"StdinOnce,omitempty"`
	NetworkDisabled   bool                `json:"NetworkDisabled,omitempty" yaml:"NetworkDisabled,omitempty"`

	// This is no longer used and has been kept here for backward
	// compatibility, please use HostConfig.VolumesFrom.
	VolumesFrom string `json:"VolumesFrom,omitempty" yaml:"VolumesFrom,omitempty"`
}

// Mount represents a mount point in the container.
//
// It has been added in the version 1.20 of the Docker API, available since
// Docker 1.8.
type Mount struct {
	Name        string
	Source      string
	Destination string
	Driver      string
	Mode        string
	RW          bool
}

// LogConfig defines the log driver type and the configuration for it.
type LogConfig struct {
	Type   string            `json:"Type,omitempty" yaml:"Type,omitempty"`
	Config map[string]string `json:"Config,omitempty" yaml:"Config,omitempty"`
}

// ULimit defines system-wide resource limitations This can help a lot in
// system administration, e.g. when a user starts too many processes and
// therefore makes the system unresponsive for other users.
type ULimit struct {
	Name string `json:"Name,omitempty" yaml:"Name,omitempty"`
	Soft int64  `json:"Soft,omitempty" yaml:"Soft,omitempty"`
	Hard int64  `json:"Hard,omitempty" yaml:"Hard,omitempty"`
}

// SwarmNode containers information about which Swarm node the container is on.
type SwarmNode struct {
	ID     string            `json:"ID,omitempty" yaml:"ID,omitempty"`
	IP     string            `json:"IP,omitempty" yaml:"IP,omitempty"`
	Addr   string            `json:"Addr,omitempty" yaml:"Addr,omitempty"`
	Name   string            `json:"Name,omitempty" yaml:"Name,omitempty"`
	CPUs   int64             `json:"CPUs,omitempty" yaml:"CPUs,omitempty"`
	Memory int64             `json:"Memory,omitempty" yaml:"Memory,omitempty"`
	Labels map[string]string `json:"Labels,omitempty" yaml:"Labels,omitempty"`
}

// GraphDriver contains information about the GraphDriver used by the
// container.
type GraphDriver struct {
	Name string            `json:"Name,omitempty" yaml:"Name,omitempty"`
	Data map[string]string `json:"Data,omitempty" yaml:"Data,omitempty"`
}

// HealthConfig holds configuration settings for the HEALTHCHECK feature
//
// It has been added in the version 1.24 of the Docker API, available since
// Docker 1.12.
type HealthConfig struct {
	// Test is the test to perform to check that the container is healthy.
	// An empty slice means to inherit the default.
	// The options are:
	// {} : inherit healthcheck
	// {"NONE"} : disable healthcheck
	// {"CMD", args...} : exec arguments directly
	// {"CMD-SHELL", command} : run command with system's default shell
	Test []string `json:"Test,omitempty" yaml:"Test,omitempty"`

	// Zero means to inherit. Durations are expressed as integer nanoseconds.
	Interval time.Duration `json:"Interval,omitempty" yaml:"Interval,omitempty"` // Interval is the time to wait between checks.
	Timeout  time.Duration `json:"Timeout,omitempty" yaml:"Timeout,omitempty"`   // Timeout is the time to wait before considering the check to have hung.

	// Retries is the number of consecutive failures needed to consider a container as unhealthy.
	// Zero means inherit.
	Retries int `json:"Retries,omitempty" yaml:"Retries,omitempty"`
}

// Container is the type encompasing everything about a container - its config,
// hostconfig, etc.
type Container struct {
	ID string `json:"Id" yaml:"Id"`

	Created time.Time `json:"Created,omitempty" yaml:"Created,omitempty"`

	Path string   `json:"Path,omitempty" yaml:"Path,omitempty"`
	Args []string `json:"Args,omitempty" yaml:"Args,omitempty"`

	Config *Config `json:"Config,omitempty" yaml:"Config,omitempty"`
	State  State   `json:"State,omitempty" yaml:"State,omitempty"`
	Image  string  `json:"Image,omitempty" yaml:"Image,omitempty"`

	Node *SwarmNode `json:"Node,omitempty" yaml:"Node,omitempty"`

	NetworkSettings *NetworkSettings `json:"NetworkSettings,omitempty" yaml:"NetworkSettings,omitempty"`

	SysInitPath    string  `json:"SysInitPath,omitempty" yaml:"SysInitPath,omitempty"`
	ResolvConfPath string  `json:"ResolvConfPath,omitempty" yaml:"ResolvConfPath,omitempty"`
	HostnamePath   string  `json:"HostnamePath,omitempty" yaml:"HostnamePath,omitempty"`
	HostsPath      string  `json:"HostsPath,omitempty" yaml:"HostsPath,omitempty"`
	LogPath        string  `json:"LogPath,omitempty" yaml:"LogPath,omitempty"`
	Name           string  `json:"Name,omitempty" yaml:"Name,omitempty"`
	Driver         string  `json:"Driver,omitempty" yaml:"Driver,omitempty"`
	Mounts         []Mount `json:"Mounts,omitempty" yaml:"Mounts,omitempty"`

	Volumes     map[string]string `json:"Volumes,omitempty" yaml:"Volumes,omitempty"`
	VolumesRW   map[string]bool   `json:"VolumesRW,omitempty" yaml:"VolumesRW,omitempty"`
	HostConfig  *HostConfig       `json:"HostConfig,omitempty" yaml:"HostConfig,omitempty"`
	ExecIDs     []string          `json:"ExecIDs,omitempty" yaml:"ExecIDs,omitempty"`
	GraphDriver *GraphDriver      `json:"GraphDriver,omitempty" yaml:"GraphDriver,omitempty"`

	RestartCount int `json:"RestartCount,omitempty" yaml:"RestartCount,omitempty"`

	AppArmorProfile string `json:"AppArmorProfile,omitempty" yaml:"AppArmorProfile,omitempty"`
}

// UpdateContainerOptions specify parameters to the UpdateContainer function.
//
// See https://goo.gl/Y6fXUy for more details.
type UpdateContainerOptions struct {
	BlkioWeight       int           `json:"BlkioWeight"`
	CPUShares         int           `json:"CpuShares"`
	CPUPeriod         int           `json:"CpuPeriod"`
	CPUQuota          int           `json:"CpuQuota"`
	CpusetCpus        string        `json:"CpusetCpus"`
	CpusetMems        string        `json:"CpusetMems"`
	Memory            int           `json:"Memory"`
	MemorySwap        int           `json:"MemorySwap"`
	MemoryReservation int           `json:"MemoryReservation"`
	KernelMemory      int           `json:"KernelMemory"`
	RestartPolicy     RestartPolicy `json:"RestartPolicy,omitempty"`
	Context           context.Context
}

// UpdateContainer updates the container at ID with the options
//
// See https://goo.gl/Y6fXUy for more details.
func (c *Client) UpdateContainer(id string, opts UpdateContainerOptions) error {
	resp, err := c.do("POST", fmt.Sprintf("/containers/"+id+"/update"), doOptions{
		data:      opts,
		forceJSON: true,
		context:   opts.Context,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// RenameContainerOptions specify parameters to the RenameContainer function.
//
// See https://goo.gl/46inai for more details.
type RenameContainerOptions struct {
	// ID of container to rename
	ID string `qs:"-"`

	// New name
	Name    string `json:"name,omitempty" yaml:"name,omitempty"`
	Context context.Context
}

// RenameContainer updates and existing containers name
//
// See https://goo.gl/46inai for more details.
func (c *Client) RenameContainer(opts RenameContainerOptions) error {
	resp, err := c.do("POST", fmt.Sprintf("/containers/"+opts.ID+"/rename?%s", queryString(opts)), doOptions{
		context: opts.Context,
	})
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// InspectContainer returns information about a container by its ID.
//
// See https://goo.gl/FaI5JT for more details.
func (c *Client) InspectContainer(id string) (*Container, error) {
	return c.inspectContainer(id, doOptions{})
}

// InspectContainerWithContext returns information about a container by its ID.
// The context object can be used to cancel the inspect request.
//
// See https://goo.gl/FaI5JT for more details.
func (c *Client) InspectContainerWithContext(id string, ctx context.Context) (*Container, error) {
	return c.inspectContainer(id, doOptions{context: ctx})
}

func (c *Client) inspectContainer(id string, opts doOptions) (*Container, error) {
	path := "/containers/" + id + "/json"
	resp, err := c.do("GET", path, opts)
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return nil, &NoSuchContainer{ID: id}
		}
		return nil, err
	}
	defer resp.Body.Close()
	var container Container
	if err := json.NewDecoder(resp.Body).Decode(&container); err != nil {
		return nil, err
	}
	return &container, nil
}

// ContainerChanges returns changes in the filesystem of the given container.
//
// See https://goo.gl/15KKzh for more details.
func (c *Client) ContainerChanges(id string) ([]Change, error) {
	path := "/containers/" + id + "/changes"
	resp, err := c.do("GET", path, doOptions{})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return nil, &NoSuchContainer{ID: id}
		}
		return nil, err
	}
	defer resp.Body.Close()
	var changes []Change
	if err := json.NewDecoder(resp.Body).Decode(&changes); err != nil {
		return nil, err
	}
	return changes, nil
}

// CreateContainerOptions specify parameters to the CreateContainer function.
//
// See https://goo.gl/tyzwVM for more details.
type CreateContainerOptions struct {
	Name             string
	Config           *Config           `qs:"-"`
	HostConfig       *HostConfig       `qs:"-"`
	NetworkingConfig *NetworkingConfig `qs:"-"`
	Context          context.Context
}

// CreateContainer creates a new container, returning the container instance,
// or an error in case of failure.
//
// The returned container instance contains only the container ID. To get more
// details about the container after creating it, use InspectContainer.
//
// See https://goo.gl/tyzwVM for more details.
func (c *Client) CreateContainer(opts CreateContainerOptions) (*Container, error) {
	path := "/containers/create?" + queryString(opts)
	resp, err := c.do(
		"POST",
		path,
		doOptions{
			data: struct {
				*Config
				HostConfig       *HostConfig       `json:"HostConfig,omitempty" yaml:"HostConfig,omitempty"`
				NetworkingConfig *NetworkingConfig `json:"NetworkingConfig,omitempty" yaml:"NetworkingConfig,omitempty"`
			}{
				opts.Config,
				opts.HostConfig,
				opts.NetworkingConfig,
			},
			context: opts.Context,
		},
	)

	if e, ok := err.(*Error); ok {
		if e.Status == http.StatusNotFound {
			return nil, ErrNoSuchImage
		}
		if e.Status == http.StatusConflict {
			return nil, ErrContainerAlreadyExists
		}
	}

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var container Container
	if err := json.NewDecoder(resp.Body).Decode(&container); err != nil {
		return nil, err
	}

	container.Name = opts.Name

	return &container, nil
}

// KeyValuePair is a type for generic key/value pairs as used in the Lxc
// configuration
type KeyValuePair struct {
	Key   string `json:"Key,omitempty" yaml:"Key,omitempty"`
	Value string `json:"Value,omitempty" yaml:"Value,omitempty"`
}

// RestartPolicy represents the policy for automatically restarting a container.
//
// Possible values are:
//
//   - always: the docker daemon will always restart the container
//   - on-failure: the docker daemon will restart the container on failures, at
//                 most MaximumRetryCount times
//   - unless-stopped: the docker daemon will always restart the container except
//                 when user has manually stopped the container
//   - no: the docker daemon will not restart the container automatically
type RestartPolicy struct {
	Name              string `json:"Name,omitempty" yaml:"Name,omitempty"`
	MaximumRetryCount int    `json:"MaximumRetryCount,omitempty" yaml:"MaximumRetryCount,omitempty"`
}

// AlwaysRestart returns a restart policy that tells the Docker daemon to
// always restart the container.
func AlwaysRestart() RestartPolicy {
	return RestartPolicy{Name: "always"}
}

// RestartOnFailure returns a restart policy that tells the Docker daemon to
// restart the container on failures, trying at most maxRetry times.
func RestartOnFailure(maxRetry int) RestartPolicy {
	return RestartPolicy{Name: "on-failure", MaximumRetryCount: maxRetry}
}

// RestartUnlessStopped returns a restart policy that tells the Docker daemon to
// always restart the container except when user has manually stopped the container.
func RestartUnlessStopped() RestartPolicy {
	return RestartPolicy{Name: "unless-stopped"}
}

// NeverRestart returns a restart policy that tells the Docker daemon to never
// restart the container on failures.
func NeverRestart() RestartPolicy {
	return RestartPolicy{Name: "no"}
}

// Device represents a device mapping between the Docker host and the
// container.
type Device struct {
	PathOnHost        string `json:"PathOnHost,omitempty" yaml:"PathOnHost,omitempty"`
	PathInContainer   string `json:"PathInContainer,omitempty" yaml:"PathInContainer,omitempty"`
	CgroupPermissions string `json:"CgroupPermissions,omitempty" yaml:"CgroupPermissions,omitempty"`
}

// BlockWeight represents a relative device weight for an individual device inside
// of a container
type BlockWeight struct {
	Path   string `json:"Path,omitempty"`
	Weight string `json:"Weight,omitempty"`
}

// BlockLimit represents a read/write limit in IOPS or Bandwidth for a device
// inside of a container
type BlockLimit struct {
	Path string `json:"Path,omitempty"`
	Rate string `json:"Rate,omitempty"`
}

// HostConfig contains the container options related to starting a container on
// a given host
type HostConfig struct {
	Binds                []string               `json:"Binds,omitempty" yaml:"Binds,omitempty"`
	CapAdd               []string               `json:"CapAdd,omitempty" yaml:"CapAdd,omitempty"`
	CapDrop              []string               `json:"CapDrop,omitempty" yaml:"CapDrop,omitempty"`
	GroupAdd             []string               `json:"GroupAdd,omitempty" yaml:"GroupAdd,omitempty"`
	ContainerIDFile      string                 `json:"ContainerIDFile,omitempty" yaml:"ContainerIDFile,omitempty"`
	LxcConf              []KeyValuePair         `json:"LxcConf,omitempty" yaml:"LxcConf,omitempty"`
	PortBindings         map[Port][]PortBinding `json:"PortBindings,omitempty" yaml:"PortBindings,omitempty"`
	Links                []string               `json:"Links,omitempty" yaml:"Links,omitempty"`
	DNS                  []string               `json:"Dns,omitempty" yaml:"Dns,omitempty"` // For Docker API v1.10 and above only
	DNSOptions           []string               `json:"DnsOptions,omitempty" yaml:"DnsOptions,omitempty"`
	DNSSearch            []string               `json:"DnsSearch,omitempty" yaml:"DnsSearch,omitempty"`
	ExtraHosts           []string               `json:"ExtraHosts,omitempty" yaml:"ExtraHosts,omitempty"`
	VolumesFrom          []string               `json:"VolumesFrom,omitempty" yaml:"VolumesFrom,omitempty"`
	UsernsMode           string                 `json:"UsernsMode,omitempty" yaml:"UsernsMode,omitempty"`
	NetworkMode          string                 `json:"NetworkMode,omitempty" yaml:"NetworkMode,omitempty"`
	IpcMode              string                 `json:"IpcMode,omitempty" yaml:"IpcMode,omitempty"`
	PidMode              string                 `json:"PidMode,omitempty" yaml:"PidMode,omitempty"`
	UTSMode              string                 `json:"UTSMode,omitempty" yaml:"UTSMode,omitempty"`
	RestartPolicy        RestartPolicy          `json:"RestartPolicy,omitempty" yaml:"RestartPolicy,omitempty"`
	Devices              []Device               `json:"Devices,omitempty" yaml:"Devices,omitempty"`
	LogConfig            LogConfig              `json:"LogConfig,omitempty" yaml:"LogConfig,omitempty"`
	SecurityOpt          []string               `json:"SecurityOpt,omitempty" yaml:"SecurityOpt,omitempty"`
	Cgroup               string                 `json:"Cgroup,omitempty" yaml:"Cgroup,omitempty"`
	CgroupParent         string                 `json:"CgroupParent,omitempty" yaml:"CgroupParent,omitempty"`
	Memory               int64                  `json:"Memory,omitempty" yaml:"Memory,omitempty"`
	MemoryReservation    int64                  `json:"MemoryReservation,omitempty" yaml:"MemoryReservation,omitempty"`
	KernelMemory         int64                  `json:"KernelMemory,omitempty" yaml:"KernelMemory,omitempty"`
	MemorySwap           int64                  `json:"MemorySwap,omitempty" yaml:"MemorySwap,omitempty"`
	MemorySwappiness     int64                  `json:"MemorySwappiness,omitempty" yaml:"MemorySwappiness,omitempty"`
	CPUShares            int64                  `json:"CpuShares,omitempty" yaml:"CpuShares,omitempty"`
	CPUSet               string                 `json:"Cpuset,omitempty" yaml:"Cpuset,omitempty"`
	CPUSetCPUs           string                 `json:"CpusetCpus,omitempty" yaml:"CpusetCpus,omitempty"`
	CPUSetMEMs           string                 `json:"CpusetMems,omitempty" yaml:"CpusetMems,omitempty"`
	CPUQuota             int64                  `json:"CpuQuota,omitempty" yaml:"CpuQuota,omitempty"`
	CPUPeriod            int64                  `json:"CpuPeriod,omitempty" yaml:"CpuPeriod,omitempty"`
	BlkioWeight          int64                  `json:"BlkioWeight,omitempty" yaml:"BlkioWeight,omitempty"`
	BlkioWeightDevice    []BlockWeight          `json:"BlkioWeightDevice,omitempty" yaml:"BlkioWeightDevice,omitempty"`
	BlkioDeviceReadBps   []BlockLimit           `json:"BlkioDeviceReadBps,omitempty" yaml:"BlkioDeviceReadBps,omitempty"`
	BlkioDeviceReadIOps  []BlockLimit           `json:"BlkioDeviceReadIOps,omitempty" yaml:"BlkioDeviceReadIOps,omitempty"`
	BlkioDeviceWriteBps  []BlockLimit           `json:"BlkioDeviceWriteBps,omitempty" yaml:"BlkioDeviceWriteBps,omitempty"`
	BlkioDeviceWriteIOps []BlockLimit           `json:"BlkioDeviceWriteIOps,omitempty" yaml:"BlkioDeviceWriteIOps,omitempty"`
	Ulimits              []ULimit               `json:"Ulimits,omitempty" yaml:"Ulimits,omitempty"`
	VolumeDriver         string                 `json:"VolumeDriver,omitempty" yaml:"VolumeDriver,omitempty"`
	OomScoreAdj          int                    `json:"OomScoreAdj,omitempty" yaml:"OomScoreAdj,omitempty"`
	PidsLimit            int64                  `json:"PidsLimit,omitempty" yaml:"PidsLimit,omitempty"`
	ShmSize              int64                  `json:"ShmSize,omitempty" yaml:"ShmSize,omitempty"`
	Tmpfs                map[string]string      `json:"Tmpfs,omitempty" yaml:"Tmpfs,omitempty"`
	Privileged           bool                   `json:"Privileged,omitempty" yaml:"Privileged,omitempty"`
	PublishAllPorts      bool                   `json:"PublishAllPorts,omitempty" yaml:"PublishAllPorts,omitempty"`
	ReadonlyRootfs       bool                   `json:"ReadonlyRootfs,omitempty" yaml:"ReadonlyRootfs,omitempty"`
	OOMKillDisable       bool                   `json:"OomKillDisable,omitempty" yaml:"OomKillDisable,omitempty"`
	AutoRemove           bool                   `json:"AutoRemove,omitempty" yaml:"AutoRemove,omitempty"`
	StorageOpt           map[string]string      `json:"StorageOpt,omitempty" yaml:"StorageOpt,omitempty"`
	Sysctls              map[string]string      `json:"Sysctls,omitempty" yaml:"Sysctls,omitempty"`
}

// NetworkingConfig represents the container's networking configuration for each of its interfaces
// Carries the networking configs specified in the `docker run` and `docker network connect` commands
type NetworkingConfig struct {
	EndpointsConfig map[string]*EndpointConfig `json:"EndpointsConfig" yaml:"EndpointsConfig"` // Endpoint configs for each connecting network
}

// StartContainer starts a container, returning an error in case of failure.
//
// Passing the HostConfig to this method has been deprecated in Docker API 1.22
// (Docker Engine 1.10.x) and totally removed in Docker API 1.24 (Docker Engine
// 1.12.x). The client will ignore the parameter when communicating with Docker
// API 1.24 or greater.
//
// See https://goo.gl/fbOSZy for more details.
func (c *Client) StartContainer(id string, hostConfig *HostConfig) error {
	return c.startContainer(id, hostConfig, doOptions{})
}

// StartContainerWithContext starts a container, returning an error in case of
// failure. The context can be used to cancel the outstanding start container
// request.
//
// Passing the HostConfig to this method has been deprecated in Docker API 1.22
// (Docker Engine 1.10.x) and totally removed in Docker API 1.24 (Docker Engine
// 1.12.x). The client will ignore the parameter when communicating with Docker
// API 1.24 or greater.
//
// See https://goo.gl/fbOSZy for more details.
func (c *Client) StartContainerWithContext(id string, hostConfig *HostConfig, ctx context.Context) error {
	return c.startContainer(id, hostConfig, doOptions{context: ctx})
}

func (c *Client) startContainer(id string, hostConfig *HostConfig, opts doOptions) error {
	path := "/containers/" + id + "/start"
	if c.serverAPIVersion == nil {
		c.checkAPIVersion()
	}
	if c.serverAPIVersion != nil && c.serverAPIVersion.LessThan(apiVersion124) {
		opts.data = hostConfig
		opts.forceJSON = true
	}
	resp, err := c.do("POST", path, opts)
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return &NoSuchContainer{ID: id, Err: err}
		}
		return err
	}
	if resp.StatusCode == http.StatusNotModified {
		return &ContainerAlreadyRunning{ID: id}
	}
	resp.Body.Close()
	return nil
}

// StopContainer stops a container, killing it after the given timeout (in
// seconds).
//
// See https://goo.gl/R9dZcV for more details.
func (c *Client) StopContainer(id string, timeout uint) error {
	return c.stopContainer(id, timeout, doOptions{})
}

// StopContainerWithContext stops a container, killing it after the given
// timeout (in seconds). The context can be used to cancel the stop
// container request.
//
// See https://goo.gl/R9dZcV for more details.
func (c *Client) StopContainerWithContext(id string, timeout uint, ctx context.Context) error {
	return c.stopContainer(id, timeout, doOptions{context: ctx})
}

func (c *Client) stopContainer(id string, timeout uint, opts doOptions) error {
	path := fmt.Sprintf("/containers/%s/stop?t=%d", id, timeout)
	resp, err := c.do("POST", path, opts)
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return &NoSuchContainer{ID: id}
		}
		return err
	}
	if resp.StatusCode == http.StatusNotModified {
		return &ContainerNotRunning{ID: id}
	}
	resp.Body.Close()
	return nil
}

// RestartContainer stops a container, killing it after the given timeout (in
// seconds), during the stop process.
//
// See https://goo.gl/MrAKQ5 for more details.
func (c *Client) RestartContainer(id string, timeout uint) error {
	path := fmt.Sprintf("/containers/%s/restart?t=%d", id, timeout)
	resp, err := c.do("POST", path, doOptions{})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return &NoSuchContainer{ID: id}
		}
		return err
	}
	resp.Body.Close()
	return nil
}

// PauseContainer pauses the given container.
//
// See https://goo.gl/D1Yaii for more details.
func (c *Client) PauseContainer(id string) error {
	path := fmt.Sprintf("/containers/%s/pause", id)
	resp, err := c.do("POST", path, doOptions{})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return &NoSuchContainer{ID: id}
		}
		return err
	}
	resp.Body.Close()
	return nil
}

// UnpauseContainer unpauses the given container.
//
// See https://goo.gl/sZ2faO for more details.
func (c *Client) UnpauseContainer(id string) error {
	path := fmt.Sprintf("/containers/%s/unpause", id)
	resp, err := c.do("POST", path, doOptions{})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return &NoSuchContainer{ID: id}
		}
		return err
	}
	resp.Body.Close()
	return nil
}

// TopResult represents the list of processes running in a container, as
// returned by /containers/<id>/top.
//
// See https://goo.gl/FLwpPl for more details.
type TopResult struct {
	Titles    []string
	Processes [][]string
}

// TopContainer returns processes running inside a container
//
// See https://goo.gl/FLwpPl for more details.
func (c *Client) TopContainer(id string, psArgs string) (TopResult, error) {
	var args string
	var result TopResult
	if psArgs != "" {
		args = fmt.Sprintf("?ps_args=%s", psArgs)
	}
	path := fmt.Sprintf("/containers/%s/top%s", id, args)
	resp, err := c.do("GET", path, doOptions{})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return result, &NoSuchContainer{ID: id}
		}
		return result, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

// Stats represents container statistics, returned by /containers/<id>/stats.
//
// See https://goo.gl/Dk3Xio for more details.
type Stats struct {
	Read      time.Time `json:"read,omitempty" yaml:"read,omitempty"`
	PidsStats struct {
		Current uint64 `json:"current,omitempty" yaml:"current,omitempty"`
	} `json:"pids_stats,omitempty" yaml:"pids_stats,omitempty"`
	Network     NetworkStats            `json:"network,omitempty" yaml:"network,omitempty"`
	Networks    map[string]NetworkStats `json:"networks,omitempty" yaml:"networks,omitempty"`
	MemoryStats struct {
		Stats struct {
			TotalPgmafault          uint64 `json:"total_pgmafault,omitempty" yaml:"total_pgmafault,omitempty"`
			Cache                   uint64 `json:"cache,omitempty" yaml:"cache,omitempty"`
			MappedFile              uint64 `json:"mapped_file,omitempty" yaml:"mapped_file,omitempty"`
			TotalInactiveFile       uint64 `json:"total_inactive_file,omitempty" yaml:"total_inactive_file,omitempty"`
			Pgpgout                 uint64 `json:"pgpgout,omitempty" yaml:"pgpgout,omitempty"`
			Rss                     uint64 `json:"rss,omitempty" yaml:"rss,omitempty"`
			TotalMappedFile         uint64 `json:"total_mapped_file,omitempty" yaml:"total_mapped_file,omitempty"`
			Writeback               uint64 `json:"writeback,omitempty" yaml:"writeback,omitempty"`
			Unevictable             uint64 `json:"unevictable,omitempty" yaml:"unevictable,omitempty"`
			Pgpgin                  uint64 `json:"pgpgin,omitempty" yaml:"pgpgin,omitempty"`
			TotalUnevictable        uint64 `json:"total_unevictable,omitempty" yaml:"total_unevictable,omitempty"`
			Pgmajfault              uint64 `json:"pgmajfault,omitempty" yaml:"pgmajfault,omitempty"`
			TotalRss                uint64 `json:"total_rss,omitempty" yaml:"total_rss,omitempty"`
			TotalRssHuge            uint64 `json:"total_rss_huge,omitempty" yaml:"total_rss_huge,omitempty"`
			TotalWriteback          uint64 `json:"total_writeback,omitempty" yaml:"total_writeback,omitempty"`
			TotalInactiveAnon       uint64 `json:"total_inactive_anon,omitempty" yaml:"total_inactive_anon,omitempty"`
			RssHuge                 uint64 `json:"rss_huge,omitempty" yaml:"rss_huge,omitempty"`
			HierarchicalMemoryLimit uint64 `json:"hierarchical_memory_limit,omitempty" yaml:"hierarchical_memory_limit,omitempty"`
			TotalPgfault            uint64 `json:"total_pgfault,omitempty" yaml:"total_pgfault,omitempty"`
			TotalActiveFile         uint64 `json:"total_active_file,omitempty" yaml:"total_active_file,omitempty"`
			ActiveAnon              uint64 `json:"active_anon,omitempty" yaml:"active_anon,omitempty"`
			TotalActiveAnon         uint64 `json:"total_active_anon,omitempty" yaml:"total_active_anon,omitempty"`
			TotalPgpgout            uint64 `json:"total_pgpgout,omitempty" yaml:"total_pgpgout,omitempty"`
			TotalCache              uint64 `json:"total_cache,omitempty" yaml:"total_cache,omitempty"`
			InactiveAnon            uint64 `json:"inactive_anon,omitempty" yaml:"inactive_anon,omitempty"`
			ActiveFile              uint64 `json:"active_file,omitempty" yaml:"active_file,omitempty"`
			Pgfault                 uint64 `json:"pgfault,omitempty" yaml:"pgfault,omitempty"`
			InactiveFile            uint64 `json:"inactive_file,omitempty" yaml:"inactive_file,omitempty"`
			TotalPgpgin             uint64 `json:"total_pgpgin,omitempty" yaml:"total_pgpgin,omitempty"`
			HierarchicalMemswLimit  uint64 `json:"hierarchical_memsw_limit,omitempty" yaml:"hierarchical_memsw_limit,omitempty"`
			Swap                    uint64 `json:"swap,omitempty" yaml:"swap,omitempty"`
		} `json:"stats,omitempty" yaml:"stats,omitempty"`
		MaxUsage uint64 `json:"max_usage,omitempty" yaml:"max_usage,omitempty"`
		Usage    uint64 `json:"usage,omitempty" yaml:"usage,omitempty"`
		Failcnt  uint64 `json:"failcnt,omitempty" yaml:"failcnt,omitempty"`
		Limit    uint64 `json:"limit,omitempty" yaml:"limit,omitempty"`
	} `json:"memory_stats,omitempty" yaml:"memory_stats,omitempty"`
	BlkioStats struct {
		IOServiceBytesRecursive []BlkioStatsEntry `json:"io_service_bytes_recursive,omitempty" yaml:"io_service_bytes_recursive,omitempty"`
		IOServicedRecursive     []BlkioStatsEntry `json:"io_serviced_recursive,omitempty" yaml:"io_serviced_recursive,omitempty"`
		IOQueueRecursive        []BlkioStatsEntry `json:"io_queue_recursive,omitempty" yaml:"io_queue_recursive,omitempty"`
		IOServiceTimeRecursive  []BlkioStatsEntry `json:"io_service_time_recursive,omitempty" yaml:"io_service_time_recursive,omitempty"`
		IOWaitTimeRecursive     []BlkioStatsEntry `json:"io_wait_time_recursive,omitempty" yaml:"io_wait_time_recursive,omitempty"`
		IOMergedRecursive       []BlkioStatsEntry `json:"io_merged_recursive,omitempty" yaml:"io_merged_recursive,omitempty"`
		IOTimeRecursive         []BlkioStatsEntry `json:"io_time_recursive,omitempty" yaml:"io_time_recursive,omitempty"`
		SectorsRecursive        []BlkioStatsEntry `json:"sectors_recursive,omitempty" yaml:"sectors_recursive,omitempty"`
	} `json:"blkio_stats,omitempty" yaml:"blkio_stats,omitempty"`
	CPUStats    CPUStats `json:"cpu_stats,omitempty" yaml:"cpu_stats,omitempty"`
	PreCPUStats CPUStats `json:"precpu_stats,omitempty"`
}

// NetworkStats is a stats entry for network stats
type NetworkStats struct {
	RxDropped uint64 `json:"rx_dropped,omitempty" yaml:"rx_dropped,omitempty"`
	RxBytes   uint64 `json:"rx_bytes,omitempty" yaml:"rx_bytes,omitempty"`
	RxErrors  uint64 `json:"rx_errors,omitempty" yaml:"rx_errors,omitempty"`
	TxPackets uint64 `json:"tx_packets,omitempty" yaml:"tx_packets,omitempty"`
	TxDropped uint64 `json:"tx_dropped,omitempty" yaml:"tx_dropped,omitempty"`
	RxPackets uint64 `json:"rx_packets,omitempty" yaml:"rx_packets,omitempty"`
	TxErrors  uint64 `json:"tx_errors,omitempty" yaml:"tx_errors,omitempty"`
	TxBytes   uint64 `json:"tx_bytes,omitempty" yaml:"tx_bytes,omitempty"`
}

// CPUStats is a stats entry for cpu stats
type CPUStats struct {
	CPUUsage struct {
		PercpuUsage       []uint64 `json:"percpu_usage,omitempty" yaml:"percpu_usage,omitempty"`
		UsageInUsermode   uint64   `json:"usage_in_usermode,omitempty" yaml:"usage_in_usermode,omitempty"`
		TotalUsage        uint64   `json:"total_usage,omitempty" yaml:"total_usage,omitempty"`
		UsageInKernelmode uint64   `json:"usage_in_kernelmode,omitempty" yaml:"usage_in_kernelmode,omitempty"`
	} `json:"cpu_usage,omitempty" yaml:"cpu_usage,omitempty"`
	SystemCPUUsage uint64 `json:"system_cpu_usage,omitempty" yaml:"system_cpu_usage,omitempty"`
	ThrottlingData struct {
		Periods          uint64 `json:"periods,omitempty"`
		ThrottledPeriods uint64 `json:"throttled_periods,omitempty"`
		ThrottledTime    uint64 `json:"throttled_time,omitempty"`
	} `json:"throttling_data,omitempty" yaml:"throttling_data,omitempty"`
}

// BlkioStatsEntry is a stats entry for blkio_stats
type BlkioStatsEntry struct {
	Major uint64 `json:"major,omitempty" yaml:"major,omitempty"`
	Minor uint64 `json:"minor,omitempty" yaml:"minor,omitempty"`
	Op    string `json:"op,omitempty" yaml:"op,omitempty"`
	Value uint64 `json:"value,omitempty" yaml:"value,omitempty"`
}

// StatsOptions specify parameters to the Stats function.
//
// See https://goo.gl/Dk3Xio for more details.
type StatsOptions struct {
	ID     string
	Stats  chan<- *Stats
	Stream bool
	// A flag that enables stopping the stats operation
	Done <-chan bool
	// Initial connection timeout
	Timeout time.Duration
	// Timeout with no data is received, it's reset every time new data
	// arrives
	InactivityTimeout time.Duration `qs:"-"`
	Context           context.Context
}

// Stats sends container statistics for the given container to the given channel.
//
// This function is blocking, similar to a streaming call for logs, and should be run
// on a separate goroutine from the caller. Note that this function will block until
// the given container is removed, not just exited. When finished, this function
// will close the given channel. Alternatively, function can be stopped by
// signaling on the Done channel.
//
// See https://goo.gl/Dk3Xio for more details.
func (c *Client) Stats(opts StatsOptions) (retErr error) {
	errC := make(chan error, 1)
	readCloser, writeCloser := io.Pipe()

	defer func() {
		close(opts.Stats)

		select {
		case err := <-errC:
			if err != nil && retErr == nil {
				retErr = err
			}
		default:
			// No errors
		}

		if err := readCloser.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	go func() {
		err := c.stream("GET", fmt.Sprintf("/containers/%s/stats?stream=%v", opts.ID, opts.Stream), streamOptions{
			rawJSONStream:     true,
			useJSONDecoder:    true,
			stdout:            writeCloser,
			timeout:           opts.Timeout,
			inactivityTimeout: opts.InactivityTimeout,
			context:           opts.Context,
		})
		if err != nil {
			dockerError, ok := err.(*Error)
			if ok {
				if dockerError.Status == http.StatusNotFound {
					err = &NoSuchContainer{ID: opts.ID}
				}
			}
		}
		if closeErr := writeCloser.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		errC <- err
		close(errC)
	}()

	quit := make(chan struct{})
	defer close(quit)
	go func() {
		// block here waiting for the signal to stop function
		select {
		case <-opts.Done:
			readCloser.Close()
		case <-quit:
			return
		}
	}()

	decoder := json.NewDecoder(readCloser)
	stats := new(Stats)
	for err := decoder.Decode(stats); err != io.EOF; err = decoder.Decode(stats) {
		if err != nil {
			return err
		}
		opts.Stats <- stats
		stats = new(Stats)
	}
	return nil
}

// KillContainerOptions represents the set of options that can be used in a
// call to KillContainer.
//
// See https://goo.gl/JnTxXZ for more details.
type KillContainerOptions struct {
	// The ID of the container.
	ID string `qs:"-"`

	// The signal to send to the container. When omitted, Docker server
	// will assume SIGKILL.
	Signal  Signal
	Context context.Context
}

// KillContainer sends a signal to a container, returning an error in case of
// failure.
//
// See https://goo.gl/JnTxXZ for more details.
func (c *Client) KillContainer(opts KillContainerOptions) error {
	path := "/containers/" + opts.ID + "/kill" + "?" + queryString(opts)
	resp, err := c.do("POST", path, doOptions{context: opts.Context})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return &NoSuchContainer{ID: opts.ID}
		}
		return err
	}
	resp.Body.Close()
	return nil
}

// RemoveContainerOptions encapsulates options to remove a container.
//
// See https://goo.gl/hL5IPC for more details.
type RemoveContainerOptions struct {
	// The ID of the container.
	ID string `qs:"-"`

	// A flag that indicates whether Docker should remove the volumes
	// associated to the container.
	RemoveVolumes bool `qs:"v"`

	// A flag that indicates whether Docker should remove the container
	// even if it is currently running.
	Force   bool
	Context context.Context
}

// RemoveContainer removes a container, returning an error in case of failure.
//
// See https://goo.gl/hL5IPC for more details.
func (c *Client) RemoveContainer(opts RemoveContainerOptions) error {
	path := "/containers/" + opts.ID + "?" + queryString(opts)
	resp, err := c.do("DELETE", path, doOptions{context: opts.Context})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return &NoSuchContainer{ID: opts.ID}
		}
		return err
	}
	resp.Body.Close()
	return nil
}

// UploadToContainerOptions is the set of options that can be used when
// uploading an archive into a container.
//
// See https://goo.gl/g25o7u for more details.
type UploadToContainerOptions struct {
	InputStream          io.Reader `json:"-" qs:"-"`
	Path                 string    `qs:"path"`
	NoOverwriteDirNonDir bool      `qs:"noOverwriteDirNonDir"`
	Context              context.Context
}

// UploadToContainer uploads a tar archive to be extracted to a path in the
// filesystem of the container.
//
// See https://goo.gl/g25o7u for more details.
func (c *Client) UploadToContainer(id string, opts UploadToContainerOptions) error {
	url := fmt.Sprintf("/containers/%s/archive?", id) + queryString(opts)

	return c.stream("PUT", url, streamOptions{
		in:      opts.InputStream,
		context: opts.Context,
	})
}

// DownloadFromContainerOptions is the set of options that can be used when
// downloading resources from a container.
//
// See https://goo.gl/W49jxK for more details.
type DownloadFromContainerOptions struct {
	OutputStream      io.Writer     `json:"-" qs:"-"`
	Path              string        `qs:"path"`
	InactivityTimeout time.Duration `qs:"-"`
	Context           context.Context
}

// DownloadFromContainer downloads a tar archive of files or folders in a container.
//
// See https://goo.gl/W49jxK for more details.
func (c *Client) DownloadFromContainer(id string, opts DownloadFromContainerOptions) error {
	url := fmt.Sprintf("/containers/%s/archive?", id) + queryString(opts)

	return c.stream("GET", url, streamOptions{
		setRawTerminal:    true,
		stdout:            opts.OutputStream,
		inactivityTimeout: opts.InactivityTimeout,
		context:           opts.Context,
	})
}

// CopyFromContainerOptions has been DEPRECATED, please use DownloadFromContainerOptions along with DownloadFromContainer.
//
// See https://goo.gl/nWk2YQ for more details.
type CopyFromContainerOptions struct {
	OutputStream io.Writer `json:"-"`
	Container    string    `json:"-"`
	Resource     string
	Context      context.Context `json:"-"`
}

// CopyFromContainer has been DEPRECATED, please use DownloadFromContainerOptions along with DownloadFromContainer.
//
// See https://goo.gl/nWk2YQ for more details.
func (c *Client) CopyFromContainer(opts CopyFromContainerOptions) error {
	if opts.Container == "" {
		return &NoSuchContainer{ID: opts.Container}
	}
	if c.serverAPIVersion == nil {
		c.checkAPIVersion()
	}
	if c.serverAPIVersion != nil && c.serverAPIVersion.GreaterThanOrEqualTo(apiVersion124) {
		return errors.New("go-dockerclient: CopyFromContainer is no longer available in Docker >= 1.12, use DownloadFromContainer instead")
	}
	url := fmt.Sprintf("/containers/%s/copy", opts.Container)
	resp, err := c.do("POST", url, doOptions{
		data:    opts,
		context: opts.Context,
	})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return &NoSuchContainer{ID: opts.Container}
		}
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(opts.OutputStream, resp.Body)
	return err
}

// WaitContainer blocks until the given container stops, return the exit code
// of the container status.
//
// See https://goo.gl/4AGweZ for more details.
func (c *Client) WaitContainer(id string) (int, error) {
	resp, err := c.do("POST", "/containers/"+id+"/wait", doOptions{})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return 0, &NoSuchContainer{ID: id}
		}
		return 0, err
	}
	defer resp.Body.Close()
	var r struct{ StatusCode int }
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return 0, err
	}
	return r.StatusCode, nil
}

// CommitContainerOptions aggregates parameters to the CommitContainer method.
//
// See https://goo.gl/CzIguf for more details.
type CommitContainerOptions struct {
	Container  string
	Repository string `qs:"repo"`
	Tag        string
	Message    string `qs:"comment"`
	Author     string
	Run        *Config `qs:"-"`
	Context    context.Context
}

// CommitContainer creates a new image from a container's changes.
//
// See https://goo.gl/CzIguf for more details.
func (c *Client) CommitContainer(opts CommitContainerOptions) (*Image, error) {
	path := "/commit?" + queryString(opts)
	resp, err := c.do("POST", path, doOptions{
		data:    opts.Run,
		context: opts.Context,
	})
	if err != nil {
		if e, ok := err.(*Error); ok && e.Status == http.StatusNotFound {
			return nil, &NoSuchContainer{ID: opts.Container}
		}
		return nil, err
	}
	defer resp.Body.Close()
	var image Image
	if err := json.NewDecoder(resp.Body).Decode(&image); err != nil {
		return nil, err
	}
	return &image, nil
}

// AttachToContainerOptions is the set of options that can be used when
// attaching to a container.
//
// See https://goo.gl/JF10Zk for more details.
type AttachToContainerOptions struct {
	Container    string    `qs:"-"`
	InputStream  io.Reader `qs:"-"`
	OutputStream io.Writer `qs:"-"`
	ErrorStream  io.Writer `qs:"-"`

	// If set, after a successful connect, a sentinel will be sent and then the
	// client will block on receive before continuing.
	//
	// It must be an unbuffered channel. Using a buffered channel can lead
	// to unexpected behavior.
	Success chan struct{}

	// Use raw terminal? Usually true when the container contains a TTY.
	RawTerminal bool `qs:"-"`

	// Get container logs, sending it to OutputStream.
	Logs bool

	// Stream the response?
	Stream bool

	// Attach to stdin, and use InputStream.
	Stdin bool

	// Attach to stdout, and use OutputStream.
	Stdout bool

	// Attach to stderr, and use ErrorStream.
	Stderr bool
}

// AttachToContainer attaches to a container, using the given options.
//
// See https://goo.gl/JF10Zk for more details.
func (c *Client) AttachToContainer(opts AttachToContainerOptions) error {
	cw, err := c.AttachToContainerNonBlocking(opts)
	if err != nil {
		return err
	}
	return cw.Wait()
}

// AttachToContainerNonBlocking attaches to a container, using the given options.
// This function does not block.
//
// See https://goo.gl/NKpkFk for more details.
func (c *Client) AttachToContainerNonBlocking(opts AttachToContainerOptions) (CloseWaiter, error) {
	if opts.Container == "" {
		return nil, &NoSuchContainer{ID: opts.Container}
	}
	path := "/containers/" + opts.Container + "/attach?" + queryString(opts)
	return c.hijack("POST", path, hijackOptions{
		success:        opts.Success,
		setRawTerminal: opts.RawTerminal,
		in:             opts.InputStream,
		stdout:         opts.OutputStream,
		stderr:         opts.ErrorStream,
	})
}

// LogsOptions represents the set of options used when getting logs from a
// container.
//
// See https://goo.gl/krK0ZH for more details.
type LogsOptions struct {
	Context           context.Context
	Container         string        `qs:"-"`
	OutputStream      io.Writer     `qs:"-"`
	ErrorStream       io.Writer     `qs:"-"`
	InactivityTimeout time.Duration `qs:"-"`
	Tail              string

	Since      int64
	Follow     bool
	Stdout     bool
	Stderr     bool
	Timestamps bool

	// Use raw terminal? Usually true when the container contains a TTY.
	RawTerminal bool `qs:"-"`
}

// Logs gets stdout and stderr logs from the specified container.
//
// See https://goo.gl/krK0ZH for more details.
func (c *Client) Logs(opts LogsOptions) error {
	if opts.Container == "" {
		return &NoSuchContainer{ID: opts.Container}
	}
	if opts.Tail == "" {
		opts.Tail = "all"
	}
	path := "/containers/" + opts.Container + "/logs?" + queryString(opts)
	return c.stream("GET", path, streamOptions{
		setRawTerminal:    opts.RawTerminal,
		stdout:            opts.OutputStream,
		stderr:            opts.ErrorStream,
		inactivityTimeout: opts.InactivityTimeout,
		context:           opts.Context,
	})
}

// ResizeContainerTTY resizes the terminal to the given height and width.
//
// See https://goo.gl/FImjeq for more details.
func (c *Client) ResizeContainerTTY(id string, height, width int) error {
	params := make(url.Values)
	params.Set("h", strconv.Itoa(height))
	params.Set("w", strconv.Itoa(width))
	resp, err := c.do("POST", "/containers/"+id+"/resize?"+params.Encode(), doOptions{})
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// ExportContainerOptions is the set of parameters to the ExportContainer
// method.
//
// See https://goo.gl/yGJCIh for more details.
type ExportContainerOptions struct {
	ID                string
	OutputStream      io.Writer
	InactivityTimeout time.Duration `qs:"-"`
	Context           context.Context
}

// ExportContainer export the contents of container id as tar archive
// and prints the exported contents to stdout.
//
// See https://goo.gl/yGJCIh for more details.
func (c *Client) ExportContainer(opts ExportContainerOptions) error {
	if opts.ID == "" {
		return &NoSuchContainer{ID: opts.ID}
	}
	url := fmt.Sprintf("/containers/%s/export", opts.ID)
	return c.stream("GET", url, streamOptions{
		setRawTerminal:    true,
		stdout:            opts.OutputStream,
		inactivityTimeout: opts.InactivityTimeout,
		context:           opts.Context,
	})
}

// NoSuchContainer is the error returned when a given container does not exist.
type NoSuchContainer struct {
	ID  string
	Err error
}

func (err *NoSuchContainer) Error() string {
	if err.Err != nil {
		return err.Err.Error()
	}
	return "No such container: " + err.ID
}

// ContainerAlreadyRunning is the error returned when a given container is
// already running.
type ContainerAlreadyRunning struct {
	ID string
}

func (err *ContainerAlreadyRunning) Error() string {
	return "Container already running: " + err.ID
}

// ContainerNotRunning is the error returned when a given container is not
// running.
type ContainerNotRunning struct {
	ID string
}

func (err *ContainerNotRunning) Error() string {
	return "Container not running: " + err.ID
}
