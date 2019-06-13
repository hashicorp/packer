package unet

/*
UnetEIPSet - DescribeEIP

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UnetEIPSet struct {

	// 弹性IP的资源ID
	EIPId string

	// 外网出口权重, 默认为50, 范围[0-100]
	Weight int

	// 带宽模式, 枚举值为: 0: 非共享带宽模式, 1: 共享带宽模式
	BandwidthType int

	// 弹性IP的带宽, 单位为Mbps, 当BandwidthType=1时, 该处显示为共享带宽值. 当BandwidthType=0时, 该处显示这个弹性IP的带宽.
	Bandwidth int

	// 弹性IP的资源绑定状态, 枚举值为: used: 已绑定, free: 未绑定, freeze: 已冻结
	Status string

	// 付费方式, 枚举值为: Year, 按年付费; Month, 按月付费; Dynamic, 按小时付费; Trial, 试用. 按小时付费和试用这两种付费模式需要开通权限.
	ChargeType string

	// 弹性IP的创建时间, 格式为Unix Timestamp
	CreateTime int

	// 弹性IP的到期时间, 格式为Unix Timestamp
	ExpireTime int

	// 弹性IP的详细信息列表, 具体结构见下方 UnetEIPResourceSet
	Resource UnetEIPResourceSet

	// 弹性IP的详细信息列表, 具体结构见下方 UnetEIPAddrSet
	EIPAddr []UnetEIPAddrSet

	// 弹性IP的名称,缺省值为 "EIP"
	Name string

	// 弹性IP的业务组标识, 缺省值为 "Default"
	Tag string

	// 弹性IP的备注, 缺省值为 ""
	Remark string

	// 弹性IP的计费模式, 枚举值为: "Bandwidth", 带宽计费; "Traffic", 流量计费; "ShareBandwidth",共享带宽模式. 默认为 "Bandwidth".
	PayMode string

	// 共享带宽信息 参见 ShareBandwidthSet
	ShareBandwidthSet ShareBandwidthSet

	// 弹性IP是否到期
	Expire bool
}
