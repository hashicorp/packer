package uhost

/*
UHostIPSet - DescribeUHostInstance

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UHostIPSet struct {

	// 【暂未支持】是否为默认网卡。True: 是默认网卡；其他值：不是。
	Default string

	// 当前网卡的Mac。
	Mac string

	// 当前EIP的权重。权重最大的为当前的出口IP。
	Weight int

	// 国际: Internation，BGP: Bgp，内网: Private
	Type string

	// 外网IP资源ID 。(内网IP无对应的资源ID)
	IPId string

	// IP地址
	IP string

	// IP对应的带宽, 单位: Mb  (内网IP不显示带宽信息)
	Bandwidth int

	// IP地址对应的VPC ID。（北京一不支持，字段返回为空）
	VPCId string

	// IP地址对应的子网 ID。（北京一不支持，字段返回为空）
	SubnetId string
}
