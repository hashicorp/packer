package unet

/*
EIPSetData - describeShareBandwidth

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type EIPSetData struct {

	// EIP带宽值
	Bandwidth int

	// EIP的IP信息，详情见EIPAddrSet
	EIPAddr []EIPAddrSet

	// EIP资源Id
	EIPId string
}
