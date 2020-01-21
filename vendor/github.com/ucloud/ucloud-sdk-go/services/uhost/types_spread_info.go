package uhost

/*
SpreadInfo - 每个可用区中硬件隔离组信息

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type SpreadInfo struct {

	// 可用区信息
	Zone string

	// 可用区中硬件隔离组中云主机的数量，不超过7。
	UHostCount int
}
