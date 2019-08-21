package uhost

/*
UHostPriceSet - 主机价格

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UHostPriceSet struct {

	// 计费类型。Year，Month，Dynamic
	ChargeType string

	// 价格，单位: 元，保留小数点后两位有效数字
	Price float64
}
