package vpc

/*
VPCIntercomInfo -

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type VPCIntercomInfo struct {

	// 项目Id
	ProjectId string

	// VPC的地址空间
	Network []string

	// 所属地域
	DstRegion string

	// VPC名字
	Name string

	// VPCId
	VPCId string

	// 业务组（未分组显示为 Default）
	Tag string
}
