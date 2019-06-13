package unet

/*
EIPPriceDetailSet - GetEIPPrice

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type EIPPriceDetailSet struct {

	// 弹性IP付费方式
	ChargeType string

	// 弹性IP价格, 单位"元"
	Price float64

	// 资源有效期, 以Unix Timestamp表示
	PurchaseValue int
}
