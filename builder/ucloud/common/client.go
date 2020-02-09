package common

import (
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/services/ufile"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

type UCloudClient struct {
	UHostConn    *uhost.UHostClient
	UNetConn     *unet.UNetClient
	VPCConn      *vpc.VPCClient
	UAccountConn *uaccount.UAccountClient
	UFileConn    *ufile.UFileClient
}

func (c *UCloudClient) DescribeFirewallById(sgId string) (*unet.FirewallDataSet, error) {
	if sgId == "" {
		return nil, NewNotFoundError("security group", sgId)
	}
	conn := c.UNetConn

	req := conn.NewDescribeFirewallRequest()
	req.FWId = ucloud.String(sgId)

	resp, err := conn.DescribeFirewall(req)

	if err != nil {
		if uErr, ok := err.(uerr.Error); ok && uErr.Code() == 54002 {
			return nil, NewNotFoundError("security group", sgId)
		}
		return nil, err
	}

	if len(resp.DataSet) < 1 {
		return nil, NewNotFoundError("security group", sgId)
	}

	return &resp.DataSet[0], nil
}

func (c *UCloudClient) DescribeSubnetById(subnetId string) (*vpc.VPCSubnetInfoSet, error) {
	if subnetId == "" {
		return nil, NewNotFoundError("Subnet", subnetId)
	}
	conn := c.VPCConn

	req := conn.NewDescribeSubnetRequest()
	req.SubnetIds = []string{subnetId}

	resp, err := conn.DescribeSubnet(req)
	if err != nil {
		return nil, err
	}

	if resp == nil || len(resp.DataSet) < 1 {
		return nil, NewNotFoundError("Subnet", subnetId)
	}

	return &resp.DataSet[0], nil
}

func (c *UCloudClient) DescribeVPCById(vpcId string) (*vpc.VPCInfo, error) {
	if vpcId == "" {
		return nil, NewNotFoundError("VPC", vpcId)
	}
	conn := c.VPCConn

	req := conn.NewDescribeVPCRequest()
	req.VPCIds = []string{vpcId}

	resp, err := conn.DescribeVPC(req)
	if err != nil {
		return nil, err
	}

	if resp == nil || len(resp.DataSet) < 1 {
		return nil, NewNotFoundError("VPC", vpcId)
	}

	return &resp.DataSet[0], nil
}

func (c *UCloudClient) DescribeImageById(imageId string) (*uhost.UHostImageSet, error) {
	if imageId == "" {
		return nil, NewNotFoundError("image", imageId)
	}
	req := c.UHostConn.NewDescribeImageRequest()
	req.ImageId = ucloud.String(imageId)

	resp, err := c.UHostConn.DescribeImage(req)
	if err != nil {
		return nil, err
	}

	if len(resp.ImageSet) < 1 {
		return nil, NewNotFoundError("image", imageId)
	}

	return &resp.ImageSet[0], nil
}

func (c *UCloudClient) DescribeUHostById(uhostId string) (*uhost.UHostInstanceSet, error) {
	if uhostId == "" {
		return nil, NewNotFoundError("instance", uhostId)
	}
	req := c.UHostConn.NewDescribeUHostInstanceRequest()
	req.UHostIds = []string{uhostId}

	resp, err := c.UHostConn.DescribeUHostInstance(req)
	if err != nil {
		return nil, err
	}
	if len(resp.UHostSet) < 1 {
		return nil, nil
	}

	return &resp.UHostSet[0], nil
}

func (c *UCloudClient) DescribeImageByInfo(projectId, regionId, imageId string) (*uhost.UHostImageSet, error) {
	req := c.UHostConn.NewDescribeImageRequest()
	req.ProjectId = ucloud.String(projectId)
	req.ImageId = ucloud.String(imageId)
	req.Region = ucloud.String(regionId)

	resp, err := c.UHostConn.DescribeImage(req)
	if err != nil {
		return nil, err
	}

	if len(resp.ImageSet) < 1 {
		return nil, NewNotFoundError("image", imageId)
	}

	return &resp.ImageSet[0], nil

}
