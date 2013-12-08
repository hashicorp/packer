// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.

package googlecompute

import (
	"errors"
	"net/http"
	"strings"

	"code.google.com/p/goauth2/oauth"
	"code.google.com/p/goauth2/oauth/jwt"
	"code.google.com/p/google-api-go-client/compute/v1beta16"
)

// GoogleComputeClient represents a GCE client.
type GoogleComputeClient struct {
	ProjectId     string
	Service       *compute.Service
	Zone          string
	clientSecrets *clientSecrets
}

// InstanceConfig represents a GCE instance configuration.
// Used for creating machine instances.
type InstanceConfig struct {
	Description       string
	Image             string
	MachineType       string
	Metadata          *compute.Metadata
	Name              string
	NetworkInterfaces []*compute.NetworkInterface
	ServiceAccounts   []*compute.ServiceAccount
	Tags              *compute.Tags
}

// New initializes and returns a *GoogleComputeClient.
//
// The projectId must be the project name, i.e. myproject, not the project
// number.
func New(projectId string, zone string, c *clientSecrets, pemKey []byte) (*GoogleComputeClient, error) {
	googleComputeClient := &GoogleComputeClient{
		ProjectId: projectId,
		Zone:      zone,
	}
	// Get the access token.
	t := jwt.NewToken(c.Web.ClientEmail, scopes(), pemKey)
	t.ClaimSet.Aud = c.Web.TokenURI
	httpClient := &http.Client{}
	token, err := t.Assert(httpClient)
	if err != nil {
		return nil, err
	}
	config := &oauth.Config{
		ClientId: c.Web.ClientId,
		Scope:    scopes(),
		TokenURL: c.Web.TokenURI,
		AuthURL:  c.Web.AuthURI,
	}
	transport := &oauth.Transport{Config: config}
	transport.Token = token
	s, err := compute.New(transport.Client())
	if err != nil {
		return nil, err
	}
	googleComputeClient.Service = s
	return googleComputeClient, nil
}

// GetZone returns a *compute.Zone representing the named zone.
func (g *GoogleComputeClient) GetZone(name string) (*compute.Zone, error) {
	zoneGetCall := g.Service.Zones.Get(g.ProjectId, name)
	zone, err := zoneGetCall.Do()
	if err != nil {
		return nil, err
	}
	return zone, nil
}

// GetMachineType returns a *compute.MachineType representing the named machine type.
func (g *GoogleComputeClient) GetMachineType(name, zone string) (*compute.MachineType, error) {
	machineTypesGetCall := g.Service.MachineTypes.Get(g.ProjectId, zone, name)
	machineType, err := machineTypesGetCall.Do()
	if err != nil {
		return nil, err
	}
	if machineType.Deprecated == nil {
		return machineType, nil
	}
	return nil, errors.New("Machine Type does not exist: " + name)
}

// GetImage returns a *compute.Image representing the named image.
func (g *GoogleComputeClient) GetImage(name string) (*compute.Image, error) {
	var err error
	var image *compute.Image
	projects := []string{g.ProjectId, "debian-cloud", "centos-cloud"}
	for _, project := range projects {
		imagesGetCall := g.Service.Images.Get(project, name)
		image, err = imagesGetCall.Do()
		if image != nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	if image != nil {
		if image.SelfLink != "" {
			return image, nil
		}
	}
	return nil, errors.New("Image does not exist: " + name)
}

// GetNetwork returns a *compute.Network representing the named network.
func (g *GoogleComputeClient) GetNetwork(name string) (*compute.Network, error) {
	networkGetCall := g.Service.Networks.Get(g.ProjectId, name)
	network, err := networkGetCall.Do()
	if err != nil {
		return nil, err
	}
	return network, nil
}

// CreateInstance creates an instance in Google Compute Engine based on the
// supplied instanceConfig.
func (g *GoogleComputeClient) CreateInstance(zone string, instanceConfig *InstanceConfig) (*compute.Operation, error) {
	instance := &compute.Instance{
		Description:       instanceConfig.Description,
		Image:             instanceConfig.Image,
		MachineType:       instanceConfig.MachineType,
		Metadata:          instanceConfig.Metadata,
		Name:              instanceConfig.Name,
		NetworkInterfaces: instanceConfig.NetworkInterfaces,
		ServiceAccounts:   instanceConfig.ServiceAccounts,
		Tags:              instanceConfig.Tags,
	}
	instanceInsertCall := g.Service.Instances.Insert(g.ProjectId, zone, instance)
	operation, err := instanceInsertCall.Do()
	if err != nil {
		return nil, err
	}
	return operation, nil
}

// InstanceStatus returns a string representing the status of the named instance.
// Status will be one of: "PROVISIONING", "STAGING", "RUNNING", "STOPPING",
// "STOPPED", "TERMINATED".
func (g *GoogleComputeClient) InstanceStatus(zone, name string) (string, error) {
	instanceGetCall := g.Service.Instances.Get(g.ProjectId, zone, name)
	instance, err := instanceGetCall.Do()
	if err != nil {
		return "", err
	}
	return instance.Status, nil
}

// CreateImage registers a GCE Image with a project.
func (g *GoogleComputeClient) CreateImage(name, description, sourceURL string) (*compute.Operation, error) {
	imageRawDisk := &compute.ImageRawDisk{
		ContainerType: "TAR",
		Source:        sourceURL,
	}
	image := &compute.Image{
		Description: description,
		Name:        name,
		RawDisk:     imageRawDisk,
		SourceType:  "RAW",
	}
	imageInsertCall := g.Service.Images.Insert(g.ProjectId, image)
	operation, err := imageInsertCall.Do()
	if err != nil {
		return nil, err
	}
	return operation, nil
}

// GetNatIp returns the public IPv4 address for named GCE instance.
func (g *GoogleComputeClient) GetNatIP(zone, name string) (string, error) {
	instanceGetCall := g.Service.Instances.Get(g.ProjectId, zone, name)
	instance, err := instanceGetCall.Do()
	if err != nil {
		return "", err
	}
	for _, ni := range instance.NetworkInterfaces {
		if ni.AccessConfigs == nil {
			continue
		}
		for _, ac := range ni.AccessConfigs {
			if ac.NatIP != "" {
				return ac.NatIP, nil
			}
		}
	}
	return "", nil
}

// ZoneOperationStatus returns the status for the named zone operation.
func (g *GoogleComputeClient) ZoneOperationStatus(zone, name string) (string, error) {
	zoneOperationsGetCall := g.Service.ZoneOperations.Get(g.ProjectId, zone, name)
	operation, err := zoneOperationsGetCall.Do()
	if err != nil {
		return "", err
	}
	if operation.Status == "DONE" {
		err = processOperationStatus(operation)
		if err != nil {
			return operation.Status, err
		}
	}
	return operation.Status, nil
}

// GlobalOperationStatus returns the status for the named global operation.
func (g *GoogleComputeClient) GlobalOperationStatus(name string) (string, error) {
	globalOperationsGetCall := g.Service.GlobalOperations.Get(g.ProjectId, name)
	operation, err := globalOperationsGetCall.Do()
	if err != nil {
		return "", err
	}
	if operation.Status == "DONE" {
		err = processOperationStatus(operation)
		if err != nil {
			return operation.Status, err
		}
	}
	return operation.Status, nil
}

// processOperationStatus extracts errors from the specified operation.
func processOperationStatus(o *compute.Operation) error {
	if o.Error != nil {
		messages := make([]string, len(o.Error.Errors))
		for _, e := range o.Error.Errors {
			messages = append(messages, e.Message)
		}
		return errors.New(strings.Join(messages, "\n"))
	}
	return nil
}

// DeleteImage deletes the named image. Returns a Global Operation.
func (g *GoogleComputeClient) DeleteImage(name string) (*compute.Operation, error) {
	imagesDeleteCall := g.Service.Images.Delete(g.ProjectId, name)
	operation, err := imagesDeleteCall.Do()
	if err != nil {
		return nil, err
	}
	return operation, nil
}

// DeleteInstance deletes the named instance. Returns a Zone Operation.
func (g *GoogleComputeClient) DeleteInstance(zone, name string) (*compute.Operation, error) {
	instanceDeleteCall := g.Service.Instances.Delete(g.ProjectId, zone, name)
	operation, err := instanceDeleteCall.Do()
	if err != nil {
		return nil, err
	}
	return operation, nil
}

// NewNetworkInterface returns a *compute.NetworkInterface based on the data provided.
func NewNetworkInterface(network *compute.Network, public bool) *compute.NetworkInterface {
	accessConfigs := make([]*compute.AccessConfig, 0)
	if public {
		c := &compute.AccessConfig{
			Name: "AccessConfig created by Packer",
			Type: "ONE_TO_ONE_NAT",
		}
		accessConfigs = append(accessConfigs, c)
	}
	return &compute.NetworkInterface{
		AccessConfigs: accessConfigs,
		Network:       network.SelfLink,
	}
}

// NewServiceAccount returns a *compute.ServiceAccount with permissions required
// for creating GCE machine images.
func NewServiceAccount(email string) *compute.ServiceAccount {
	return &compute.ServiceAccount{
		Email: email,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/compute",
			"https://www.googleapis.com/auth/devstorage.full_control",
		},
	}
}

// MapToMetadata converts a map[string]string to a *compute.Metadata.
func MapToMetadata(metadata map[string]string) *compute.Metadata {
	items := make([]*compute.MetadataItems, len(metadata))
	for k, v := range metadata {
		items = append(items, &compute.MetadataItems{k, v})
	}
	return &compute.Metadata{
		Items: items,
	}
}

// SliceToTags converts a []string to a *compute.Tags.
func SliceToTags(tags []string) *compute.Tags {
	return &compute.Tags{
		Items: tags,
	}
}

// scopes return a space separated list of scopes.
func scopes() string {
	s := []string{
		"https://www.googleapis.com/auth/compute",
		"https://www.googleapis.com/auth/devstorage.full_control",
	}
	return strings.Join(s, " ")
}
