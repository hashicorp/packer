package uhost

/*
UHostImageSet - DescribeImage

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UHostImageSet struct {

	// 可用区，参见 [可用区列表](../summary/regionlist.html) |
	Zone string

	// 镜像ID
	ImageId string

	// 镜像名称
	ImageName string

	// 操作系统类型：Liunx，Windows
	OsType string

	// 操作系统名称
	OsName string

	// 镜像类型 标准镜像：Base， 行业镜像：Business，自定义镜像：Custom
	ImageType string

	// 特殊状态标识， 目前包含NetEnhnced（网络增强1.0）, NetEnhanced_Ultra]（网络增强2.0）
	Features []string

	// 行业镜像类型（仅行业镜像将返回这个值）
	FuncType string

	// 集成软件名称（仅行业镜像将返回这个值）
	IntegratedSoftware string

	// 供应商（仅行业镜像将返回这个值）
	Vendor string

	// 介绍链接（仅行业镜像将返回这个值）
	Links string

	// 镜像状态， 可用：Available，制作中：Making， 不可用：Unavailable
	State string

	// 镜像描述
	ImageDescription string

	// 创建时间，格式为Unix时间戳
	CreateTime int

	// 镜像大小
	ImageSize int

	// 默认值为空'''。当CentOS 7.3/7.4/7.5等镜像会标记为“Broadwell”
	MinimalCPU string
}
