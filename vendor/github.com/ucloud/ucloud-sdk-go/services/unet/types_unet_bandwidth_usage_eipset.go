package unet

/*
UnetBandwidthUsageEIPSet - DescribeBandwidthUsage

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UnetBandwidthUsageEIPSet struct {

	// 最近5分钟带宽用量, 单位Mbps
	CurBandwidth float64

	// 弹性IP资源ID
	EIPId string
}
