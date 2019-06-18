package unet

/*
FirewallDataSet - DescribeFirewall

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type FirewallDataSet struct {

	// 防火墙ID
	FWId string

	// 安全组ID（即将废弃）
	GroupId string

	// 防火墙名称
	Name string

	// 防火墙业务组
	Tag string

	// 防火墙备注
	Remark string

	// 防火墙绑定资源数量
	ResourceCount int

	// 防火墙组创建时间，格式为Unix Timestamp
	CreateTime int

	// 防火墙组类型，枚举值为： "user defined", 用户自定义防火墙； "recommend web", 默认Web防火墙； "recommend non web", 默认非Web防火墙
	Type string

	// 防火墙组中的规则列表，参见 FirewallRuleSet
	Rule []FirewallRuleSet
}
