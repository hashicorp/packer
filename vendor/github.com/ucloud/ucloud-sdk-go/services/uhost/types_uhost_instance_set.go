package uhost

/*
UHostInstanceSet - DescribeUHostInstance

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UHostInstanceSet struct {

	// 可用区。参见 [可用区列表](../summary/regionlist.html)
	Zone string

	// UHost实例ID
	UHostId string

	// 【建议不再使用】云主机机型（旧）。参考[[api:uhost-api:uhost_type|云主机机型说明]]。
	UHostType string

	// 云主机机型（新）。参考[[api:uhost-api:uhost_type|云主机机型说明]]。
	MachineType string

	// 【建议不再使用】主机磁盘类型。 枚举值为：\\ > LocalDisk，本地磁盘; \\ > UDisk 云盘。\\只要有一块磁盘为本地盘，即返回LocalDisk。
	StorageType string

	// 【建议不再使用】主机的系统盘ID。
	ImageId string

	// 基础镜像ID（指当前自定义镜像的来源镜像）
	BasicImageId string

	// 基础镜像名称（指当前自定义镜像的来源镜像）
	BasicImageName string

	// 业务组名称
	Tag string

	// 备注
	Remark string

	// UHost实例名称
	Name string

	// 实例状态，枚举值：\\ >初始化: Initializing; \\ >启动中: Starting; \\> 运行中: Running; \\> 关机中: Stopping; \\ >关机: Stopped \\ >安装失败: Install Fail; \\ >重启中: Rebooting
	State string

	// 创建时间，格式为Unix时间戳
	CreateTime int

	// 计费模式，枚举值为： Year，按年付费； Month，按月付费； Dynamic，按需付费（需开启权限）；
	ChargeType string

	// 到期时间，格式为Unix时间戳
	ExpireTime int

	// 虚拟CPU核数，单位: 个
	CPU int

	// 内存大小，单位: MB
	Memory int

	// 是否自动续费，自动续费：“Yes”，不自动续费：“No”
	AutoRenew string

	// 磁盘信息见 UHostDiskSet
	DiskSet []UHostDiskSet

	// 详细信息见 UHostIPSet
	IPSet []UHostIPSet

	// 网络增强。Normal: 无；Super： 网络增强1.0； Ultra: 网络增强2.0
	NetCapability string

	// 【建议不再使用】网络状态。 连接：Connected， 断开：NotConnected
	NetworkState string

	// 【建议不再使用】数据方舟模式。枚举值：\\ > Yes: 开启方舟； \\ > no，未开启方舟
	TimemachineFeature string

	// true: 开启热升级； false，未开启热升级
	HotplugFeature bool

	// 【建议不再使用】仅北京A的云主机会返回此字段。基础网络模式：Default；子网模式：Private
	SubnetType string

	// 内网的IP地址
	IPs []string

	// 创建主机的最初来源镜像的操作系统名称（若直接通过基础镜像创建，此处返回和BasicImageName一致）
	OsName string

	// 操作系统类别。返回"Linux"或者"Windows"
	OsType string

	// 删除时间，格式为Unix时间戳
	DeleteTime int

	// 主机系列：N2，表示系列2；N1，表示系列1
	HostType string

	// 主机的生命周期类型。目前仅支持Normal：普通；
	LifeCycle string

	// GPU个数
	GPU int

	// 系统盘状态 Normal表示初始化完成；Initializing表示在初始化。仍在初始化的系统盘无法制作镜像。
	BootDiskState string

	// 总的数据盘存储空间。
	TotalDiskSpace int

	// 隔离组id，不在隔离组则返回""
	IsolationGroup string
}
