package vpc

/*
VPCSubnetInfoSet - DescribeSubnet

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type VPCSubnetInfoSet struct {

	// VPC id
	VPCId string

	// VPC名称
	VPCName string

	// 子网id
	SubnetId string

	// 子网名称
	SubnetName string

	// 地址
	Zone string

	// 名称
	Name string

	// 备注
	Remark string

	// Tag
	Tag string

	// 子网类型
	SubnetType int

	// 子网网段
	Subnet string

	// 子网掩码
	Netmask string

	// 子网网关
	Gateway string

	// 创建时间
	CreateTime int

	// 虚拟路由 id
	VRouterId string

	// 是否关联NATGW
	HasNATGW bool
}
