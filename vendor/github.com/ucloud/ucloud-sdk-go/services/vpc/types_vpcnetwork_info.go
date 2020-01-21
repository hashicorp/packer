package vpc

// VPCNetworkInfo - vpc地址空间信息
type VPCNetworkInfo struct {

	// 地址空间段
	Network string

	// 地址空间中子网数量
	SubnetCount int
}
