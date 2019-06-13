package uhost

/*
UHostDisk - the request query for disk of uhost
*/
type UHostDisk struct {
	// 磁盘大小，单位GB。请参考[[api:uhost-api:disk_type|磁盘类型]]。
	Size *int `required:"true"`

	// 磁盘类型。枚举值：LOCAL_NORMAL 普通本地盘 | CLOUD_NORMAL 普通云盘 |LOCAL_SSD SSD本地盘 | CLOUD_SSD SSD云盘，默认为LOCAL_NORMAL。请参考[[api:uhost-api:disk_type|磁盘类型]]。
	Type *string `required:"true"`

	// 是否是系统盘。枚举值：\\ > True，是系统盘 \\ > False，是数据盘（默认）。Disks数组中有且只能有一块盘是系统盘。
	IsBoot *string `required:"true"`

	// 磁盘备份方案。枚举值：\\ > NONE，无备份 \\ > DATAARK，数据方舟 \\ 当前磁盘支持的备份模式参考 [[api:uhost-api:disk_type|磁盘类型]]
	BackupType *string `required:"false"`
}
