package unet

/*
UnetAllocateEIPSet - AllocateEIP

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UnetAllocateEIPSet struct {

	// 申请到的EIP资源ID
	EIPId string

	// 申请到的IPv4地址.
	EIPAddr []UnetEIPAddrSet
}
