package unet

/*
UnetBandwidthPackageSet - DescribeBandwidthPackage

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UnetBandwidthPackageSet struct {

	// 带宽包的资源ID
	BandwidthPackageId string

	// 生效时间, 格式为 Unix Timestamp
	EnableTime int

	// 失效时间, 格式为 Unix Timestamp
	DisableTime int

	// 创建时间, 格式为 Unix Timestamp
	CreateTime int

	// 带宽包的临时带宽值, 单位Mbps
	Bandwidth int

	// 带宽包所绑定弹性IP的资源ID
	EIPId string

	// 带宽包所绑定弹性IP的详细信息,只有当EIPId对应双线IP时, EIPAddr的长度为2, 其他情况, EIPAddr长度均为1.参见 EIPAddrSet
	EIPAddr []EIPAddrSet
}
