package uhost

/*
IsolationGroup - 硬件隔离组信息

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type IsolationGroup struct {

	// 硬件隔离组名称
	GroupName string

	// 硬件隔离组id
	GroupId string

	// 每个可用区中的机器数量。参见数据结构SpreadInfo。
	SpreadInfoSet []SpreadInfo

	// 备注
	Remark string
}
