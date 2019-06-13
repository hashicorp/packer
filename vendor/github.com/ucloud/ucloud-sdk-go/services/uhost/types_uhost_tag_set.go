package uhost

/*
UHostTagSet - DescribeUHostTags

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UHostTagSet struct {

	// 业务组名称
	Tag string

	// 该业务组中包含的主机个数
	TotalCount int

	// 可用区
	Zone string
}
