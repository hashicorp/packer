package unet

/*
VIPDetailSet - VIPDetailSet

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type VIPDetailSet struct {

	// 地域
	Zone string

	// 虚拟ip id
	VIPId string

	// 创建时间
	CreateTime int

	// 真实主机ip
	RealIp string

	// 虚拟ip
	VIP string

	// 子网id
	SubnetId string

	// VPC id
	VPCId string

	// Virtual IP 名称
	Name string
}
