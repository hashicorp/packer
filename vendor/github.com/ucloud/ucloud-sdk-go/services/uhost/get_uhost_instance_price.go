package uhost

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

// GetUHostInstancePriceRequest is request schema for GetUHostInstancePrice action
type GetUHostInstancePriceRequest struct {
	request.CommonBase

	// [公共参数] 地域。 参见 [地域和可用区列表](../summary/regionlist.html)
	// Region *string `required:"true"`

	// [公共参数] 可用区。参见 [可用区列表](../summary/regionlist.html)
	// Zone *string `required:"false"`

	// [公共参数] 项目ID。不填写为默认项目，子帐号必须填写。 请参考[GetProjectList接口](../summary/get_project_list.html)
	// ProjectId *string `required:"false"`

	// 镜像Id，可通过 [DescribeImage](describe_image.html) 获取镜像ID
	ImageId *string `required:"true"`

	// 虚拟CPU核数。可选参数：1-32（可选范围与UHostType相关）。默认值: 4
	CPU *int `required:"true"`

	// 内存大小。单位：MB。范围 ：[1024, 262144]，取值为1024的倍数（可选范围与UHostType相关）。默认值：8192
	Memory *int `required:"true"`

	// 【未启用】购买台数，范围[1,5]
	Count *int `required:"false"`

	// 磁盘列表
	Disks []UHostDisk

	// GPU卡核心数。仅GPU机型支持此字段（可选范围与UHostType相关）。
	GPU *int `required:"false"`

	// 计费模式。枚举值为： \\ > Year，按年付费； \\ > Month，按月付费；\\ > Dynamic，按小时付费 \\ 默认为月付。
	ChargeType *string `required:"false"`

	// 【待废弃】磁盘类型，同时设定系统盘和数据盘， 枚举值为：LocalDisk，本地磁盘; UDisk，云硬盘; 默认为LocalDisk 仅部分可用区支持云硬盘方式的主机存储方式，具体请查询控制台。
	StorageType *string `required:"false"`

	// 【待废弃】数据盘大小，单位: GB，范围[0,1000]，步长: 10，默认值: 0
	DiskSpace *int `required:"false"`

	// 网络增强。枚举值：\\ > Normal，不开启 \\ > Super，开启 \\ 默认值未为Normal。
	NetCapability *string `required:"false"`

	// 【待废弃】方舟机型。No，Yes。默认是No。
	TimemachineFeature *string `required:"false"`

	// 主机类型 Normal: 标准机型 SSD：SSD机型 BigData:大数据    GPU:GPU型G1(原GPU型)   GPU_G2:GPU型G2 GPU_G3:GPU型G3  不同机房的主机类型支持情况不同。详情请参考控制台。
	UHostType *string `required:"false"`
	// 【未支持】1：普通云主机；2：抢占性云主机；默认普通
	LifeCycle *int `required:"false"`

	// 购买时长。默认: 1。按小时购买(Dynamic)时无需此参数。 月付时，此参数传0，代表了购买至月末。
	Quantity *int `required:"false"`
}

// GetUHostInstancePriceResponse is response schema for GetUHostInstancePrice action
type GetUHostInstancePriceResponse struct {
	response.CommonBase

	// 价格列表 UHostPriceSet
	PriceSet []UHostPriceSet
}

// NewGetUHostInstancePriceRequest will create request of GetUHostInstancePrice action.
func (c *UHostClient) NewGetUHostInstancePriceRequest() *GetUHostInstancePriceRequest {
	req := &GetUHostInstancePriceRequest{}

	// setup request with client config
	c.Client.SetupRequest(req)

	// setup retryable with default retry policy (retry for non-create action and common error)
	req.SetRetryable(true)
	return req
}

// GetUHostInstancePrice - 根据UHost实例配置，获取UHost实例的价格。
func (c *UHostClient) GetUHostInstancePrice(req *GetUHostInstancePriceRequest) (*GetUHostInstancePriceResponse, error) {
	var err error
	var res GetUHostInstancePriceResponse

	err = c.Client.InvokeAction("GetUHostInstancePrice", req, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
