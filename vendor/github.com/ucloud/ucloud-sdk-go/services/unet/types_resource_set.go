package unet

/*
ResourceSet - 资源信息

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type ResourceSet struct {

	// 可用区
	Zone int

	// 名称
	Name string

	// 内网IP
	PrivateIP string

	// 备注
	Remark string

	// 绑定该防火墙的资源id
	ResourceID string

	// 绑定防火墙组的资源类型。"unatgw"，NAT网关； "uhost"，云主机； "upm"，物理云主机； "hadoophost"，hadoop节点； "fortresshost"，堡垒机； "udhost"，私有专区主机；"udockhost"，容器；"dbaudit"，数据库审计.
	ResourceType string

	// 状态
	Status int

	// 业务组
	Tag string
}
