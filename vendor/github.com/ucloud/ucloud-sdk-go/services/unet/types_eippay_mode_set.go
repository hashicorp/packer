package unet

/*
EIPPayModeSet - GetEIPPayModeEIP

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type EIPPayModeSet struct {

	// EIP的资源ID
	EIPId string

	// EIP的计费模式. 枚举值为：Bandwidth, 带宽计费;Traffic, 流量计费; "ShareBandwidth",共享带宽模式
	EIPPayMode string
}
