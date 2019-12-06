package uhost

import (
	"encoding/base64"

	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

// CreateUHostInstanceRequest is request schema for CreateUHostInstance action
type CreateUHostInstanceRequest struct {
	request.CommonBase

	// 可用区。参见 [可用区列表](../summary/regionlist.html)
	// Zone *string `required:"true"`

	// 镜像ID。 请通过 [DescribeImage](describe_image.html)获取
	ImageId *string `required:"true"`

	// UHost密码，LoginMode为Password时此项必须（密码需使用base64进行编码）
	Password *string `required:"true"`

	// 磁盘列表
	Disks []UHostDisk `required:"true"`

	// UHost实例名称。默认：UHost。请遵照[[api:uhost-api:specification|字段规范]]设定实例名称。
	Name *string `required:"false"`

	// 业务组。默认：Default（Default即为未分组）
	Tag *string `required:"false"`

	// 计费模式。枚举值为： Year，按年付费； Month，按月付费； Dynamic，按小时付费（需开启权限）。默认为月付
	ChargeType *string `required:"false"`

	// 购买时长。默认: 1。按小时购买(Dynamic)时无需此参数。 月付时，此参数传0，代表了购买至月末。
	Quantity *int `required:"false"`

	// 云主机机型。枚举值：N1：系列1标准型；N2：系列2标准型；I1: 系列1高IO型；I2，系列2高IO型； D1: 系列1大数据机型；G1: 系列1GPU型，型号为K80；G2：系列2GPU型，型号为P40；G3：系列2GPU型，型号为V100；北京A、北京C、上海二A、香港A可用区默认N1，其他机房默认N2。参考[[api:uhost-api:uhost_type|云主机机型说明]]。
	UHostType *string `required:"false"`

	// 虚拟CPU核数。可选参数：1-32（可选范围与UHostType相关）。默认值: 4
	CPU *int `required:"false"`

	// 内存大小。单位：MB。范围 ：[1024, 262144]，取值为1024的倍数（可选范围与UHostType相关）。默认值：8192
	Memory *int `required:"false"`

	// GPU卡核心数。仅GPU机型支持此字段；系列1可选1,2；系列2可选1,2,3,4。GPU可选数量与CPU有关联，详情请参考控制台。
	GPU *int `required:"false"`

	// 主机登陆模式。密码（默认选项）: Password，key: KeyPair（此项暂不支持）
	LoginMode *string `required:"false"`

	// 【暂不支持】Keypair公钥，LoginMode为KeyPair时此项必须
	KeyPair *string `required:"false"`

	// 【待废弃，不建议调用】磁盘类型，同时设定系统盘和数据盘的磁盘类型。枚举值为：LocalDisk，本地磁盘; UDisk，云硬盘；默认为LocalDisk。仅部分可用区支持云硬盘方式的主机存储方式，具体请查询控制台。
	StorageType *string `required:"false"`

	// 【待废弃，不建议调用】系统盘大小。 单位：GB， 范围[20,100]， 步长：10
	BootDiskSpace *int `required:"false"`

	// 【待废弃，不建议调用】数据盘大小。 单位：GB， 范围[0,8000]， 步长：10， 默认值：20，云盘支持0-8000；本地普通盘支持0-2000；本地SSD盘（包括所有GPU机型）支持100-1000
	DiskSpace *int `required:"false"`

	// 网络增强。目前仅Normal（不开启） 和Super（开启）可用。默认Normal。 不同机房的网络增强支持情况不同。详情请参考控制台。
	NetCapability *string `required:"false"`

	// 是否开启方舟特性。Yes为开启方舟，No为关闭方舟。目前仅选择普通本地盘+普通本地盘 或 SSD云盘+普通云盘的组合支持开启方舟。
	TimemachineFeature *string `required:"false"`

	// 是否开启热升级特性。True为开启，False为未开启，默认False。仅系列1云主机需要使用此字段，系列2云主机根据镜像是否支持云主机。
	HotplugFeature *bool `required:"false"`

	// 网络ID（VPC2.0情况下无需填写）。VPC1.0情况下，若不填写，代表选择基础网络； 若填写，代表选择子网。参见DescribeSubnet。
	NetworkId *string `required:"false"`

	// VPC ID。VPC2.0下需要填写此字段。
	VPCId *string `required:"false"`

	// 子网ID。VPC2.0下需要填写此字段。
	SubnetId *string `required:"false"`

	// 【数组】创建云主机时指定内网IP。当前只支持一个内网IP。调用方式举例：PrivateIp.0=x.x.x.x。
	PrivateIp []string `required:"false"`

	// 创建云主机时指定Mac。调用方式举例：PrivateMac="xx:xx:xx:xx:xx:xx"。
	PrivateMac *string `required:"false"`

	// 防火墙Id，默认：Web推荐防火墙。如何查询SecurityGroupId请参见 [DescribeSecurityGroup](../unet-api/describe_security_group.html)
	SecurityGroupId *string `required:"false"`

	// 硬件隔离组id。可通过DescribeIsolationGroup获取。
	IsolationGroup *string `required:"false"`

	// 【暂不支持】cloudinit方式下，用户初始化脚本
	UserDataScript *string `required:"false"`

	// 【已废弃】宿主机类型，N2，N1
	HostType *string `required:"false"`

	// 【暂不支持】是否安装UGA。'yes': 安装；其他或者不填：不安装。
	InstallAgent *string `required:"false"`

	// 【内部参数】资源类型
	ResourceType *int `required:"false"`

	// 代金券ID。请通过DescribeCoupon接口查询，或登录用户中心查看
	CouponId *string `required:"false"`

	// 云主机类型，枚举值["N", "C", "G", "O"]
	MachineType *string `required:"false"`

	// 最低cpu平台，枚举值["Intel/Auto", "Intel/LvyBridge", "Intel/Haswell", "Intel/Broadwell", "Intel/Skylake", "Intel/Cascadelake"(只有O型云主机可选)]
	MinimalCpuPlatform *string `required:"false"`

	// 【批量创建主机时必填】最大创建主机数量，取值范围是[1,100];
	MaxCount *int `required:"false"`

	// GPU类型，枚举值["K80", "P40", "V100"]，MachineType为G时必填
	GpuType *string `required:"false"`

	// NetworkInterface
	NetworkInterface []CreateUHostInstanceParamNetworkInterface
}

/*
CreateUHostInstanceParamNetworkInterface is request schema for complex param
*/
type CreateUHostInstanceParamNetworkInterface struct {

	// EIP
	EIP *CreateUHostInstanceParamNetworkInterfaceEIP
}

/*
CreateUHostInstanceParamNetworkInterfaceEIP is request schema for complex param
*/
type CreateUHostInstanceParamNetworkInterfaceEIP struct {

	// 弹性IP的计费模式. 枚举值: "Traffic", 流量计费; "Bandwidth", 带宽计费; "ShareBandwidth",共享带宽模式. "Free":免费带宽模式.默认为 "Bandwidth".
	PayMode *string

	// 当前EIP代金券id。请通过DescribeCoupon接口查询，或登录用户中心查看
	CouponId *string

	// 【如果绑定EIP这个参数必填】弹性IP的外网带宽, 单位为Mbps. 共享带宽模式必须指定0M带宽, 非共享带宽模式必须指定非0Mbps带宽. 各地域非共享带宽的带宽范围如下： 流量计费[1-300]，带宽计费[1-800]
	Bandwidth *int

	// 绑定的共享带宽Id，仅当PayMode为ShareBandwidth时有效
	ShareBandwidthId *string

	// GlobalSSH
	GlobalSSH *CreateUHostInstanceParamNetworkInterfaceEIPGlobalSSH

	// 【如果绑定EIP这个参数必填】弹性IP的线路如下: 国际: International BGP: Bgp 各地域允许的线路参数如下: cn-sh1: Bgp cn-sh2: Bgp cn-gd: Bgp cn-bj1: Bgp cn-bj2: Bgp hk: International us-ca: International th-bkk: International kr-seoul:International us-ws:International ge-fra:International sg:International tw-kh:International.其他海外线路均为 International
	OperatorName *string
}

/*
CreateUHostInstanceParamNetworkInterfaceEIPGlobalSSH is request schema for complex param
*/
type CreateUHostInstanceParamNetworkInterfaceEIPGlobalSSH struct {

	// 填写支持SSH访问IP的地区名称，如“洛杉矶”，“新加坡”，“香港”，“东京”，“华盛顿”，“法兰克福”。Area和AreaCode两者必填一个
	Area *string

	// AreaCode, 区域航空港国际通用代码。Area和AreaCode两者必填一个
	AreaCode *string

	// SSH端口，1-65535且不能使用80，443端口
	Port *int
}

// CreateUHostInstanceResponse is response schema for CreateUHostInstance action
type CreateUHostInstanceResponse struct {
	response.CommonBase

	// UHost实例Id集合
	UHostIds []string

	// IP信息
	IPs []string
}

// NewCreateUHostInstanceRequest will create request of CreateUHostInstance action.
func (c *UHostClient) NewCreateUHostInstanceRequest() *CreateUHostInstanceRequest {
	req := &CreateUHostInstanceRequest{}

	// setup request with client config
	c.Client.SetupRequest(req)

	// setup retryable with default retry policy (retry for non-create action and common error)
	req.SetRetryable(false)
	return req
}

// CreateUHostInstance - 指定数据中心，根据资源使用量创建指定数量的UHost实例。
func (c *UHostClient) CreateUHostInstance(req *CreateUHostInstanceRequest) (*CreateUHostInstanceResponse, error) {
	var err error
	var res CreateUHostInstanceResponse
	var reqImmutable = *req
	reqImmutable.Password = ucloud.String(base64.StdEncoding.EncodeToString([]byte(ucloud.StringValue(req.Password))))

	err = c.Client.InvokeAction("CreateUHostInstance", &reqImmutable, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
