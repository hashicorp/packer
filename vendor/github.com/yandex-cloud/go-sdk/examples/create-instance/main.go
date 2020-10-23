// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Maxim Kolganov <manykey@yandex-team.ru>

package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

func main() {
	token := flag.String("token", "", "")
	folderID := flag.String("folder-id", "", "Your Yandex.Cloud folder id")
	zone := flag.String("zone", "ru-central1-b", "Compute Engine zone to deploy to.")
	name := flag.String("name", "demo-instance", "New instance name.")
	subnetID := flag.String("subnet-id", "", "Subnet of the instance")
	flag.Parse()

	ctx := context.Background()

	sdk, err := ycsdk.Build(ctx, ycsdk.Config{
		Credentials: ycsdk.OAuthToken(*token),
	})
	if err != nil {
		log.Fatal(err)
	}
	op, err := sdk.WrapOperation(createInstance(
		ctx, sdk, *folderID, *zone, *name, *subnetID))
	if err != nil {
		log.Fatal(err)
	}
	meta, err := op.Metadata()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Creating instance %s\n",
		meta.(*compute.CreateInstanceMetadata).InstanceId)
	err = op.Wait(ctx)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := op.Response()
	if err != nil {
		log.Fatal(err)
	}
	instance := resp.(*compute.Instance)
	fmt.Printf("Deleting instance %s\n", instance.Id)
	op, err = sdk.WrapOperation(deleteInstance(ctx, sdk, instance.Id))
	if err != nil {
		log.Fatal(err)
	}
	err = op.Wait(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func createInstance(ctx context.Context, sdk *ycsdk.SDK, folderID, zone, name, subnetID string) (*operation.Operation, error) {
	if subnetID == "" {
		subnetID = findSubnet(ctx, sdk, folderID, zone)
	}
	sourceImageID := sourceImage(ctx, sdk)
	request := &compute.CreateInstanceRequest{
		FolderId:   folderID,
		Name:       name,
		ZoneId:     zone,
		PlatformId: "standard-v1",
		ResourcesSpec: &compute.ResourcesSpec{
			Cores:  1,
			Memory: 2 * 1024 * 1024 * 1024,
		},
		BootDiskSpec: &compute.AttachedDiskSpec{
			AutoDelete: true,
			Disk: &compute.AttachedDiskSpec_DiskSpec_{
				DiskSpec: &compute.AttachedDiskSpec_DiskSpec{
					TypeId: "network-hdd",
					Size:   20 * 1024 * 1024 * 1024,
					Source: &compute.AttachedDiskSpec_DiskSpec_ImageId{
						ImageId: sourceImageID,
					},
				},
			},
		},
		NetworkInterfaceSpecs: []*compute.NetworkInterfaceSpec{
			{
				SubnetId: subnetID,
				PrimaryV4AddressSpec: &compute.PrimaryAddressSpec{
					OneToOneNatSpec: &compute.OneToOneNatSpec{
						IpVersion: compute.IpVersion_IPV4,
					},
				},
			},
		},
	}
	op, err := sdk.Compute().Instance().Create(ctx, request)
	return op, err
}

func sourceImage(ctx context.Context, sdk *ycsdk.SDK) string {
	image, err := sdk.Compute().Image().GetLatestByFamily(ctx, &compute.GetImageLatestByFamilyRequest{
		FolderId: "standard-images",
		Family:   "debian-9",
	})
	if err != nil {
		log.Fatal(err)
	}
	return image.Id
}

func deleteInstance(ctx context.Context, sdk *ycsdk.SDK, id string) (*operation.Operation, error) {
	return sdk.Compute().Instance().Delete(ctx, &compute.DeleteInstanceRequest{
		InstanceId: id,
	})
}

func findSubnet(ctx context.Context, sdk *ycsdk.SDK, folderID string, zone string) string {
	resp, err := sdk.VPC().Subnet().List(ctx, &vpc.ListSubnetsRequest{
		FolderId: folderID,
		PageSize: 100,
	})
	if err != nil {
		log.Fatal(err)
	}
	subnetID := ""
	for _, subnet := range resp.Subnets {
		if subnet.ZoneId != zone {
			continue
		}
		subnetID = subnet.Id
		break
	}
	if subnetID == "" {
		log.Fatal(fmt.Sprintf("no subnets in zone: %s", zone))
	}
	return subnetID
}
