package unet

/*
UnetShareBandwidthSet - DescribeShareBandwidth

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UnetShareBandwidthSet struct {

	// 共享带宽值(预付费)/共享带宽峰值(后付费), 单位Mbps
	ShareBandwidth int

	// 共享带宽的资源ID
	ShareBandwidthId string

	// 付费方式, 预付费:Year 按年,Month 按月,Dynamic 按需;后付费:PostPay(按月)
	ChargeType string

	// 创建时间, 格式为Unix Timestamp
	CreateTime int

	// 过期时间, 格式为Unix Timestamp
	ExpireTime int

	// EIP信息,详情见 EIPSetData
	EIPSet []EIPSetData

	// 共享带宽保底值(后付费)
	BandwidthGuarantee int

	// 共享带宽后付费开始计费时间(后付费)
	PostPayStartTime int

	// 共享带宽名称
	Name string
}
