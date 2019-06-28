package vpc

// VPCInfo - vpc信息
type VPCInfo struct {

	// 业务组
	Tag string

	// 创建时间
	CreateTime int

	// vpc名称
	Name string

	// vpc地址空间
	Network []string

	// vpc地址空间信息
	NetworkInfo []VPCNetworkInfo

	// vpc中子网数量
	SubnetCount int

	// 更新时间
	UpdateTime int

	// vpc的资源ID
	VPCId string
}
