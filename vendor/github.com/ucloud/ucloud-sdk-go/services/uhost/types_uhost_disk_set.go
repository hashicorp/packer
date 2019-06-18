package uhost

/*
UHostDiskSet - DescribeUHostInstance

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UHostDiskSet struct {

	// 磁盘类型。请参考[[api:uhost-api:disk_type|磁盘类型]]。
	DiskType string

	// 是否是系统盘。枚举值：\\ > True，是系统盘 \\ > False，是数据盘（默认）。Disks数组中有且只能有一块盘是系统盘。
	IsBoot string

	// 【建议不再使用】磁盘类型。系统盘: Boot，数据盘: Data,网络盘：Udisk
	Type string

	// 磁盘ID
	DiskId string

	// UDisk名字（仅当磁盘是UDisk时返回）
	Name string

	// 磁盘盘符
	Drive string

	// 磁盘大小，单位: GB
	Size int

	// 备份方案。若开通了数据方舟，则为DataArk
	BackupType string
}
