package uaccount

/*
ProjectListInfo - 项目信息

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type ProjectListInfo struct {

	// 项目ID
	ProjectId string

	// 项目名称
	ProjectName string

	// 父项目ID
	ParentId string

	// 父项目名称
	ParentName string

	// 创建时间(Unix时间戳)
	CreateTime int

	// 是否为默认项目
	IsDefault bool

	// 项目下资源数量
	ResourceCount int

	// 项目下成员数量
	MemberCount int
}
