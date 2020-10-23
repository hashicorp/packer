package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/c2h5oh/datasize"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

func main() {
	flags := parseCmd()
	ctx := context.Background()

	sdk, err := ycsdk.Build(ctx, ycsdk.Config{
		Credentials: ycsdk.OAuthToken(*flags.token),
	})

	if err != nil {
		log.Fatal(err)
	}

	fillMissingFlags(ctx, sdk, flags)

	cluster := createCluster(ctx, sdk, flags)
	defer deleteCluster(ctx, sdk, cluster)
	changeClusterDescription(ctx, sdk, cluster)
	addClusterHost(ctx, sdk, cluster, flags)
}

type cmdFlags struct {
	token        *string
	folderID     *string
	zoneID       *string
	networkID    *string
	subnetID     *string
	clusterName  *string
	clusterDesc  *string
	dbName       *string
	userName     *string
	userPassword *string
}

func parseCmd() (ret *cmdFlags) {
	ret = &cmdFlags{}
	ret.token = flag.String("token", "", "")
	ret.folderID = flag.String("folder-id", "", "Your Yandex.Cloud folder id")
	ret.zoneID = flag.String("zone", "ru-central1-b", "Compute Engine zone to deploy to.")
	ret.networkID = flag.String("network-id", "", "Your Yandex.Cloud network id")
	ret.subnetID = flag.String("subnet-id", "", "Subnet of the instance")
	ret.clusterName = flag.String("cluster-name", "mongodb666", "")
	ret.clusterDesc = flag.String("cluster-desc", "", "")
	ret.dbName = flag.String("db-name", "db1", "")
	ret.userName = flag.String("user-name", "user1", "")
	ret.userPassword = flag.String("user-password", "password123", "")

	flag.Parse()
	return
}

func fillMissingFlags(ctx context.Context, sdk *ycsdk.SDK, flags *cmdFlags) {
	if *flags.networkID == "" {
		flags.networkID = findNetwork(ctx, sdk, *flags.folderID)
	}

	if *flags.subnetID == "" {
		flags.subnetID = findSubnet(ctx, sdk, *flags.folderID, *flags.networkID, *flags.zoneID)
	}
}

func findNetwork(ctx context.Context, sdk *ycsdk.SDK, folderID string) *string {
	resp, err := sdk.VPC().Network().List(ctx, &vpc.ListNetworksRequest{
		FolderId: folderID,
		PageSize: 100,
	})
	if err != nil {
		log.Fatal(err)
	}
	networkID := ""
	for _, network := range resp.Networks {
		if network.FolderId != folderID {
			continue
		}
		networkID = network.Id
		break
	}
	if networkID == "" {
		log.Fatal(fmt.Sprintf("no networks in folder: %s", folderID))
	}
	return &networkID
}

func findSubnet(ctx context.Context, sdk *ycsdk.SDK, folderID string, networkID string, zone string) *string {
	resp, err := sdk.VPC().Subnet().List(ctx, &vpc.ListSubnetsRequest{
		FolderId: folderID,
		PageSize: 100,
	})
	if err != nil {
		log.Fatal(err)
	}
	subnetID := ""
	for _, subnet := range resp.Subnets {
		if subnet.ZoneId != zone || subnet.NetworkId != networkID {
			continue
		}
		subnetID = subnet.Id
		break
	}
	if subnetID == "" {
		log.Fatal(fmt.Sprintf("no subnets in zone: %s", zone))
	}
	return &subnetID
}

func createCluster(ctx context.Context, sdk *ycsdk.SDK, flags *cmdFlags) *postgresql.Cluster {
	req := createClusterRequest(flags)

	op, err := sdk.WrapOperation(sdk.MDB().PostgreSQL().Cluster().Create(ctx, req))

	if err != nil {
		log.Fatal(err)
	}
	meta, err := op.Metadata()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Creating cluster %s\n",
		meta.(*postgresql.CreateClusterMetadata).ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := op.Response()
	if err != nil {
		log.Fatal(err)
	}

	return resp.(*postgresql.Cluster)
}

func addClusterHost(ctx context.Context, sdk *ycsdk.SDK, cluster *postgresql.Cluster, params *cmdFlags) {
	fmt.Printf("Adding host to cluster %s\n", cluster.Id)
	hostSpec := postgresql.HostSpec{
		ZoneId:         *params.zoneID,
		SubnetId:       *params.subnetID,
		AssignPublicIp: false}

	hostSpecs := []*postgresql.HostSpec{&hostSpec}
	req := postgresql.AddClusterHostsRequest{ClusterId: cluster.Id, HostSpecs: hostSpecs}
	op, err := sdk.WrapOperation(sdk.MDB().PostgreSQL().Cluster().AddHosts(ctx, &req))
	if err != nil {
		log.Panic(err)
	}
	err = op.Wait(ctx)
	if err != nil {
		log.Panic(err)
	}
}

func changeClusterDescription(ctx context.Context, sdk *ycsdk.SDK, cluster *postgresql.Cluster) {
	fmt.Printf("Updating cluster %s\n", cluster.Id)
	mask := &field_mask.FieldMask{
		Paths: []string{
			"description",
		},
	}
	req := postgresql.UpdateClusterRequest{ClusterId: cluster.Id, UpdateMask: mask, Description: "New Description!!!"}
	op, err := sdk.WrapOperation(sdk.MDB().PostgreSQL().Cluster().Update(ctx, &req))
	if err != nil {
		log.Panic(err)
	}
	err = op.Wait(ctx)
	if err != nil {
		log.Panic(err)
	}
}

func deleteCluster(ctx context.Context, sdk *ycsdk.SDK, cluster *postgresql.Cluster) {
	fmt.Printf("Deleting cluster %s\n", cluster.Id)
	op, err := sdk.WrapOperation(sdk.MDB().PostgreSQL().Cluster().Delete(ctx, &postgresql.DeleteClusterRequest{ClusterId: cluster.Id}))
	if err != nil {
		log.Fatal(err)
	}
	err = op.Wait(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func createClusterRequest(params *cmdFlags) *postgresql.CreateClusterRequest {
	dbSpec := postgresql.DatabaseSpec{Name: *params.dbName, Owner: *params.userName}
	dbSpecs := []*postgresql.DatabaseSpec{&dbSpec}

	permission := postgresql.Permission{DatabaseName: *params.dbName}
	permissions := []*postgresql.Permission{&permission}
	userSpec := postgresql.UserSpec{Name: *params.userName, Password: *params.userPassword, Permissions: permissions}
	userSpecs := []*postgresql.UserSpec{&userSpec}

	hostSpec := postgresql.HostSpec{
		ZoneId:         *params.zoneID,
		SubnetId:       *params.subnetID,
		AssignPublicIp: false}

	hostSpecs := []*postgresql.HostSpec{&hostSpec}

	res := &postgresql.Resources{ResourcePresetId: "s1.nano", DiskSize: int64(10 * datasize.GB.Bytes()), DiskTypeId: "network-nvme"}
	configSpec := postgresql.ConfigSpec{Version: "10", Resources: res}

	req := postgresql.CreateClusterRequest{
		FolderId:      *params.folderID,
		Name:          *params.clusterName,
		Description:   *params.clusterDesc,
		Environment:   postgresql.Cluster_PRODUCTION,
		ConfigSpec:    &configSpec,
		DatabaseSpecs: dbSpecs,
		UserSpecs:     userSpecs,
		HostSpecs:     hostSpecs,
		NetworkId:     *params.networkID,
	}
	return &req
}
