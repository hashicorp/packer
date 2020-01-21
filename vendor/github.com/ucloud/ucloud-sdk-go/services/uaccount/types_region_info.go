package uaccount

/*
RegionInfo - 数据中心信息

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type RegionInfo struct {

	// 数据中心ID
	RegionId int

	// 数据中心名称
	RegionName string

	// 是否用户当前默认数据中心
	IsDefault bool

	// 用户在此数据中心的权限位
	BitMaps string

	// 地域名字，如cn-bj
	Region string

	// 可用区名字，如cn-bj-01
	Zone string
}
