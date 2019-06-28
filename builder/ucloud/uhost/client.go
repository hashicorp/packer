package uhost

import (
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

type UCloudClient struct {
	uhostconn    *uhost.UHostClient
	unetconn     *unet.UNetClient
	vpcconn      *vpc.VPCClient
	uaccountconn *uaccount.UAccountClient
}

func (c *UCloudClient) describeFirewallById(sgId string) (*unet.FirewallDataSet, error) {
	if sgId == "" {
		return nil, newNotFoundError("security group", sgId)
	}
	conn := c.unetconn

	req := conn.NewDescribeFirewallRequest()
	req.FWId = ucloud.String(sgId)

	resp, err := conn.DescribeFirewall(req)

	if err != nil {
		if uErr, ok := err.(uerr.Error); ok && uErr.Code() == 54002 {
			return nil, newNotFoundError("security group", sgId)
		}
		return nil, err
	}

	if len(resp.DataSet) < 1 {
		return nil, newNotFoundError("security group", sgId)
	}

	return &resp.DataSet[0], nil
}

func (c *UCloudClient) describeSubnetById(subnetId string) (*vpc.VPCSubnetInfoSet, error) {
	if subnetId == "" {
		return nil, newNotFoundError("Subnet", subnetId)
	}
	conn := c.vpcconn

	req := conn.NewDescribeSubnetRequest()
	req.SubnetIds = []string{subnetId}

	resp, err := conn.DescribeSubnet(req)
	if err != nil {
		return nil, err
	}

	if resp == nil || len(resp.DataSet) < 1 {
		return nil, newNotFoundError("Subnet", subnetId)
	}

	return &resp.DataSet[0], nil
}

func (c *UCloudClient) describeVPCById(vpcId string) (*vpc.VPCInfo, error) {
	if vpcId == "" {
		return nil, newNotFoundError("VPC", vpcId)
	}
	conn := c.vpcconn

	req := conn.NewDescribeVPCRequest()
	req.VPCIds = []string{vpcId}

	resp, err := conn.DescribeVPC(req)
	if err != nil {
		return nil, err
	}

	if resp == nil || len(resp.DataSet) < 1 {
		return nil, newNotFoundError("VPC", vpcId)
	}

	return &resp.DataSet[0], nil
}

func (c *UCloudClient) DescribeImageById(imageId string) (*uhost.UHostImageSet, error) {
	if imageId == "" {
		return nil, newNotFoundError("image", imageId)
	}
	req := c.uhostconn.NewDescribeImageRequest()
	req.ImageId = ucloud.String(imageId)

	resp, err := c.uhostconn.DescribeImage(req)
	if err != nil {
		return nil, err
	}

	if len(resp.ImageSet) < 1 {
		return nil, newNotFoundError("image", imageId)
	}

	return &resp.ImageSet[0], nil
}

func (c *UCloudClient) describeUHostById(uhostId string) (*uhost.UHostInstanceSet, error) {
	if uhostId == "" {
		return nil, newNotFoundError("instance", uhostId)
	}
	req := c.uhostconn.NewDescribeUHostInstanceRequest()
	req.UHostIds = []string{uhostId}

	resp, err := c.uhostconn.DescribeUHostInstance(req)
	if err != nil {
		return nil, err
	}
	if len(resp.UHostSet) < 1 {
		return nil, nil
	}

	return &resp.UHostSet[0], nil
}

func (c *UCloudClient) describeImageByInfo(projectId, regionId, imageId string) (*uhost.UHostImageSet, error) {
	req := c.uhostconn.NewDescribeImageRequest()
	req.ProjectId = ucloud.String(projectId)
	req.ImageId = ucloud.String(imageId)
	req.Region = ucloud.String(regionId)

	resp, err := c.uhostconn.DescribeImage(req)
	if err != nil {
		return nil, err
	}

	if len(resp.ImageSet) < 1 {
		return nil, newNotFoundError("image", imageId)
	}

	return &resp.ImageSet[0], nil

}
