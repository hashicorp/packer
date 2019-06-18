package uaccount

/*
UserInfo - 用户信息

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UserInfo struct {

	// 用户Id
	UserId int

	// 用户邮箱
	UserEmail string

	// 用户手机
	UserPhone string

	// 国际号码前缀
	PhonePrefix string

	// 会员类型
	UserType int

	// 称呼
	UserName string

	// 公司名称
	CompanyName string

	// 所属行业
	IndustryType int

	// 省份
	Province string

	// 城市
	City string

	// 公司地址
	UserAddress string

	// 是否超级管理员 0:否 1:是
	Admin int

	// 是否子帐户(大于100为子帐户)
	UserVersion int

	// 是否有财务权限 0:否 1:是
	Finance int

	// 管理员
	Administrator string

	// 实名认证状态
	AuthState string
}
