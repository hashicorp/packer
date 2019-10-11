package unet

/*
FirewallRuleSet - DescribeFirewall

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type FirewallRuleSet struct {

	// 源地址
	SrcIP string

	// 优先级
	Priority string

	// 协议类型
	ProtocolType string

	// 目标端口
	DstPort string

	// 防火墙动作
	RuleAction string

	// 防火墙规则备注
	Remark string
}
