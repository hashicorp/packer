// GENERATED FILE: DO NOT EDIT!

package oapi

// Types used by the API.
// implements the service definition of AcceptNetPeeringRequest
type AcceptNetPeeringRequest struct {
	DryRun       bool   `json:"DryRun,omitempty"`
	NetPeeringId string `json:"NetPeeringId,omitempty"`
}

// implements the service definition of AcceptNetPeeringResponse
type AcceptNetPeeringResponse struct {
	NetPeering      NetPeering      `json:"NetPeering,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of AccepterNet
type AccepterNet struct {
	AccountId string `json:"AccountId,omitempty"`
	IpRange   string `json:"IpRange,omitempty"`
	NetId     string `json:"NetId,omitempty"`
}

// implements the service definition of AccessLog
type AccessLog struct {
	IsEnabled           bool   `json:"IsEnabled,omitempty"`
	OsuBucketName       string `json:"OsuBucketName,omitempty"`
	OsuBucketPrefix     string `json:"OsuBucketPrefix,omitempty"`
	PublicationInterval int64  `json:"PublicationInterval,omitempty"`
}

// implements the service definition of Account
type Account struct {
	AccountId     string `json:"AccountId,omitempty"`
	City          string `json:"City,omitempty"`
	CompanyName   string `json:"CompanyName,omitempty"`
	Country       string `json:"Country,omitempty"`
	CustomerId    string `json:"CustomerId,omitempty"`
	Email         string `json:"Email,omitempty"`
	FirstName     string `json:"FirstName,omitempty"`
	JobTitle      string `json:"JobTitle,omitempty"`
	LastName      string `json:"LastName,omitempty"`
	Mobile        string `json:"Mobile,omitempty"`
	Phone         string `json:"Phone,omitempty"`
	StateProvince string `json:"StateProvince,omitempty"`
	VatNumber     string `json:"VatNumber,omitempty"`
	ZipCode       string `json:"ZipCode,omitempty"`
}

// implements the service definition of ApiKey
type ApiKey struct {
	AccountId string        `json:"AccountId,omitempty"`
	ApiKeyId  string        `json:"ApiKeyId,omitempty"`
	SecretKey string        `json:"SecretKey,omitempty"`
	State     string        `json:"State,omitempty"`
	Tags      []ResourceTag `json:"Tags,omitempty"`
	UserName  string        `json:"UserName,omitempty"`
}

// implements the service definition of ApplicationStickyCookiePolicy
type ApplicationStickyCookiePolicy struct {
	CookieName string `json:"CookieName,omitempty"`
	PolicyName string `json:"PolicyName,omitempty"`
}

// implements the service definition of Attribute
type Attribute struct {
	Key   string `json:"Key,omitempty"`
	Value string `json:"Value,omitempty"`
}

// implements the service definition of AuthenticateAccountRequest
type AuthenticateAccountRequest struct {
	DryRun   bool   `json:"DryRun,omitempty"`
	Login    string `json:"Login,omitempty"`
	Password string `json:"Password,omitempty"`
}

// implements the service definition of AuthenticateAccountResponse
type AuthenticateAccountResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of BackendVmsHealth
type BackendVmsHealth struct {
	Description string `json:"Description,omitempty"`
	State       string `json:"State,omitempty"`
	StateReason string `json:"StateReason,omitempty"`
	VmId        string `json:"VmId,omitempty"`
}

// implements the service definition of BlockDeviceMapping
type BlockDeviceMapping struct {
	Bsu               Bsu    `json:"Bsu,omitempty"`
	DeviceName        string `json:"DeviceName,omitempty"`
	NoDevice          string `json:"NoDevice,omitempty"`
	VirtualDeviceName string `json:"VirtualDeviceName,omitempty"`
}

// implements the service definition of BlockDeviceMappingCreated
type BlockDeviceMappingCreated struct {
	Bsu        BsuCreated `json:"Bsu,omitempty"`
	DeviceName string     `json:"DeviceName,omitempty"`
}

// implements the service definition of BlockDeviceMappingImage
type BlockDeviceMappingImage struct {
	Bsu               BsuToCreate `json:"Bsu,omitempty"`
	DeviceName        string      `json:"DeviceName,omitempty"`
	VirtualDeviceName string      `json:"VirtualDeviceName,omitempty"`
}

// implements the service definition of BlockDeviceMappingVmCreation
type BlockDeviceMappingVmCreation struct {
	Bsu               BsuToCreate `json:"Bsu,omitempty"`
	DeviceName        string      `json:"DeviceName,omitempty"`
	NoDevice          string      `json:"NoDevice,omitempty"`
	VirtualDeviceName string      `json:"VirtualDeviceName,omitempty"`
}

// implements the service definition of BlockDeviceMappingVmUpdate
type BlockDeviceMappingVmUpdate struct {
	Bsu               BsuToUpdateVm `json:"Bsu,omitempty"`
	DeviceName        string        `json:"DeviceName,omitempty"`
	NoDevice          string        `json:"NoDevice,omitempty"`
	VirtualDeviceName string        `json:"VirtualDeviceName,omitempty"`
}

// implements the service definition of Bsu
type Bsu struct {
	DeleteOnVmDeletion *bool  `json:"DeleteOnVmDeletion,omitempty"`
	Iops               int64  `json:"Iops,omitempty"`
	LinkDate           string `json:"LinkDate,omitempty"`
	SnapshotId         string `json:"SnapshotId,omitempty"`
	VolumeId           string `json:"VolumeId,omitempty"`
	VolumeSize         int64  `json:"VolumeSize,omitempty"`
	VolumeType         string `json:"VolumeType,omitempty"`
}

// implements the service definition of BsuCreated
type BsuCreated struct {
	DeleteOnVmDeletion *bool  `json:"DeleteOnVmDeletion,omitempty"`
	LinkDate           string `json:"LinkDate,omitempty"`
	State              string `json:"State,omitempty"`
	VolumeId           string `json:"VolumeId,omitempty"`
}

// implements the service definition of BsuToCreate
type BsuToCreate struct {
	DeleteOnVmDeletion *bool  `json:"DeleteOnVmDeletion,omitempty"`
	Iops               int64  `json:"Iops,omitempty"`
	SnapshotId         string `json:"SnapshotId,omitempty"`
	VolumeSize         int64  `json:"VolumeSize,omitempty"`
	VolumeType         string `json:"VolumeType,omitempty"`
}

// implements the service definition of BsuToUpdateVm
type BsuToUpdateVm struct {
	DeleteOnVmDeletion *bool  `json:"DeleteOnVmDeletion,omitempty"`
	VolumeId           string `json:"VolumeId,omitempty"`
}

// implements the service definition of CatalogAttribute
type CatalogAttribute struct {
	Key   string `json:"Key,omitempty"`
	Value string `json:"Value,omitempty"`
}

// implements the service definition of CatalogEntry
type CatalogEntry struct {
	CatalogAttributes []CatalogAttribute `json:"CatalogAttributes,omitempty"`
	EntryKey          string             `json:"EntryKey,omitempty"`
	EntryValue        string             `json:"EntryValue,omitempty"`
	ShortDescription  string             `json:"ShortDescription,omitempty"`
}

// implements the service definition of Catalog_0
type Catalog_0 struct {
	Domain           string `json:"Domain,omitempty"`
	Instance         string `json:"Instance,omitempty"`
	SourceRegionName string `json:"SourceRegionName,omitempty"`
	TargetRegionName string `json:"TargetRegionName,omitempty"`
	Version          string `json:"Version,omitempty"`
}

// implements the service definition of Catalog_1
type Catalog_1 struct {
	CatalogAttributes []CatalogAttribute `json:"CatalogAttributes,omitempty"`
	CatalogEntries    []CatalogEntry     `json:"CatalogEntries,omitempty"`
}

// implements the service definition of CheckSignatureRequest
type CheckSignatureRequest struct {
	ApiKeyId      string `json:"ApiKeyId,omitempty"`
	DryRun        bool   `json:"DryRun,omitempty"`
	RegionName    string `json:"RegionName,omitempty"`
	RequestDate   string `json:"RequestDate,omitempty"`
	Service       string `json:"Service,omitempty"`
	Signature     string `json:"Signature,omitempty"`
	SignedContent string `json:"SignedContent,omitempty"`
}

// implements the service definition of CheckSignatureResponse
type CheckSignatureResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ClientGateway
type ClientGateway struct {
	BgpAsn          int64         `json:"BgpAsn,omitempty"`
	ClientGatewayId string        `json:"ClientGatewayId,omitempty"`
	ConnectionType  string        `json:"ConnectionType,omitempty"`
	PublicIp        string        `json:"PublicIp,omitempty"`
	State           string        `json:"State,omitempty"`
	Tags            []ResourceTag `json:"Tags,omitempty"`
}

// implements the service definition of ConsumptionEntries
type ConsumptionEntries struct {
	Category         string `json:"Category,omitempty"`
	ConsumptionValue string `json:"ConsumptionValue,omitempty"`
	Entry            string `json:"Entry,omitempty"`
	ResourceType     string `json:"ResourceType,omitempty"`
	Service          string `json:"Service,omitempty"`
	ShortDescription string `json:"ShortDescription,omitempty"`
}

// implements the service definition of CopyAccountRequest
type CopyAccountRequest struct {
	DestinationRegionName string `json:"DestinationRegionName,omitempty"`
	DryRun                bool   `json:"DryRun,omitempty"`
	QuotaProfile          string `json:"QuotaProfile,omitempty"`
}

// implements the service definition of CopyAccountResponse
type CopyAccountResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of CreateAccountRequest
type CreateAccountRequest struct {
	AccountId     string   `json:"AccountId,omitempty"`
	ApiKeys       []ApiKey `json:"ApiKeys,omitempty"`
	City          string   `json:"City,omitempty"`
	CompanyName   string   `json:"CompanyName,omitempty"`
	Country       string   `json:"Country,omitempty"`
	CustomerId    string   `json:"CustomerId,omitempty"`
	DryRun        bool     `json:"DryRun,omitempty"`
	Email         string   `json:"Email,omitempty"`
	FirstName     string   `json:"FirstName,omitempty"`
	JobTitle      string   `json:"JobTitle,omitempty"`
	LastName      string   `json:"LastName,omitempty"`
	Mobile        string   `json:"Mobile,omitempty"`
	Password      string   `json:"Password,omitempty"`
	Phone         string   `json:"Phone,omitempty"`
	QuotaProfile  string   `json:"QuotaProfile,omitempty"`
	StateProvince string   `json:"StateProvince,omitempty"`
	VatNumber     string   `json:"VatNumber,omitempty"`
	ZipCode       string   `json:"ZipCode,omitempty"`
}

// implements the service definition of CreateAccountResponse
type CreateAccountResponse struct {
	Account         Account         `json:"Account,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of CreateApiKeyRequest
type CreateApiKeyRequest struct {
	ApiKeyId  string        `json:"ApiKeyId,omitempty"`
	DryRun    bool          `json:"DryRun,omitempty"`
	SecretKey string        `json:"SecretKey,omitempty"`
	Tags      []ResourceTag `json:"Tags,omitempty"`
	UserName  string        `json:"UserName,omitempty"`
}

// implements the service definition of CreateApiKeyResponse
type CreateApiKeyResponse struct {
	ApiKey          ApiKey          `json:"ApiKey,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of CreateClientGatewayRequest
type CreateClientGatewayRequest struct {
	BgpAsn         int64  `json:"BgpAsn,omitempty"`
	ConnectionType string `json:"ConnectionType,omitempty"`
	DryRun         bool   `json:"DryRun,omitempty"`
	PublicIp       string `json:"PublicIp,omitempty"`
}

// implements the service definition of CreateClientGatewayResponse
type CreateClientGatewayResponse struct {
	ClientGateway   ClientGateway   `json:"ClientGateway,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of CreateDhcpOptionsRequest
type CreateDhcpOptionsRequest struct {
	DomainName        string   `json:"DomainName,omitempty"`
	DomainNameServers []string `json:"DomainNameServers,omitempty"`
	DryRun            bool     `json:"DryRun,omitempty"`
	NtpServers        []string `json:"NtpServers,omitempty"`
}

// implements the service definition of CreateDhcpOptionsResponse
type CreateDhcpOptionsResponse struct {
	DhcpOptionsSet  DhcpOptionsSet  `json:"DhcpOptionsSet,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of CreateDirectLinkInterfaceRequest
type CreateDirectLinkInterfaceRequest struct {
	DirectLinkId        string              `json:"DirectLinkId,omitempty"`
	DirectLinkInterface DirectLinkInterface `json:"DirectLinkInterface,omitempty"`
	DryRun              bool                `json:"DryRun,omitempty"`
}

// implements the service definition of CreateDirectLinkInterfaceResponse
type CreateDirectLinkInterfaceResponse struct {
	DirectLinkInterface DirectLinkInterfaces `json:"DirectLinkInterface,omitempty"`
	ResponseContext     ResponseContext      `json:"ResponseContext,omitempty"`
}

// implements the service definition of CreateDirectLinkRequest
type CreateDirectLinkRequest struct {
	Bandwidth      string `json:"Bandwidth,omitempty"`
	DirectLinkName string `json:"DirectLinkName,omitempty"`
	DryRun         bool   `json:"DryRun,omitempty"`
	Location       string `json:"Location,omitempty"`
}

// implements the service definition of CreateDirectLinkResponse
type CreateDirectLinkResponse struct {
	DirectLink      DirectLink      `json:"DirectLink,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of CreateImageExportTaskRequest
type CreateImageExportTaskRequest struct {
	DryRun    bool      `json:"DryRun,omitempty"`
	ImageId   string    `json:"ImageId,omitempty"`
	OsuExport OsuExport `json:"OsuExport,omitempty"`
}

// implements the service definition of CreateImageExportTaskResponse
type CreateImageExportTaskResponse struct {
	ImageExportTask ImageExportTask `json:"ImageExportTask,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of CreateImageRequest
type CreateImageRequest struct {
	Architecture        string                    `json:"Architecture,omitempty"`
	BlockDeviceMappings []BlockDeviceMappingImage `json:"BlockDeviceMappings,omitempty"`
	Description         string                    `json:"Description,omitempty"`
	DryRun              bool                      `json:"DryRun,omitempty"`
	FileLocation        string                    `json:"FileLocation,omitempty"`
	ImageName           string                    `json:"ImageName,omitempty"`
	NoReboot            bool                      `json:"NoReboot,omitempty"`
	RootDeviceName      string                    `json:"RootDeviceName,omitempty"`
	SourceImageId       string                    `json:"SourceImageId,omitempty"`
	SourceRegionName    string                    `json:"SourceRegionName,omitempty"`
	VmId                string                    `json:"VmId,omitempty"`
}

// implements the service definition of CreateImageResponse
type CreateImageResponse struct {
	Image           Image           `json:"Image,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of CreateInternetServiceRequest
type CreateInternetServiceRequest struct {
	DryRun bool `json:"DryRun,omitempty"`
}

// implements the service definition of CreateInternetServiceResponse
type CreateInternetServiceResponse struct {
	InternetService InternetService `json:"InternetService,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of CreateKeypairRequest
type CreateKeypairRequest struct {
	DryRun      bool   `json:"DryRun,omitempty"`
	KeypairName string `json:"KeypairName,omitempty"`
	PublicKey   string `json:"PublicKey,omitempty"`
}

// implements the service definition of CreateKeypairResponse
type CreateKeypairResponse struct {
	Keypair         KeypairCreated  `json:"Keypair,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of CreateListenerRuleRequest
type CreateListenerRuleRequest struct {
	DryRun       bool              `json:"DryRun,omitempty"`
	Listener     LoadBalancerLight `json:"Listener,omitempty"`
	ListenerRule ListenerRule      `json:"ListenerRule,omitempty"`
	VmIds        []string          `json:"VmIds,omitempty"`
}

// implements the service definition of CreateListenerRuleResponse
type CreateListenerRuleResponse struct {
	ListenerId      string          `json:"ListenerId,omitempty"`
	ListenerRule    ListenerRule    `json:"ListenerRule,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	VmIds           []string        `json:"VmIds,omitempty"`
}

// implements the service definition of CreateLoadBalancerListenersRequest
type CreateLoadBalancerListenersRequest struct {
	DryRun           bool                  `json:"DryRun,omitempty"`
	Listeners        []ListenerForCreation `json:"Listeners,omitempty"`
	LoadBalancerName string                `json:"LoadBalancerName,omitempty"`
}

// implements the service definition of CreateLoadBalancerListenersResponse
type CreateLoadBalancerListenersResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of CreateLoadBalancerPolicyRequest
type CreateLoadBalancerPolicyRequest struct {
	CookieName       string `json:"CookieName,omitempty"`
	DryRun           bool   `json:"DryRun,omitempty"`
	LoadBalancerName string `json:"LoadBalancerName,omitempty"`
	PolicyName       string `json:"PolicyName,omitempty"`
	PolicyType       string `json:"PolicyType,omitempty"`
}

// implements the service definition of CreateLoadBalancerPolicyResponse
type CreateLoadBalancerPolicyResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of CreateLoadBalancerRequest
type CreateLoadBalancerRequest struct {
	DryRun           bool                  `json:"DryRun,omitempty"`
	Listeners        []ListenerForCreation `json:"Listeners,omitempty"`
	LoadBalancerName string                `json:"LoadBalancerName,omitempty"`
	LoadBalancerType string                `json:"LoadBalancerType,omitempty"`
	SecurityGroups   []string              `json:"SecurityGroups,omitempty"`
	Subnets          []string              `json:"Subnets,omitempty"`
	SubregionNames   []string              `json:"SubregionNames,omitempty"`
	Tags             []ResourceTag         `json:"Tags,omitempty"`
}

// implements the service definition of CreateLoadBalancerResponse
type CreateLoadBalancerResponse struct {
	LoadBalancer    LoadBalancer    `json:"LoadBalancer,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of CreateNatServiceRequest
type CreateNatServiceRequest struct {
	DryRun     bool   `json:"DryRun,omitempty"`
	PublicIpId string `json:"PublicIpId,omitempty"`
	SubnetId   string `json:"SubnetId,omitempty"`
}

// implements the service definition of CreateNatServiceResponse
type CreateNatServiceResponse struct {
	NatService      NatService      `json:"NatService,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of CreateNetAccessPointRequest
type CreateNetAccessPointRequest struct {
	DryRun         bool     `json:"DryRun,omitempty"`
	NetId          string   `json:"NetId,omitempty"`
	PrefixListName string   `json:"PrefixListName,omitempty"`
	RouteTableIds  []string `json:"RouteTableIds,omitempty"`
}

// implements the service definition of CreateNetAccessPointResponse
type CreateNetAccessPointResponse struct {
	NetAccessPoint  NetAccessPoint  `json:"NetAccessPoint,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of CreateNetPeeringRequest
type CreateNetPeeringRequest struct {
	AccepterNetId string `json:"AccepterNetId,omitempty"`
	DryRun        bool   `json:"DryRun,omitempty"`
	SourceNetId   string `json:"SourceNetId,omitempty"`
}

// implements the service definition of CreateNetPeeringResponse
type CreateNetPeeringResponse struct {
	NetPeering      NetPeering      `json:"NetPeering,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of CreateNetRequest
type CreateNetRequest struct {
	DryRun  bool   `json:"DryRun,omitempty"`
	IpRange string `json:"IpRange,omitempty"`
	Tenancy string `json:"Tenancy,omitempty"`
}

// implements the service definition of CreateNetResponse
type CreateNetResponse struct {
	Net             Net             `json:"Net,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of CreateNicRequest
type CreateNicRequest struct {
	Description      string           `json:"Description,omitempty"`
	DryRun           bool             `json:"DryRun,omitempty"`
	PrivateIps       []PrivateIpLight `json:"PrivateIps,omitempty"`
	SecurityGroupIds []string         `json:"SecurityGroupIds,omitempty"`
	SubnetId         string           `json:"SubnetId,omitempty"`
}

// implements the service definition of CreateNicResponse
type CreateNicResponse struct {
	Nic             Nic             `json:"Nic,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of CreatePolicyRequest
type CreatePolicyRequest struct {
	Description string `json:"Description,omitempty"`
	Document    string `json:"Document,omitempty"`
	DryRun      bool   `json:"DryRun,omitempty"`
	Path        string `json:"Path,omitempty"`
	PolicyName  string `json:"PolicyName,omitempty"`
}

// implements the service definition of CreatePolicyResponse
type CreatePolicyResponse struct {
	Policy          Policy          `json:"Policy,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of CreatePublicIpRequest
type CreatePublicIpRequest struct {
	DryRun bool `json:"DryRun,omitempty"`
}

// implements the service definition of CreatePublicIpResponse
type CreatePublicIpResponse struct {
	PublicIp        PublicIp        `json:"PublicIp,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of CreateRouteRequest
type CreateRouteRequest struct {
	DestinationIpRange string `json:"DestinationIpRange,omitempty"`
	DryRun             bool   `json:"DryRun,omitempty"`
	GatewayId          string `json:"GatewayId,omitempty"`
	NatServiceId       string `json:"NatServiceId,omitempty"`
	NetPeeringId       string `json:"NetPeeringId,omitempty"`
	NicId              string `json:"NicId,omitempty"`
	RouteTableId       string `json:"RouteTableId,omitempty"`
	VmId               string `json:"VmId,omitempty"`
}

// implements the service definition of CreateRouteResponse
type CreateRouteResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	Route           Route           `json:"Route,omitempty"`
}

// implements the service definition of CreateRouteTableRequest
type CreateRouteTableRequest struct {
	DryRun bool   `json:"DryRun,omitempty"`
	NetId  string `json:"NetId,omitempty"`
}

// implements the service definition of CreateRouteTableResponse
type CreateRouteTableResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	RouteTable      RouteTable      `json:"RouteTable,omitempty"`
}

// implements the service definition of CreateSecurityGroupRequest
type CreateSecurityGroupRequest struct {
	Description       string `json:"Description,omitempty"`
	DryRun            bool   `json:"DryRun,omitempty"`
	NetId             string `json:"NetId,omitempty"`
	SecurityGroupName string `json:"SecurityGroupName,omitempty"`
}

// implements the service definition of CreateSecurityGroupResponse
type CreateSecurityGroupResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	SecurityGroup   SecurityGroup   `json:"SecurityGroup,omitempty"`
}

// implements the service definition of CreateSecurityGroupRuleRequest
type CreateSecurityGroupRuleRequest struct {
	DryRun                       bool                `json:"DryRun,omitempty"`
	Flow                         string              `json:"Flow,omitempty"`
	FromPortRange                int64               `json:"FromPortRange,omitempty"`
	IpProtocol                   string              `json:"IpProtocol,omitempty"`
	IpRange                      string              `json:"IpRange,omitempty"`
	Rules                        []SecurityGroupRule `json:"Rules,omitempty"`
	SecurityGroupAccountIdToLink string              `json:"SecurityGroupAccountIdToLink,omitempty"`
	SecurityGroupId              string              `json:"SecurityGroupId,omitempty"`
	SecurityGroupNameToLink      string              `json:"SecurityGroupNameToLink,omitempty"`
	ToPortRange                  int64               `json:"ToPortRange,omitempty"`
}

// implements the service definition of CreateSecurityGroupRuleResponse
type CreateSecurityGroupRuleResponse struct {
	ResponseContext   ResponseContext   `json:"ResponseContext,omitempty"`
	SecurityGroupRule SecurityGroupRule `json:"SecurityGroupRule,omitempty"`
}

// implements the service definition of CreateServerCertificateRequest
type CreateServerCertificateRequest struct {
	DryRun                 bool   `json:"DryRun,omitempty"`
	PrivateKey             string `json:"PrivateKey,omitempty"`
	ServerCertificateBody  string `json:"ServerCertificateBody,omitempty"`
	ServerCertificateChain string `json:"ServerCertificateChain,omitempty"`
	ServerCertificateName  string `json:"ServerCertificateName,omitempty"`
	ServerCertificatePath  string `json:"ServerCertificatePath,omitempty"`
}

// implements the service definition of CreateServerCertificateResponse
type CreateServerCertificateResponse struct {
	ResponseContext   ResponseContext   `json:"ResponseContext,omitempty"`
	ServerCertificate ServerCertificate `json:"ServerCertificate,omitempty"`
}

// implements the service definition of CreateSnapshotExportTaskRequest
type CreateSnapshotExportTaskRequest struct {
	DryRun     bool      `json:"DryRun,omitempty"`
	OsuExport  OsuExport `json:"OsuExport,omitempty"`
	SnapshotId string    `json:"SnapshotId,omitempty"`
}

// implements the service definition of CreateSnapshotExportTaskResponse
type CreateSnapshotExportTaskResponse struct {
	ResponseContext    ResponseContext    `json:"ResponseContext,omitempty"`
	SnapshotExportTask SnapshotExportTask `json:"SnapshotExportTask,omitempty"`
}

// implements the service definition of CreateSnapshotRequest
type CreateSnapshotRequest struct {
	Description      string `json:"Description,omitempty"`
	DryRun           bool   `json:"DryRun,omitempty"`
	FileLocation     string `json:"FileLocation,omitempty"`
	SnapshotSize     int64  `json:"SnapshotSize,omitempty"`
	SourceRegionName string `json:"SourceRegionName,omitempty"`
	SourceSnapshotId string `json:"SourceSnapshotId,omitempty"`
	VolumeId         string `json:"VolumeId,omitempty"`
}

// implements the service definition of CreateSnapshotResponse
type CreateSnapshotResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	Snapshot        Snapshot        `json:"Snapshot,omitempty"`
}

// implements the service definition of CreateSubnetRequest
type CreateSubnetRequest struct {
	DryRun        bool   `json:"DryRun,omitempty"`
	IpRange       string `json:"IpRange,omitempty"`
	NetId         string `json:"NetId,omitempty"`
	SubregionName string `json:"SubregionName,omitempty"`
}

// implements the service definition of CreateSubnetResponse
type CreateSubnetResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	Subnet          Subnet          `json:"Subnet,omitempty"`
}

// implements the service definition of CreateTagsRequest
type CreateTagsRequest struct {
	DryRun      bool          `json:"DryRun,omitempty"`
	ResourceIds []string      `json:"ResourceIds,omitempty"`
	Tags        []ResourceTag `json:"Tags,omitempty"`
}

// implements the service definition of CreateTagsResponse
type CreateTagsResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of CreateUserGroupRequest
type CreateUserGroupRequest struct {
	DryRun        bool   `json:"DryRun,omitempty"`
	Path          string `json:"Path,omitempty"`
	UserGroupName string `json:"UserGroupName,omitempty"`
}

// implements the service definition of CreateUserGroupResponse
type CreateUserGroupResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	UserGroup       UserGroup       `json:"UserGroup,omitempty"`
}

// implements the service definition of CreateUserRequest
type CreateUserRequest struct {
	DryRun   bool   `json:"DryRun,omitempty"`
	Path     string `json:"Path,omitempty"`
	UserName string `json:"UserName,omitempty"`
}

// implements the service definition of CreateUserResponse
type CreateUserResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	User            User            `json:"User,omitempty"`
}

// implements the service definition of CreateVirtualGatewayRequest
type CreateVirtualGatewayRequest struct {
	ConnectionType string `json:"ConnectionType,omitempty"`
	DryRun         bool   `json:"DryRun,omitempty"`
}

// implements the service definition of CreateVirtualGatewayResponse
type CreateVirtualGatewayResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	VirtualGateway  VirtualGateway  `json:"VirtualGateway,omitempty"`
}

// implements the service definition of CreateVmsRequest
type CreateVmsRequest struct {
	BlockDeviceMappings         []BlockDeviceMappingVmCreation `json:"BlockDeviceMappings,omitempty"`
	BsuOptimized                bool                           `json:"BsuOptimized,omitempty"`
	ClientToken                 string                         `json:"ClientToken,omitempty"`
	DeletionProtection          bool                           `json:"DeletionProtection,omitempty"`
	DryRun                      bool                           `json:"DryRun,omitempty"`
	ImageId                     string                         `json:"ImageId,omitempty"`
	KeypairName                 string                         `json:"KeypairName,omitempty"`
	MaxVmsCount                 int64                          `json:"MaxVmsCount,omitempty"`
	MinVmsCount                 int64                          `json:"MinVmsCount,omitempty"`
	Nics                        []NicForVmCreation             `json:"Nics,omitempty"`
	Placement                   Placement                      `json:"Placement,omitempty"`
	PrivateIps                  []string                       `json:"PrivateIps,omitempty"`
	SecurityGroupIds            []string                       `json:"SecurityGroupIds,omitempty"`
	SecurityGroups              []string                       `json:"SecurityGroups,omitempty"`
	SubnetId                    string                         `json:"SubnetId,omitempty"`
	UserData                    string                         `json:"UserData,omitempty"`
	VmInitiatedShutdownBehavior string                         `json:"VmInitiatedShutdownBehavior,omitempty"`
	VmType                      string                         `json:"VmType,omitempty"`
}

// implements the service definition of CreateVmsResponse
type CreateVmsResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	Vms             []Vm            `json:"Vms,omitempty"`
}

// implements the service definition of CreateVolumeRequest
type CreateVolumeRequest struct {
	DryRun        bool   `json:"DryRun,omitempty"`
	Iops          int64  `json:"Iops,omitempty"`
	Size          int64  `json:"Size,omitempty"`
	SnapshotId    string `json:"SnapshotId,omitempty"`
	SubregionName string `json:"SubregionName,omitempty"`
	VolumeType    string `json:"VolumeType,omitempty"`
}

// implements the service definition of CreateVolumeResponse
type CreateVolumeResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	Volume          Volume          `json:"Volume,omitempty"`
}

// implements the service definition of CreateVpnConnectionRequest
type CreateVpnConnectionRequest struct {
	ClientGatewayId  string `json:"ClientGatewayId,omitempty"`
	ConnectionType   string `json:"ConnectionType,omitempty"`
	DryRun           bool   `json:"DryRun,omitempty"`
	StaticRoutesOnly bool   `json:"StaticRoutesOnly,omitempty"`
	VirtualGatewayId string `json:"VirtualGatewayId,omitempty"`
}

// implements the service definition of CreateVpnConnectionResponse
type CreateVpnConnectionResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	VpnConnection   VpnConnection   `json:"VpnConnection,omitempty"`
}

// implements the service definition of CreateVpnConnectionRouteRequest
type CreateVpnConnectionRouteRequest struct {
	DestinationIpRange string `json:"DestinationIpRange,omitempty"`
	DryRun             bool   `json:"DryRun,omitempty"`
	VpnConnectionId    string `json:"VpnConnectionId,omitempty"`
}

// implements the service definition of CreateVpnConnectionRouteResponse
type CreateVpnConnectionRouteResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteApiKeyRequest
type DeleteApiKeyRequest struct {
	ApiKeyId string `json:"ApiKeyId,omitempty"`
	DryRun   bool   `json:"DryRun,omitempty"`
}

// implements the service definition of DeleteApiKeyResponse
type DeleteApiKeyResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteClientGatewayRequest
type DeleteClientGatewayRequest struct {
	ClientGatewayId string `json:"ClientGatewayId,omitempty"`
	DryRun          bool   `json:"DryRun,omitempty"`
}

// implements the service definition of DeleteClientGatewayResponse
type DeleteClientGatewayResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteDhcpOptionsRequest
type DeleteDhcpOptionsRequest struct {
	DhcpOptionsSetId string `json:"DhcpOptionsSetId,omitempty"`
	DryRun           bool   `json:"DryRun,omitempty"`
}

// implements the service definition of DeleteDhcpOptionsResponse
type DeleteDhcpOptionsResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteDirectLinkInterfaceRequest
type DeleteDirectLinkInterfaceRequest struct {
	DirectLinkInterfaceId string `json:"DirectLinkInterfaceId,omitempty"`
	DryRun                bool   `json:"DryRun,omitempty"`
}

// implements the service definition of DeleteDirectLinkInterfaceResponse
type DeleteDirectLinkInterfaceResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteDirectLinkRequest
type DeleteDirectLinkRequest struct {
	DirectLinkId string `json:"DirectLinkId,omitempty"`
	DryRun       bool   `json:"DryRun,omitempty"`
}

// implements the service definition of DeleteDirectLinkResponse
type DeleteDirectLinkResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteExportTaskRequest
type DeleteExportTaskRequest struct {
	DryRun       bool   `json:"DryRun,omitempty"`
	ExportTaskId string `json:"ExportTaskId,omitempty"`
}

// implements the service definition of DeleteExportTaskResponse
type DeleteExportTaskResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteImageRequest
type DeleteImageRequest struct {
	DryRun  bool   `json:"DryRun,omitempty"`
	ImageId string `json:"ImageId,omitempty"`
}

// implements the service definition of DeleteImageResponse
type DeleteImageResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteInternetServiceRequest
type DeleteInternetServiceRequest struct {
	DryRun            bool   `json:"DryRun,omitempty"`
	InternetServiceId string `json:"InternetServiceId,omitempty"`
}

// implements the service definition of DeleteInternetServiceResponse
type DeleteInternetServiceResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteKeypairRequest
type DeleteKeypairRequest struct {
	DryRun      bool   `json:"DryRun,omitempty"`
	KeypairName string `json:"KeypairName,omitempty"`
}

// implements the service definition of DeleteKeypairResponse
type DeleteKeypairResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteListenerRuleRequest
type DeleteListenerRuleRequest struct {
	DryRun           bool   `json:"DryRun,omitempty"`
	ListenerRuleName string `json:"ListenerRuleName,omitempty"`
}

// implements the service definition of DeleteListenerRuleResponse
type DeleteListenerRuleResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteLoadBalancerListenersRequest
type DeleteLoadBalancerListenersRequest struct {
	DryRun            bool    `json:"DryRun,omitempty"`
	LoadBalancerName  string  `json:"LoadBalancerName,omitempty"`
	LoadBalancerPorts []int64 `json:"LoadBalancerPorts,omitempty"`
}

// implements the service definition of DeleteLoadBalancerListenersResponse
type DeleteLoadBalancerListenersResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteLoadBalancerPolicyRequest
type DeleteLoadBalancerPolicyRequest struct {
	DryRun           bool   `json:"DryRun,omitempty"`
	LoadBalancerName string `json:"LoadBalancerName,omitempty"`
	PolicyName       string `json:"PolicyName,omitempty"`
}

// implements the service definition of DeleteLoadBalancerPolicyResponse
type DeleteLoadBalancerPolicyResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteLoadBalancerRequest
type DeleteLoadBalancerRequest struct {
	DryRun           bool   `json:"DryRun,omitempty"`
	LoadBalancerName string `json:"LoadBalancerName,omitempty"`
}

// implements the service definition of DeleteLoadBalancerResponse
type DeleteLoadBalancerResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteNatServiceRequest
type DeleteNatServiceRequest struct {
	DryRun       bool   `json:"DryRun,omitempty"`
	NatServiceId string `json:"NatServiceId,omitempty"`
}

// implements the service definition of DeleteNatServiceResponse
type DeleteNatServiceResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteNetAccessPointsRequest
type DeleteNetAccessPointsRequest struct {
	DryRun            bool     `json:"DryRun,omitempty"`
	NetAccessPointIds []string `json:"NetAccessPointIds,omitempty"`
}

// implements the service definition of DeleteNetAccessPointsResponse
type DeleteNetAccessPointsResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteNetPeeringRequest
type DeleteNetPeeringRequest struct {
	DryRun       bool   `json:"DryRun,omitempty"`
	NetPeeringId string `json:"NetPeeringId,omitempty"`
}

// implements the service definition of DeleteNetPeeringResponse
type DeleteNetPeeringResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteNetRequest
type DeleteNetRequest struct {
	DryRun bool   `json:"DryRun,omitempty"`
	NetId  string `json:"NetId,omitempty"`
}

// implements the service definition of DeleteNetResponse
type DeleteNetResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteNicRequest
type DeleteNicRequest struct {
	DryRun bool   `json:"DryRun,omitempty"`
	NicId  string `json:"NicId,omitempty"`
}

// implements the service definition of DeleteNicResponse
type DeleteNicResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeletePolicyRequest
type DeletePolicyRequest struct {
	DryRun   bool   `json:"DryRun,omitempty"`
	PolicyId string `json:"PolicyId,omitempty"`
}

// implements the service definition of DeletePolicyResponse
type DeletePolicyResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeletePublicIpRequest
type DeletePublicIpRequest struct {
	DryRun     bool   `json:"DryRun,omitempty"`
	PublicIp   string `json:"PublicIp,omitempty"`
	PublicIpId string `json:"PublicIpId,omitempty"`
}

// implements the service definition of DeletePublicIpResponse
type DeletePublicIpResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteRouteRequest
type DeleteRouteRequest struct {
	DestinationIpRange string `json:"DestinationIpRange,omitempty"`
	DryRun             bool   `json:"DryRun,omitempty"`
	RouteTableId       string `json:"RouteTableId,omitempty"`
}

// implements the service definition of DeleteRouteResponse
type DeleteRouteResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteRouteTableRequest
type DeleteRouteTableRequest struct {
	DryRun       bool   `json:"DryRun,omitempty"`
	RouteTableId string `json:"RouteTableId,omitempty"`
}

// implements the service definition of DeleteRouteTableResponse
type DeleteRouteTableResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteSecurityGroupRequest
type DeleteSecurityGroupRequest struct {
	DryRun            bool   `json:"DryRun,omitempty"`
	SecurityGroupId   string `json:"SecurityGroupId,omitempty"`
	SecurityGroupName string `json:"SecurityGroupName,omitempty"`
}

// implements the service definition of DeleteSecurityGroupResponse
type DeleteSecurityGroupResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteSecurityGroupRuleRequest
type DeleteSecurityGroupRuleRequest struct {
	DryRun                         bool                `json:"DryRun,omitempty"`
	Flow                           string              `json:"Flow,omitempty"`
	FromPortRange                  int64               `json:"FromPortRange,omitempty"`
	IpProtocol                     string              `json:"IpProtocol,omitempty"`
	IpRange                        string              `json:"IpRange,omitempty"`
	Rules                          []SecurityGroupRule `json:"Rules,omitempty"`
	SecurityGroupAccountIdToUnlink string              `json:"SecurityGroupAccountIdToUnlink,omitempty"`
	SecurityGroupId                string              `json:"SecurityGroupId,omitempty"`
	SecurityGroupNameToUnlink      string              `json:"SecurityGroupNameToUnlink,omitempty"`
	ToPortRange                    int64               `json:"ToPortRange,omitempty"`
}

// implements the service definition of DeleteSecurityGroupRuleResponse
type DeleteSecurityGroupRuleResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteServerCertificateRequest
type DeleteServerCertificateRequest struct {
	DryRun                bool   `json:"DryRun,omitempty"`
	ServerCertificateName string `json:"ServerCertificateName,omitempty"`
}

// implements the service definition of DeleteServerCertificateResponse
type DeleteServerCertificateResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteSnapshotRequest
type DeleteSnapshotRequest struct {
	DryRun     bool   `json:"DryRun,omitempty"`
	SnapshotId string `json:"SnapshotId,omitempty"`
}

// implements the service definition of DeleteSnapshotResponse
type DeleteSnapshotResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteSubnetRequest
type DeleteSubnetRequest struct {
	DryRun   bool   `json:"DryRun,omitempty"`
	SubnetId string `json:"SubnetId,omitempty"`
}

// implements the service definition of DeleteSubnetResponse
type DeleteSubnetResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteTagsRequest
type DeleteTagsRequest struct {
	DryRun      bool          `json:"DryRun,omitempty"`
	ResourceIds []string      `json:"ResourceIds,omitempty"`
	Tags        []ResourceTag `json:"Tags,omitempty"`
}

// implements the service definition of DeleteTagsResponse
type DeleteTagsResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteUserGroupRequest
type DeleteUserGroupRequest struct {
	DryRun        bool   `json:"DryRun,omitempty"`
	UserGroupName string `json:"UserGroupName,omitempty"`
}

// implements the service definition of DeleteUserGroupResponse
type DeleteUserGroupResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteUserRequest
type DeleteUserRequest struct {
	DryRun   bool   `json:"DryRun,omitempty"`
	UserName string `json:"UserName,omitempty"`
}

// implements the service definition of DeleteUserResponse
type DeleteUserResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteVirtualGatewayRequest
type DeleteVirtualGatewayRequest struct {
	DryRun           bool   `json:"DryRun,omitempty"`
	VirtualGatewayId string `json:"VirtualGatewayId,omitempty"`
}

// implements the service definition of DeleteVirtualGatewayResponse
type DeleteVirtualGatewayResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteVmsRequest
type DeleteVmsRequest struct {
	DryRun bool     `json:"DryRun,omitempty"`
	VmIds  []string `json:"VmIds,omitempty"`
}

// implements the service definition of DeleteVmsResponse
type DeleteVmsResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	Vms             []VmState       `json:"Vms,omitempty"`
}

// implements the service definition of DeleteVolumeRequest
type DeleteVolumeRequest struct {
	DryRun   bool   `json:"DryRun,omitempty"`
	VolumeId string `json:"VolumeId,omitempty"`
}

// implements the service definition of DeleteVolumeResponse
type DeleteVolumeResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteVpnConnectionRequest
type DeleteVpnConnectionRequest struct {
	DryRun          bool   `json:"DryRun,omitempty"`
	VpnConnectionId string `json:"VpnConnectionId,omitempty"`
}

// implements the service definition of DeleteVpnConnectionResponse
type DeleteVpnConnectionResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeleteVpnConnectionRouteRequest
type DeleteVpnConnectionRouteRequest struct {
	DestinationIpRange string `json:"DestinationIpRange,omitempty"`
	DryRun             bool   `json:"DryRun,omitempty"`
	VpnConnectionId    string `json:"VpnConnectionId,omitempty"`
}

// implements the service definition of DeleteVpnConnectionRouteResponse
type DeleteVpnConnectionRouteResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeregisterUserInUserGroupRequest
type DeregisterUserInUserGroupRequest struct {
	DryRun        bool   `json:"DryRun,omitempty"`
	UserGroupName string `json:"UserGroupName,omitempty"`
	UserName      string `json:"UserName,omitempty"`
}

// implements the service definition of DeregisterUserInUserGroupResponse
type DeregisterUserInUserGroupResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DeregisterVmsInLoadBalancerRequest
type DeregisterVmsInLoadBalancerRequest struct {
	BackendVmsIds    []string `json:"BackendVmsIds,omitempty"`
	DryRun           bool     `json:"DryRun,omitempty"`
	LoadBalancerName string   `json:"LoadBalancerName,omitempty"`
}

// implements the service definition of DeregisterVmsInLoadBalancerResponse
type DeregisterVmsInLoadBalancerResponse struct {
	BackendVmsIds   []string        `json:"BackendVmsIds,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of DhcpOptionsSet
type DhcpOptionsSet struct {
	Default           bool          `json:"Default,omitempty"`
	DhcpOptionsName   string        `json:"DhcpOptionsName,omitempty"`
	DhcpOptionsSetId  string        `json:"DhcpOptionsSetId,omitempty"`
	DomainName        string        `json:"DomainName,omitempty"`
	DomainNameServers []string      `json:"DomainNameServers,omitempty"`
	NtpServers        []string      `json:"NtpServers,omitempty"`
	Tags              []ResourceTag `json:"Tags,omitempty"`
}

// implements the service definition of DirectLink
type DirectLink struct {
	AccountId      string `json:"AccountId,omitempty"`
	Bandwidth      string `json:"Bandwidth,omitempty"`
	DirectLinkId   string `json:"DirectLinkId,omitempty"`
	DirectLinkName string `json:"DirectLinkName,omitempty"`
	Location       string `json:"Location,omitempty"`
	RegionName     string `json:"RegionName,omitempty"`
	State          string `json:"State,omitempty"`
}

// implements the service definition of DirectLinkInterface
type DirectLinkInterface struct {
	BgpAsn                  int64  `json:"BgpAsn,omitempty"`
	BgpKey                  string `json:"BgpKey,omitempty"`
	ClientPrivateIp         string `json:"ClientPrivateIp,omitempty"`
	DirectLinkInterfaceName string `json:"DirectLinkInterfaceName,omitempty"`
	OutscalePrivateIp       string `json:"OutscalePrivateIp,omitempty"`
	VirtualGatewayId        string `json:"VirtualGatewayId,omitempty"`
	Vlan                    int64  `json:"Vlan,omitempty"`
}

// implements the service definition of DirectLinkInterfaces
type DirectLinkInterfaces struct {
	AccountId               string `json:"AccountId,omitempty"`
	BgpAsn                  int64  `json:"BgpAsn,omitempty"`
	BgpKey                  string `json:"BgpKey,omitempty"`
	ClientPrivateIp         string `json:"ClientPrivateIp,omitempty"`
	DirectLinkId            string `json:"DirectLinkId,omitempty"`
	DirectLinkInterfaceId   string `json:"DirectLinkInterfaceId,omitempty"`
	DirectLinkInterfaceName string `json:"DirectLinkInterfaceName,omitempty"`
	InterfaceType           string `json:"InterfaceType,omitempty"`
	Location                string `json:"Location,omitempty"`
	OutscalePrivateIp       string `json:"OutscalePrivateIp,omitempty"`
	State                   string `json:"State,omitempty"`
	VirtualGatewayId        string `json:"VirtualGatewayId,omitempty"`
	Vlan                    int64  `json:"Vlan,omitempty"`
}

// implements the service definition of ErrorResponse
type ErrorResponse struct {
	Errors          []Errors        `json:"Errors,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of Errors
type Errors struct {
	Code    string `json:"Code,omitempty"`
	Details string `json:"Details,omitempty"`
	Type    string `json:"Type,omitempty"`
}

// implements the service definition of FiltersApiLog
type FiltersApiLog struct {
	QueryAccessKeys     []string `json:"QueryAccessKeys,omitempty"`
	QueryApiNames       []string `json:"QueryApiNames,omitempty"`
	QueryCallNames      []string `json:"QueryCallNames,omitempty"`
	QueryDateAfter      string   `json:"QueryDateAfter,omitempty"`
	QueryDateBefore     string   `json:"QueryDateBefore,omitempty"`
	QueryIpAddresses    []string `json:"QueryIpAddresses,omitempty"`
	QueryUserAgents     []string `json:"QueryUserAgents,omitempty"`
	ResponseIds         []string `json:"ResponseIds,omitempty"`
	ResponseStatusCodes []int64  `json:"ResponseStatusCodes,omitempty"`
}

// implements the service definition of FiltersDhcpOptions
type FiltersDhcpOptions struct {
	Defaults          []string `json:"Defaults,omitempty"`
	DhcpOptionsSetIds []string `json:"DhcpOptionsSetIds,omitempty"`
	DomainNameServers []string `json:"DomainNameServers,omitempty"`
	DomainNames       []string `json:"DomainNames,omitempty"`
	NtpServers        []string `json:"NtpServers,omitempty"`
	TagKeys           []string `json:"TagKeys,omitempty"`
	TagValues         []string `json:"TagValues,omitempty"`
	Tags              []string `json:"Tags,omitempty"`
}

// implements the service definition of FiltersExportTask
type FiltersExportTask struct {
	TaskIds []string `json:"TaskIds,omitempty"`
}

// implements the service definition of FiltersImage
type FiltersImage struct {
	AccountAliases                          []string `json:"AccountAliases,omitempty"`
	AccountIds                              []string `json:"AccountIds,omitempty"`
	Architectures                           []string `json:"Architectures,omitempty"`
	BlockDeviceMappingDeleteOnVmTermination bool     `json:"BlockDeviceMappingDeleteOnVmTermination,omitempty"`
	BlockDeviceMappingDeviceNames           []string `json:"BlockDeviceMappingDeviceNames,omitempty"`
	BlockDeviceMappingSnapshotIds           []string `json:"BlockDeviceMappingSnapshotIds,omitempty"`
	BlockDeviceMappingVolumeSize            []int64  `json:"BlockDeviceMappingVolumeSize,omitempty"`
	BlockDeviceMappingVolumeType            []string `json:"BlockDeviceMappingVolumeType,omitempty"`
	Descriptions                            []string `json:"Descriptions,omitempty"`
	Hypervisors                             []string `json:"Hypervisors,omitempty"`
	ImageIds                                []string `json:"ImageIds,omitempty"`
	ImageNames                              []string `json:"ImageNames,omitempty"`
	ImageTypes                              []string `json:"ImageTypes,omitempty"`
	KernelIds                               []string `json:"KernelIds,omitempty"`
	ManifestLocation                        []string `json:"ManifestLocation,omitempty"`
	PermissionsToLaunchAccountIds           []string `json:"PermissionsToLaunchAccountIds,omitempty"`
	PermissionsToLaunchGlobalPermission     bool     `json:"PermissionsToLaunchGlobalPermission,omitempty"`
	ProductCodes                            []string `json:"ProductCodes,omitempty"`
	RamDiskIds                              []string `json:"RamDiskIds,omitempty"`
	RootDeviceNames                         []string `json:"RootDeviceNames,omitempty"`
	RootDeviceTypes                         []string `json:"RootDeviceTypes,omitempty"`
	States                                  []string `json:"States,omitempty"`
	System                                  []string `json:"System,omitempty"`
	TagKeys                                 []string `json:"TagKeys,omitempty"`
	TagValues                               []string `json:"TagValues,omitempty"`
	Tags                                    []string `json:"Tags,omitempty"`
	VirtualizationTypes                     []string `json:"VirtualizationTypes,omitempty"`
}

// implements the service definition of FiltersInternetService
type FiltersInternetService struct {
	InternetServiceIds []string `json:"InternetServiceIds,omitempty"`
}

// implements the service definition of FiltersKeypair
type FiltersKeypair struct {
	KeypairFingerprints []string `json:"KeypairFingerprints,omitempty"`
	KeypairNames        []string `json:"KeypairNames,omitempty"`
}

// implements the service definition of FiltersLoadBalancer
type FiltersLoadBalancer struct {
	LoadBalancerNames []string `json:"LoadBalancerNames,omitempty"`
}

// implements the service definition of FiltersNatService
type FiltersNatService struct {
	NatServiceIds []string `json:"NatServiceIds,omitempty"`
	NetIds        []string `json:"NetIds,omitempty"`
	States        []string `json:"States,omitempty"`
	SubnetIds     []string `json:"SubnetIds,omitempty"`
	TagKeys       []string `json:"TagKeys,omitempty"`
	TagValues     []string `json:"TagValues,omitempty"`
	Tags          []string `json:"Tags,omitempty"`
}

// implements the service definition of FiltersNet
type FiltersNet struct {
	DhcpOptionsSetIds []string `json:"DhcpOptionsSetIds,omitempty"`
	IpRanges          []string `json:"IpRanges,omitempty"`
	IsDefault         bool     `json:"IsDefault,omitempty"`
	NetIds            []string `json:"NetIds,omitempty"`
	States            []string `json:"States,omitempty"`
	TagKeys           []string `json:"TagKeys,omitempty"`
	TagValues         []string `json:"TagValues,omitempty"`
	Tags              []string `json:"Tags,omitempty"`
}

// implements the service definition of FiltersNetPeering
type FiltersNetPeering struct {
	AccepterNetAccountIds []string `json:"AccepterNetAccountIds,omitempty"`
	AccepterNetIpRanges   []string `json:"AccepterNetIpRanges,omitempty"`
	AccepterNetNetIds     []string `json:"AccepterNetNetIds,omitempty"`
	NetPeeringIds         []string `json:"NetPeeringIds,omitempty"`
	SourceNetAccountIds   []string `json:"SourceNetAccountIds,omitempty"`
	SourceNetIpRanges     []string `json:"SourceNetIpRanges,omitempty"`
	SourceNetNetIds       []string `json:"SourceNetNetIds,omitempty"`
	StateMessages         []string `json:"StateMessages,omitempty"`
	StateNames            []string `json:"StateNames,omitempty"`
	TagKeys               []string `json:"TagKeys,omitempty"`
	TagValues             []string `json:"TagValues,omitempty"`
	Tags                  []string `json:"Tags,omitempty"`
}

// implements the service definition of FiltersNic
type FiltersNic struct {
	AccountIds                       []string `json:"AccountIds,omitempty"`
	ActivatedChecks                  []string `json:"ActivatedChecks,omitempty"`
	Descriptions                     []string `json:"Descriptions,omitempty"`
	LinkNicDeleteOnVmDeletion        bool     `json:"LinkNicDeleteOnVmDeletion,omitempty"`
	LinkNicLinkDates                 []string `json:"LinkNicLinkDates,omitempty"`
	LinkNicLinkNicIds                []string `json:"LinkNicLinkNicIds,omitempty"`
	LinkNicSortNumbers               []int64  `json:"LinkNicSortNumbers,omitempty"`
	LinkNicStates                    []string `json:"LinkNicStates,omitempty"`
	LinkNicVmAccountIds              []string `json:"LinkNicVmAccountIds,omitempty"`
	LinkNicVmIds                     []string `json:"LinkNicVmIds,omitempty"`
	LinkPublicIpAccountIds           []string `json:"LinkPublicIpAccountIds,omitempty"`
	LinkPublicIpLinkPublicIpIds      []string `json:"LinkPublicIpLinkPublicIpIds,omitempty"`
	LinkPublicIpPublicIpIds          []string `json:"LinkPublicIpPublicIpIds,omitempty"`
	LinkPublicIpPublicIps            []string `json:"LinkPublicIpPublicIps,omitempty"`
	MacAddresses                     []string `json:"MacAddresses,omitempty"`
	NetIds                           []string `json:"NetIds,omitempty"`
	NicIds                           []string `json:"NicIds,omitempty"`
	PrivateDnsNames                  []string `json:"PrivateDnsNames,omitempty"`
	PrivateIpsLinkPublicIpAccountIds []string `json:"PrivateIpsLinkPublicIpAccountIds,omitempty"`
	PrivateIpsLinkPublicIpPublicIps  []string `json:"PrivateIpsLinkPublicIpPublicIps,omitempty"`
	PrivateIpsPrimaryIp              bool     `json:"PrivateIpsPrimaryIp,omitempty"`
	PrivateIpsPrivateIps             []string `json:"PrivateIpsPrivateIps,omitempty"`
	SecurityGroupIds                 []string `json:"SecurityGroupIds,omitempty"`
	SecurityGroupNames               []string `json:"SecurityGroupNames,omitempty"`
	States                           []string `json:"States,omitempty"`
	SubnetIds                        []string `json:"SubnetIds,omitempty"`
	SubregionNames                   []string `json:"SubregionNames,omitempty"`
}

// implements the service definition of FiltersOldFormat
type FiltersOldFormat struct {
	Name   string   `json:"Name,omitempty"`
	Values []string `json:"Values,omitempty"`
}

// implements the service definition of FiltersPublicIp
type FiltersPublicIp struct {
	LinkPublicIpIds []string `json:"LinkPublicIpIds,omitempty"`
	NicAccountIds   []string `json:"NicAccountIds,omitempty"`
	NicIds          []string `json:"NicIds,omitempty"`
	Placements      []string `json:"Placements,omitempty"`
	PrivateIps      []string `json:"PrivateIps,omitempty"`
	PublicIpIds     []string `json:"PublicIpIds,omitempty"`
	PublicIps       []string `json:"PublicIps,omitempty"`
	VmIds           []string `json:"VmIds,omitempty"`
}

// implements the service definition of FiltersRouteTable
type FiltersRouteTable struct {
	LinkRouteTableIds               []string `json:"LinkRouteTableIds,omitempty"`
	LinkRouteTableLinkRouteTableIds []string `json:"LinkRouteTableLinkRouteTableIds,omitempty"`
	LinkRouteTableMain              bool     `json:"LinkRouteTableMain,omitempty"`
	LinkSubnetIds                   []string `json:"LinkSubnetIds,omitempty"`
	NetIds                          []string `json:"NetIds,omitempty"`
	RouteCreationMethods            []string `json:"RouteCreationMethods,omitempty"`
	RouteDestinationIpRanges        []string `json:"RouteDestinationIpRanges,omitempty"`
	RouteDestinationPrefixListIds   []string `json:"RouteDestinationPrefixListIds,omitempty"`
	RouteGatewayIds                 []string `json:"RouteGatewayIds,omitempty"`
	RouteNatServiceIds              []string `json:"RouteNatServiceIds,omitempty"`
	RouteNetPeeringIds              []string `json:"RouteNetPeeringIds,omitempty"`
	RouteStates                     []string `json:"RouteStates,omitempty"`
	RouteTableIds                   []string `json:"RouteTableIds,omitempty"`
	RouteVmIds                      []string `json:"RouteVmIds,omitempty"`
	TagKeys                         []string `json:"TagKeys,omitempty"`
	TagValues                       []string `json:"TagValues,omitempty"`
	Tags                            []string `json:"Tags,omitempty"`
}

// implements the service definition of FiltersSecurityGroup
type FiltersSecurityGroup struct {
	AccountIds                     []string `json:"AccountIds,omitempty"`
	Descriptions                   []string `json:"Descriptions,omitempty"`
	InboundRuleAccountIds          []string `json:"InboundRuleAccountIds,omitempty"`
	InboundRuleFromPortRanges      []int64  `json:"InboundRuleFromPortRanges,omitempty"`
	InboundRuleIpRanges            []string `json:"InboundRuleIpRanges,omitempty"`
	InboundRuleProtocols           []string `json:"InboundRuleProtocols,omitempty"`
	InboundRuleSecurityGroupIds    []string `json:"InboundRuleSecurityGroupIds,omitempty"`
	InboundRuleSecurityGroupNames  []string `json:"InboundRuleSecurityGroupNames,omitempty"`
	InboundRuleToPortRanges        []int64  `json:"InboundRuleToPortRanges,omitempty"`
	NetIds                         []string `json:"NetIds,omitempty"`
	OutboundRuleAccountIds         []string `json:"OutboundRuleAccountIds,omitempty"`
	OutboundRuleFromPortRanges     []int64  `json:"OutboundRuleFromPortRanges,omitempty"`
	OutboundRuleIpRanges           []string `json:"OutboundRuleIpRanges,omitempty"`
	OutboundRuleProtocols          []string `json:"OutboundRuleProtocols,omitempty"`
	OutboundRuleSecurityGroupIds   []string `json:"OutboundRuleSecurityGroupIds,omitempty"`
	OutboundRuleSecurityGroupNames []string `json:"OutboundRuleSecurityGroupNames,omitempty"`
	OutboundRuleToPortRanges       []int64  `json:"OutboundRuleToPortRanges,omitempty"`
	SecurityGroupIds               []string `json:"SecurityGroupIds,omitempty"`
	SecurityGroupNames             []string `json:"SecurityGroupNames,omitempty"`
	TagKeys                        []string `json:"TagKeys,omitempty"`
	TagValues                      []string `json:"TagValues,omitempty"`
	Tags                           []string `json:"Tags,omitempty"`
}

// implements the service definition of FiltersServices
type FiltersServices struct {
	Attributes  []Attribute `json:"Attributes,omitempty"`
	Endpoint    string      `json:"Endpoint,omitempty"`
	Schema      string      `json:"Schema,omitempty"`
	ServiceName string      `json:"ServiceName,omitempty"`
}

// implements the service definition of FiltersSnapshot
type FiltersSnapshot struct {
	AccountAliases                            []string `json:"AccountAliases,omitempty"`
	AccountIds                                []string `json:"AccountIds,omitempty"`
	Descriptions                              []string `json:"Descriptions,omitempty"`
	PermissionsToCreateVolumeAccountIds       []string `json:"PermissionsToCreateVolumeAccountIds,omitempty"`
	PermissionsToCreateVolumeGlobalPermission bool     `json:"PermissionsToCreateVolumeGlobalPermission,omitempty"`
	Progresses                                []int64  `json:"Progresses,omitempty"`
	SnapshotIds                               []string `json:"SnapshotIds,omitempty"`
	States                                    []string `json:"States,omitempty"`
	TagKeys                                   []string `json:"TagKeys,omitempty"`
	TagValues                                 []string `json:"TagValues,omitempty"`
	Tags                                      []string `json:"Tags,omitempty"`
	VolumeIds                                 []string `json:"VolumeIds,omitempty"`
	VolumeSizes                               []int64  `json:"VolumeSizes,omitempty"`
}

// implements the service definition of FiltersSubnet
type FiltersSubnet struct {
	AvailableIpsCounts []int64  `json:"AvailableIpsCounts,omitempty"`
	IpRanges           []string `json:"IpRanges,omitempty"`
	NetIds             []string `json:"NetIds,omitempty"`
	States             []string `json:"States,omitempty"`
	SubnetIds          []string `json:"SubnetIds,omitempty"`
	SubregionNames     []string `json:"SubregionNames,omitempty"`
}

// implements the service definition of FiltersTag
type FiltersTag struct {
	Keys          []string `json:"Keys,omitempty"`
	ResourceIds   []string `json:"ResourceIds,omitempty"`
	ResourceTypes []string `json:"ResourceTypes,omitempty"`
	Values        []string `json:"Values,omitempty"`
}

// implements the service definition of FiltersUserGroup
type FiltersUserGroup struct {
	Paths     []string `json:"Paths,omitempty"`
	UserNames []string `json:"UserNames,omitempty"`
}

// implements the service definition of FiltersVm
type FiltersVm struct {
	AccountIds                           []string `json:"AccountIds,omitempty"`
	ActivatedCheck                       bool     `json:"ActivatedCheck,omitempty"`
	Architectures                        []string `json:"Architectures,omitempty"`
	BlockDeviceMappingDeleteOnVmDeletion *bool    `json:"BlockDeviceMappingDeleteOnVmDeletion,omitempty"`
	BlockDeviceMappingDeviceNames        []string `json:"BlockDeviceMappingDeviceNames,omitempty"`
	BlockDeviceMappingLinkDates          []string `json:"BlockDeviceMappingLinkDates,omitempty"`
	BlockDeviceMappingStates             []string `json:"BlockDeviceMappingStates,omitempty"`
	BlockDeviceMappingVolumeIds          []string `json:"BlockDeviceMappingVolumeIds,omitempty"`
	Comments                             []string `json:"Comments,omitempty"`
	CreationDates                        []string `json:"CreationDates,omitempty"`
	DnsNames                             []string `json:"DnsNames,omitempty"`
	Hypervisors                          []string `json:"Hypervisors,omitempty"`
	ImageIds                             []string `json:"ImageIds,omitempty"`
	KernelIds                            []string `json:"KernelIds,omitempty"`
	KeypairNames                         []string `json:"KeypairNames,omitempty"`
	LaunchSortNumbers                    []int64  `json:"LaunchSortNumbers,omitempty"`
	LinkNicDeleteOnVmDeletion            bool     `json:"LinkNicDeleteOnVmDeletion,omitempty"`
	LinkNicLinkDates                     []string `json:"LinkNicLinkDates,omitempty"`
	LinkNicLinkNicIds                    []string `json:"LinkNicLinkNicIds,omitempty"`
	LinkNicLinkPublicIpIds               []string `json:"LinkNicLinkPublicIpIds,omitempty"`
	LinkNicNicIds                        []string `json:"LinkNicNicIds,omitempty"`
	LinkNicNicSortNumbers                []int64  `json:"LinkNicNicSortNumbers,omitempty"`
	LinkNicPublicIpAccountIds            []string `json:"LinkNicPublicIpAccountIds,omitempty"`
	LinkNicPublicIpIds                   []string `json:"LinkNicPublicIpIds,omitempty"`
	LinkNicPublicIps                     []string `json:"LinkNicPublicIps,omitempty"`
	LinkNicStates                        []string `json:"LinkNicStates,omitempty"`
	LinkNicVmAccountIds                  []string `json:"LinkNicVmAccountIds,omitempty"`
	LinkNicVmIds                         []string `json:"LinkNicVmIds,omitempty"`
	MonitoringStates                     []string `json:"MonitoringStates,omitempty"`
	NetIds                               []string `json:"NetIds,omitempty"`
	NicAccountIds                        []string `json:"NicAccountIds,omitempty"`
	NicActivatedCheck                    bool     `json:"NicActivatedCheck,omitempty"`
	NicDescriptions                      []string `json:"NicDescriptions,omitempty"`
	NicMacAddresses                      []string `json:"NicMacAddresses,omitempty"`
	NicNetIds                            []string `json:"NicNetIds,omitempty"`
	NicNicIds                            []string `json:"NicNicIds,omitempty"`
	NicPrivateDnsNames                   []string `json:"NicPrivateDnsNames,omitempty"`
	NicSecurityGroupIds                  []string `json:"NicSecurityGroupIds,omitempty"`
	NicSecurityGroupNames                []string `json:"NicSecurityGroupNames,omitempty"`
	NicStates                            []string `json:"NicStates,omitempty"`
	NicSubnetIds                         []string `json:"NicSubnetIds,omitempty"`
	NicSubregionNames                    []string `json:"NicSubregionNames,omitempty"`
	PlacementGroups                      []string `json:"PlacementGroups,omitempty"`
	PrivateDnsNames                      []string `json:"PrivateDnsNames,omitempty"`
	PrivateIpLinkPrivateIpAccountIds     []string `json:"PrivateIpLinkPrivateIpAccountIds,omitempty"`
	PrivateIpLinkPublicIps               []string `json:"PrivateIpLinkPublicIps,omitempty"`
	PrivateIpPrimaryIps                  []string `json:"PrivateIpPrimaryIps,omitempty"`
	PrivateIpPrivateIps                  []string `json:"PrivateIpPrivateIps,omitempty"`
	PrivateIps                           []string `json:"PrivateIps,omitempty"`
	ProductCodes                         []string `json:"ProductCodes,omitempty"`
	PublicIps                            []string `json:"PublicIps,omitempty"`
	RamDiskIds                           []string `json:"RamDiskIds,omitempty"`
	RootDeviceNames                      []string `json:"RootDeviceNames,omitempty"`
	RootDeviceTypes                      []string `json:"RootDeviceTypes,omitempty"`
	SecurityGroupIds                     []string `json:"SecurityGroupIds,omitempty"`
	SecurityGroupNames                   []string `json:"SecurityGroupNames,omitempty"`
	SpotVmRequestIds                     []string `json:"SpotVmRequestIds,omitempty"`
	SpotVms                              []string `json:"SpotVms,omitempty"`
	StateComments                        []string `json:"StateComments,omitempty"`
	SubnetIds                            []string `json:"SubnetIds,omitempty"`
	SubregionNames                       []string `json:"SubregionNames,omitempty"`
	Systems                              []string `json:"Systems,omitempty"`
	TagKeys                              []string `json:"TagKeys,omitempty"`
	TagValues                            []string `json:"TagValues,omitempty"`
	Tags                                 []string `json:"Tags,omitempty"`
	Tenancies                            []string `json:"Tenancies,omitempty"`
	Tokens                               []string `json:"Tokens,omitempty"`
	VirtualizationTypes                  []string `json:"VirtualizationTypes,omitempty"`
	VmIds                                []string `json:"VmIds,omitempty"`
	VmStates                             []string `json:"VmStates,omitempty"`
	VmTypes                              []string `json:"VmTypes,omitempty"`
	VmsSecurityGroupIds                  []string `json:"VmsSecurityGroupIds,omitempty"`
	VmsSecurityGroupNames                []string `json:"VmsSecurityGroupNames,omitempty"`
}

// implements the service definition of FiltersVmsState
type FiltersVmsState struct {
	MaintenanceEventCodes        []string `json:"MaintenanceEventCodes,omitempty"`
	MaintenanceEventDescriptions []string `json:"MaintenanceEventDescriptions,omitempty"`
	MaintenanceEventsNotAfter    []string `json:"MaintenanceEventsNotAfter,omitempty"`
	MaintenanceEventsNotBefore   []string `json:"MaintenanceEventsNotBefore,omitempty"`
	SubregionNames               []string `json:"SubregionNames,omitempty"`
	VmIds                        []string `json:"VmIds,omitempty"`
	VmStates                     []string `json:"VmStates,omitempty"`
}

// implements the service definition of FiltersVolume
type FiltersVolume struct {
	CreationDates  []string `json:"CreationDates,omitempty"`
	SnapshotIds    []string `json:"SnapshotIds,omitempty"`
	SubregionNames []string `json:"SubregionNames,omitempty"`
	TagKeys        []string `json:"TagKeys,omitempty"`
	TagValues      []string `json:"TagValues,omitempty"`
	Tags           []string `json:"Tags,omitempty"`
	VolumeIds      []string `json:"VolumeIds,omitempty"`
	VolumeSizes    []int64  `json:"VolumeSizes,omitempty"`
	VolumeTypes    []string `json:"VolumeTypes,omitempty"`
}

// implements the service definition of FiltersVpnConnection
type FiltersVpnConnection struct {
	ConnectionTypes               []string `json:"ConnectionTypes,omitempty"`
	NetToVirtualGatewayLinkNetIds []string `json:"NetToVirtualGatewayLinkNetIds,omitempty"`
	NetToVirtualGatewayLinkStates []string `json:"NetToVirtualGatewayLinkStates,omitempty"`
	States                        []string `json:"States,omitempty"`
	TagKeys                       []string `json:"TagKeys,omitempty"`
	TagValues                     []string `json:"TagValues,omitempty"`
	Tags                          []string `json:"Tags,omitempty"`
	VirtualGatewayIds             []string `json:"VirtualGatewayIds,omitempty"`
}

// implements the service definition of HealthCheck
type HealthCheck struct {
	CheckInterval      int64  `json:"CheckInterval,omitempty"`
	HealthyThreshold   int64  `json:"HealthyThreshold,omitempty"`
	Path               string `json:"Path,omitempty"`
	Port               int64  `json:"Port,omitempty"`
	Protocol           string `json:"Protocol,omitempty"`
	Timeout            int64  `json:"Timeout,omitempty"`
	UnhealthyThreshold int64  `json:"UnhealthyThreshold,omitempty"`
}

// implements the service definition of Image
type Image struct {
	AccountAlias        string                    `json:"AccountAlias,omitempty"`
	AccountId           string                    `json:"AccountId,omitempty"`
	Architecture        string                    `json:"Architecture,omitempty"`
	BlockDeviceMappings []BlockDeviceMappingImage `json:"BlockDeviceMappings,omitempty"`
	CreationDate        string                    `json:"CreationDate,omitempty"`
	Description         string                    `json:"Description,omitempty"`
	FileLocation        string                    `json:"FileLocation,omitempty"`
	ImageId             string                    `json:"ImageId,omitempty"`
	ImageName           string                    `json:"ImageName,omitempty"`
	ImageType           string                    `json:"ImageType,omitempty"`
	PermissionsToLaunch PermissionsOnResource     `json:"PermissionsToLaunch,omitempty"`
	ProductCodes        []string                  `json:"ProductCodes,omitempty"`
	RootDeviceName      string                    `json:"RootDeviceName,omitempty"`
	RootDeviceType      string                    `json:"RootDeviceType,omitempty"`
	State               string                    `json:"State,omitempty"`
	StateComment        StateComment              `json:"StateComment,omitempty"`
	Tags                []ResourceTag             `json:"Tags,omitempty"`
}

// implements the service definition of ImageExportTask
type ImageExportTask struct {
	Comment   string    `json:"Comment,omitempty"`
	ImageId   string    `json:"ImageId,omitempty"`
	OsuExport OsuExport `json:"OsuExport,omitempty"`
	Progress  int64     `json:"Progress,omitempty"`
	State     string    `json:"State,omitempty"`
	TaskId    string    `json:"TaskId,omitempty"`
}

// implements the service definition of InternetService
type InternetService struct {
	InternetServiceId string        `json:"InternetServiceId,omitempty"`
	NetId             string        `json:"NetId,omitempty"`
	State             string        `json:"State,omitempty"`
	Tags              []ResourceTag `json:"Tags,omitempty"`
}

// implements the service definition of Item
type Item struct {
	AccountId       string      `json:"AccountId,omitempty"`
	Catalog         []Catalog_0 `json:"Catalog,omitempty"`
	ComsuptionValue int         `json:"ComsuptionValue,omitempty"`
	Entry           string      `json:"Entry,omitempty"`
	FromDate        string      `json:"FromDate,omitempty"`
	PayingAccountId string      `json:"PayingAccountId,omitempty"`
	Service         string      `json:"Service,omitempty"`
	SubregionName   string      `json:"SubregionName,omitempty"`
	ToDate          string      `json:"ToDate,omitempty"`
	Type            string      `json:"Type,omitempty"`
}

// implements the service definition of Keypair
type Keypair struct {
	KeypairFingerprint string `json:"KeypairFingerprint,omitempty"`
	KeypairName        string `json:"KeypairName,omitempty"`
}

// implements the service definition of KeypairCreated
type KeypairCreated struct {
	KeypairFingerprint string `json:"KeypairFingerprint,omitempty"`
	KeypairName        string `json:"KeypairName,omitempty"`
	PrivateKey         string `json:"PrivateKey,omitempty"`
}

// implements the service definition of LinkInternetServiceRequest
type LinkInternetServiceRequest struct {
	DryRun            bool   `json:"DryRun,omitempty"`
	InternetServiceId string `json:"InternetServiceId,omitempty"`
	NetId             string `json:"NetId,omitempty"`
}

// implements the service definition of LinkInternetServiceResponse
type LinkInternetServiceResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of LinkNic
type LinkNic struct {
	DeleteOnVmDeletion *bool  `json:"DeleteOnVmDeletion,omitempty"`
	DeviceNumber       int64  `json:"DeviceNumber,omitempty"`
	LinkNicId          string `json:"LinkNicId,omitempty"`
	State              string `json:"State,omitempty"`
	VmAccountId        string `json:"VmAccountId,omitempty"`
	VmId               string `json:"VmId,omitempty"`
}

// implements the service definition of LinkNicLight
type LinkNicLight struct {
	DeleteOnVmDeletion *bool  `json:"DeleteOnVmDeletion,omitempty"`
	DeviceNumber       int64  `json:"DeviceNumber,omitempty"`
	LinkNicId          string `json:"LinkNicId,omitempty"`
	State              string `json:"State,omitempty"`
}

// implements the service definition of LinkNicRequest
type LinkNicRequest struct {
	DeviceNumber int64  `json:"DeviceNumber,omitempty"`
	DryRun       bool   `json:"DryRun,omitempty"`
	NicId        string `json:"NicId,omitempty"`
	VmId         string `json:"VmId,omitempty"`
}

// implements the service definition of LinkNicResponse
type LinkNicResponse struct {
	LinkNicId       string          `json:"LinkNicId,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of LinkNicToUpdate
type LinkNicToUpdate struct {
	DeleteOnVmDeletion *bool  `json:"DeleteOnVmDeletion,omitempty"`
	LinkNicId          string `json:"LinkNicId,omitempty"`
}

// implements the service definition of LinkPolicyRequest
type LinkPolicyRequest struct {
	DryRun        bool   `json:"DryRun,omitempty"`
	PolicyId      string `json:"PolicyId,omitempty"`
	UserGroupName string `json:"UserGroupName,omitempty"`
	UserName      string `json:"UserName,omitempty"`
}

// implements the service definition of LinkPolicyResponse
type LinkPolicyResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of LinkPrivateIpsRequest
type LinkPrivateIpsRequest struct {
	AllowRelink             bool     `json:"AllowRelink,omitempty"`
	DryRun                  bool     `json:"DryRun,omitempty"`
	NicId                   string   `json:"NicId,omitempty"`
	PrivateIps              []string `json:"PrivateIps,omitempty"`
	SecondaryPrivateIpCount int64    `json:"SecondaryPrivateIpCount,omitempty"`
}

// implements the service definition of LinkPrivateIpsResponse
type LinkPrivateIpsResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of LinkPublicIp
type LinkPublicIp struct {
	LinkPublicIpId    string `json:"LinkPublicIpId,omitempty"`
	PublicDnsName     string `json:"PublicDnsName,omitempty"`
	PublicIp          string `json:"PublicIp,omitempty"`
	PublicIpAccountId string `json:"PublicIpAccountId,omitempty"`
	PublicIpId        string `json:"PublicIpId,omitempty"`
}

// implements the service definition of LinkPublicIpLightForVm
type LinkPublicIpLightForVm struct {
	PublicDnsName     string `json:"PublicDnsName,omitempty"`
	PublicIp          string `json:"PublicIp,omitempty"`
	PublicIpAccountId string `json:"PublicIpAccountId,omitempty"`
}

// implements the service definition of LinkPublicIpRequest
type LinkPublicIpRequest struct {
	AllowRelink bool   `json:"AllowRelink,omitempty"`
	DryRun      bool   `json:"DryRun,omitempty"`
	NicId       string `json:"NicId,omitempty"`
	PrivateIp   string `json:"PrivateIp,omitempty"`
	PublicIp    string `json:"PublicIp,omitempty"`
	PublicIpId  string `json:"PublicIpId,omitempty"`
	VmId        string `json:"VmId,omitempty"`
}

// implements the service definition of LinkPublicIpResponse
type LinkPublicIpResponse struct {
	LinkPublicIpId  string          `json:"LinkPublicIpId,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of LinkRouteTable
type LinkRouteTable struct {
	LinkRouteTableId string `json:"LinkRouteTableId,omitempty"`
	Main             bool   `json:"Main,omitempty"`
	RouteTableId     string `json:"RouteTableId,omitempty"`
	SubnetId         string `json:"SubnetId,omitempty"`
}

// implements the service definition of LinkRouteTableRequest
type LinkRouteTableRequest struct {
	DryRun       bool   `json:"DryRun,omitempty"`
	RouteTableId string `json:"RouteTableId,omitempty"`
	SubnetId     string `json:"SubnetId,omitempty"`
}

// implements the service definition of LinkRouteTableResponse
type LinkRouteTableResponse struct {
	LinkRouteTableId string          `json:"LinkRouteTableId,omitempty"`
	ResponseContext  ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of LinkVirtualGatewayRequest
type LinkVirtualGatewayRequest struct {
	DryRun           bool   `json:"DryRun,omitempty"`
	NetId            string `json:"NetId,omitempty"`
	VirtualGatewayId string `json:"VirtualGatewayId,omitempty"`
}

// implements the service definition of LinkVirtualGatewayResponse
type LinkVirtualGatewayResponse struct {
	NetToVirtualGatewayLink NetToVirtualGatewayLink `json:"NetToVirtualGatewayLink,omitempty"`
	ResponseContext         ResponseContext         `json:"ResponseContext,omitempty"`
}

// implements the service definition of LinkVolumeRequest
type LinkVolumeRequest struct {
	DeviceName string `json:"DeviceName,omitempty"`
	DryRun     bool   `json:"DryRun,omitempty"`
	VmId       string `json:"VmId,omitempty"`
	VolumeId   string `json:"VolumeId,omitempty"`
}

// implements the service definition of LinkVolumeResponse
type LinkVolumeResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of LinkedVolume
type LinkedVolume struct {
	DeleteOnVmDeletion *bool  `json:"DeleteOnVmDeletion,omitempty"`
	DeviceName         string `json:"DeviceName,omitempty"`
	State              string `json:"State,omitempty"`
	VmId               string `json:"VmId,omitempty"`
	VolumeId           string `json:"VolumeId,omitempty"`
}

// implements the service definition of Listener
type Listener struct {
	BackendPort          int64    `json:"BackendPort,omitempty"`
	BackendProtocol      string   `json:"BackendProtocol,omitempty"`
	LoadBalancerPort     int64    `json:"LoadBalancerPort,omitempty"`
	LoadBalancerProtocol string   `json:"LoadBalancerProtocol,omitempty"`
	PolicyNames          []string `json:"PolicyNames,omitempty"`
	ServerCertificateId  string   `json:"ServerCertificateId,omitempty"`
}

// implements the service definition of ListenerForCreation
type ListenerForCreation struct {
	BackendPort          int64  `json:"BackendPort,omitempty"`
	BackendProtocol      string `json:"BackendProtocol,omitempty"`
	LoadBalancerPort     int64  `json:"LoadBalancerPort,omitempty"`
	LoadBalancerProtocol string `json:"LoadBalancerProtocol,omitempty"`
	ServerCertificateId  string `json:"ServerCertificateId,omitempty"`
}

// implements the service definition of ListenerRule
type ListenerRule struct {
	Action           string `json:"Action,omitempty"`
	HostNamePattern  string `json:"HostNamePattern,omitempty"`
	ListenerRuleId   string `json:"ListenerRuleId,omitempty"`
	ListenerRuleName string `json:"ListenerRuleName,omitempty"`
	PathPattern      string `json:"PathPattern,omitempty"`
	Priority         int64  `json:"Priority,omitempty"`
}

// implements the service definition of ListenerRules
type ListenerRules struct {
	ListenerId   string       `json:"ListenerId,omitempty"`
	ListenerRule ListenerRule `json:"ListenerRule,omitempty"`
	VmIds        []string     `json:"VmIds,omitempty"`
}

// implements the service definition of LoadBalancer
type LoadBalancer struct {
	AccessLog                        AccessLog                        `json:"AccessLog,omitempty"`
	ApplicationStickyCookiePolicies  []ApplicationStickyCookiePolicy  `json:"ApplicationStickyCookiePolicies,omitempty"`
	BackendVmsIds                    []string                         `json:"BackendVmsIds,omitempty"`
	DnsName                          string                           `json:"DnsName,omitempty"`
	HealthCheck                      HealthCheck                      `json:"HealthCheck,omitempty"`
	Listeners                        []Listener                       `json:"Listeners,omitempty"`
	LoadBalancerName                 string                           `json:"LoadBalancerName,omitempty"`
	LoadBalancerStickyCookiePolicies []LoadBalancerStickyCookiePolicy `json:"LoadBalancerStickyCookiePolicies,omitempty"`
	LoadBalancerType                 string                           `json:"LoadBalancerType,omitempty"`
	NetId                            string                           `json:"NetId,omitempty"`
	SecurityGroups                   []string                         `json:"SecurityGroups,omitempty"`
	SourceSecurityGroup              SourceSecurityGroup              `json:"SourceSecurityGroup,omitempty"`
	Subnets                          []string                         `json:"Subnets,omitempty"`
	SubregionNames                   []string                         `json:"SubregionNames,omitempty"`
	Tags                             []ResourceTag                    `json:"Tags,omitempty"`
}

// implements the service definition of LoadBalancerLight
type LoadBalancerLight struct {
	LoadBalancerName string `json:"LoadBalancerName,omitempty"`
	LoadBalancerPort int64  `json:"LoadBalancerPort,omitempty"`
}

// implements the service definition of LoadBalancerStickyCookiePolicy
type LoadBalancerStickyCookiePolicy struct {
	PolicyName string `json:"PolicyName,omitempty"`
}

// implements the service definition of Location
type Location struct {
	Code string `json:"Code,omitempty"`
	Name string `json:"Name,omitempty"`
}

// implements the service definition of Log
type Log struct {
	CallDuration       int64  `json:"CallDuration,omitempty"`
	QueryAccessKey     string `json:"QueryAccessKey,omitempty"`
	QueryApiName       string `json:"QueryApiName,omitempty"`
	QueryApiVersion    string `json:"QueryApiVersion,omitempty"`
	QueryCallName      string `json:"QueryCallName,omitempty"`
	QueryDate          string `json:"QueryDate,omitempty"`
	QueryIpAddress     string `json:"QueryIpAddress,omitempty"`
	QueryRaw           string `json:"QueryRaw,omitempty"`
	QuerySize          int64  `json:"QuerySize,omitempty"`
	QueryUserAgent     string `json:"QueryUserAgent,omitempty"`
	ResponseId         string `json:"ResponseId,omitempty"`
	ResponseSize       int64  `json:"ResponseSize,omitempty"`
	ResponseStatusCode int64  `json:"ResponseStatusCode,omitempty"`
}

// implements the service definition of MaintenanceEvent
type MaintenanceEvent struct {
	Code        string `json:"Code,omitempty"`
	Description string `json:"Description,omitempty"`
	NotAfter    string `json:"NotAfter,omitempty"`
	NotBefore   string `json:"NotBefore,omitempty"`
}

// implements the service definition of NatService
type NatService struct {
	NatServiceId string          `json:"NatServiceId,omitempty"`
	NetId        string          `json:"NetId,omitempty"`
	PublicIps    []PublicIpLight `json:"PublicIps,omitempty"`
	State        string          `json:"State,omitempty"`
	SubnetId     string          `json:"SubnetId,omitempty"`
}

// implements the service definition of Net
type Net struct {
	DhcpOptionsSetId string        `json:"DhcpOptionsSetId,omitempty"`
	IpRange          string        `json:"IpRange,omitempty"`
	NetId            string        `json:"NetId,omitempty"`
	State            string        `json:"State,omitempty"`
	Tags             []ResourceTag `json:"Tags,omitempty"`
	Tenancy          string        `json:"Tenancy,omitempty"`
}

// implements the service definition of NetAccessPoint
type NetAccessPoint struct {
	NetAccessPointId string   `json:"NetAccessPointId,omitempty"`
	NetId            string   `json:"NetId,omitempty"`
	PrefixListName   string   `json:"PrefixListName,omitempty"`
	RouteTableIds    []string `json:"RouteTableIds,omitempty"`
	State            string   `json:"State,omitempty"`
}

// implements the service definition of NetPeering
type NetPeering struct {
	AccepterNet  AccepterNet     `json:"AccepterNet,omitempty"`
	NetPeeringId string          `json:"NetPeeringId,omitempty"`
	SourceNet    SourceNet       `json:"SourceNet,omitempty"`
	State        NetPeeringState `json:"State,omitempty"`
	Tags         []ResourceTag   `json:"Tags,omitempty"`
}

// implements the service definition of NetPeeringState
type NetPeeringState struct {
	Message string `json:"Message,omitempty"`
	Name    string `json:"Name,omitempty"`
}

// implements the service definition of NetToVirtualGatewayLink
type NetToVirtualGatewayLink struct {
	NetId string `json:"NetId,omitempty"`
	State string `json:"State,omitempty"`
}

// implements the service definition of Nic
type Nic struct {
	AccountId           string               `json:"AccountId,omitempty"`
	Description         string               `json:"Description,omitempty"`
	IsSourceDestChecked bool                 `json:"IsSourceDestChecked,omitempty"`
	LinkNic             LinkNic              `json:"LinkNic,omitempty"`
	LinkPublicIp        LinkPublicIp         `json:"LinkPublicIp,omitempty"`
	MacAddress          string               `json:"MacAddress,omitempty"`
	NetId               string               `json:"NetId,omitempty"`
	NicId               string               `json:"NicId,omitempty"`
	PrivateDnsName      string               `json:"PrivateDnsName,omitempty"`
	PrivateIps          []PrivateIp          `json:"PrivateIps,omitempty"`
	SecurityGroups      []SecurityGroupLight `json:"SecurityGroups,omitempty"`
	State               string               `json:"State,omitempty"`
	SubnetId            string               `json:"SubnetId,omitempty"`
	SubregionName       string               `json:"SubregionName,omitempty"`
	Tags                []ResourceTag        `json:"Tags,omitempty"`
}

// implements the service definition of NicForVmCreation
type NicForVmCreation struct {
	DeleteOnVmDeletion      bool             `json:"DeleteOnVmDeletion,omitempty"`
	Description             string           `json:"Description,omitempty"`
	DeviceNumber            int64            `json:"DeviceNumber,omitempty"`
	NicId                   string           `json:"NicId,omitempty"`
	PrivateIps              []PrivateIpLight `json:"PrivateIps,omitempty"`
	SecondaryPrivateIpCount int64            `json:"SecondaryPrivateIpCount,omitempty"`
	SecurityGroupIds        []string         `json:"SecurityGroupIds,omitempty"`
	SubnetId                string           `json:"SubnetId,omitempty"`
}

// implements the service definition of NicLight
type NicLight struct {
	AccountId           string                 `json:"AccountId,omitempty"`
	Description         string                 `json:"Description,omitempty"`
	IsSourceDestChecked bool                   `json:"IsSourceDestChecked,omitempty"`
	LinkNic             LinkNicLight           `json:"LinkNic,omitempty"`
	LinkPublicIp        LinkPublicIpLightForVm `json:"LinkPublicIp,omitempty"`
	MacAddress          string                 `json:"MacAddress,omitempty"`
	NetId               string                 `json:"NetId,omitempty"`
	NicId               string                 `json:"NicId,omitempty"`
	PrivateDnsName      string                 `json:"PrivateDnsName,omitempty"`
	PrivateIps          []PrivateIpLightForVm  `json:"PrivateIps,omitempty"`
	SecurityGroups      []SecurityGroupLight   `json:"SecurityGroups,omitempty"`
	State               string                 `json:"State,omitempty"`
	SubnetId            string                 `json:"SubnetId,omitempty"`
}

// implements the service definition of OsuApiKey
type OsuApiKey struct {
	ApiKeyId  string `json:"ApiKeyId,omitempty"`
	SecretKey string `json:"SecretKey,omitempty"`
}

// implements the service definition of OsuExport
type OsuExport struct {
	DiskImageFormat string    `json:"DiskImageFormat,omitempty"`
	OsuApiKey       OsuApiKey `json:"OsuApiKey,omitempty"`
	OsuBucket       string    `json:"OsuBucket,omitempty"`
	OsuManifestUrl  string    `json:"OsuManifestUrl,omitempty"`
	OsuPrefix       string    `json:"OsuPrefix,omitempty"`
}

// implements the service definition of PermissionsOnResource
type PermissionsOnResource struct {
	AccountIds       []string `json:"AccountIds,omitempty"`
	GlobalPermission bool     `json:"GlobalPermission,omitempty"`
}

// implements the service definition of PermissionsOnResourceCreation
type PermissionsOnResourceCreation struct {
	Additions PermissionsOnResource `json:"Additions,omitempty"`
	Removals  PermissionsOnResource `json:"Removals,omitempty"`
}

// implements the service definition of Placement
type Placement struct {
	SubregionName string `json:"SubregionName,omitempty"`
	Tenancy       string `json:"Tenancy,omitempty"`
}

// implements the service definition of Policy
type Policy struct {
	Description            string `json:"Description,omitempty"`
	IsLinkable             bool   `json:"IsLinkable,omitempty"`
	Path                   string `json:"Path,omitempty"`
	PolicyDefaultVersionId string `json:"PolicyDefaultVersionId,omitempty"`
	PolicyId               string `json:"PolicyId,omitempty"`
	PolicyName             string `json:"PolicyName,omitempty"`
	ResourcesCount         int64  `json:"ResourcesCount,omitempty"`
}

// implements the service definition of PrefixLists
type PrefixLists struct {
	IpRanges       []string `json:"IpRanges,omitempty"`
	PrefixListId   string   `json:"PrefixListId,omitempty"`
	PrefixListName string   `json:"PrefixListName,omitempty"`
}

// implements the service definition of PricingDetail
type PricingDetail struct {
	Count int64 `json:"Count,omitempty"`
}

// implements the service definition of PrivateIp
type PrivateIp struct {
	IsPrimary      bool         `json:"IsPrimary,omitempty"`
	LinkPublicIp   LinkPublicIp `json:"LinkPublicIp,omitempty"`
	PrivateDnsName string       `json:"PrivateDnsName,omitempty"`
	PrivateIp      string       `json:"PrivateIp,omitempty"`
}

// implements the service definition of PrivateIpLight
type PrivateIpLight struct {
	IsPrimary bool   `json:"IsPrimary,omitempty"`
	PrivateIp string `json:"PrivateIp,omitempty"`
}

// implements the service definition of PrivateIpLightForVm
type PrivateIpLightForVm struct {
	IsPrimary      bool                   `json:"IsPrimary,omitempty"`
	LinkPublicIp   LinkPublicIpLightForVm `json:"LinkPublicIp,omitempty"`
	PrivateDnsName string                 `json:"PrivateDnsName,omitempty"`
	PrivateIp      string                 `json:"PrivateIp,omitempty"`
}

// implements the service definition of ProductType
type ProductType struct {
	Description   string `json:"Description,omitempty"`
	ProductTypeId string `json:"ProductTypeId,omitempty"`
	Vendor        string `json:"Vendor,omitempty"`
}

// implements the service definition of PublicIp
type PublicIp struct {
	LinkPublicIpId string `json:"LinkPublicIpId,omitempty"`
	NicAccountId   string `json:"NicAccountId,omitempty"`
	NicId          string `json:"NicId,omitempty"`
	PrivateIp      string `json:"PrivateIp,omitempty"`
	PublicIp       string `json:"PublicIp,omitempty"`
	PublicIpId     string `json:"PublicIpId,omitempty"`
	VmId           string `json:"VmId,omitempty"`
}

// implements the service definition of PublicIpLight
type PublicIpLight struct {
	PublicIp   string `json:"PublicIp,omitempty"`
	PublicIpId string `json:"PublicIpId,omitempty"`
}

// implements the service definition of PurchaseReservedVmsOfferRequest
type PurchaseReservedVmsOfferRequest struct {
	DryRun             bool   `json:"DryRun,omitempty"`
	ReservedVmsOfferId string `json:"ReservedVmsOfferId,omitempty"`
	VmCount            int64  `json:"VmCount,omitempty"`
}

// implements the service definition of PurchaseReservedVmsOfferResponse
type PurchaseReservedVmsOfferResponse struct {
	ReservedVmsId   string          `json:"ReservedVmsId,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of Quota
type Quota struct {
	AccountId        string `json:"AccountId,omitempty"`
	Description      string `json:"Description,omitempty"`
	MaxValue         int64  `json:"MaxValue,omitempty"`
	Name             string `json:"Name,omitempty"`
	QuotaCollection  string `json:"QuotaCollection,omitempty"`
	ShortDescription string `json:"ShortDescription,omitempty"`
	UsedValue        int64  `json:"UsedValue,omitempty"`
}

// implements the service definition of QuotaTypes
type QuotaTypes struct {
	QuotaType string  `json:"QuotaType,omitempty"`
	Quotas    []Quota `json:"Quotas,omitempty"`
}

// implements the service definition of ReadAccountConsumptionRequest
type ReadAccountConsumptionRequest struct {
	DryRun   bool   `json:"DryRun,omitempty"`
	FromDate string `json:"FromDate,omitempty"`
	ToDate   string `json:"ToDate,omitempty"`
}

// implements the service definition of ReadAccountConsumptionResponse
type ReadAccountConsumptionResponse struct {
	ConsumptionEntries ConsumptionEntries `json:"ConsumptionEntries,omitempty"`
	ResponseContext    ResponseContext    `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadAccountRequest
type ReadAccountRequest struct {
	DryRun bool `json:"DryRun,omitempty"`
}

// implements the service definition of ReadAccountResponse
type ReadAccountResponse struct {
	Account         Account         `json:"Account,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadAdminPasswordRequest
type ReadAdminPasswordRequest struct {
	DryRun bool   `json:"DryRun,omitempty"`
	VmId   string `json:"VmId,omitempty"`
}

// implements the service definition of ReadAdminPasswordResponse
type ReadAdminPasswordResponse struct {
	AdminPassword   string          `json:"AdminPassword,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	VmId            string          `json:"VmId,omitempty"`
}

// implements the service definition of ReadApiKeysRequest
type ReadApiKeysRequest struct {
	DryRun   bool          `json:"DryRun,omitempty"`
	Tags     []ResourceTag `json:"Tags,omitempty"`
	UserName string        `json:"UserName,omitempty"`
}

// implements the service definition of ReadApiKeysResponse
type ReadApiKeysResponse struct {
	ApiKeys         []ApiKey        `json:"ApiKeys,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadApiLogsRequest
type ReadApiLogsRequest struct {
	DryRun  bool          `json:"DryRun,omitempty"`
	Filters FiltersApiLog `json:"Filters,omitempty"`
	With    With          `json:"With,omitempty"`
}

// implements the service definition of ReadApiLogsResponse
type ReadApiLogsResponse struct {
	Logs            []Log           `json:"Logs,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadBillableDigestRequest
type ReadBillableDigestRequest struct {
	AccountId      string `json:"AccountId,omitempty"`
	DryRun         bool   `json:"DryRun,omitempty"`
	FromDate       string `json:"FromDate,omitempty"`
	InvoiceState   string `json:"InvoiceState,omitempty"`
	IsConsolidated bool   `json:"IsConsolidated,omitempty"`
	ToDate         string `json:"ToDate,omitempty"`
}

// implements the service definition of ReadBillableDigestResponse
type ReadBillableDigestResponse struct {
	Items           []Item          `json:"Items,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadCatalogRequest
type ReadCatalogRequest struct {
	DryRun bool `json:"DryRun,omitempty"`
}

// implements the service definition of ReadCatalogResponse
type ReadCatalogResponse struct {
	Catalog         Catalog_1       `json:"Catalog,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadClientGatewaysRequest
type ReadClientGatewaysRequest struct {
	ClientGatewayIds []string           `json:"ClientGatewayIds,omitempty"`
	DryRun           bool               `json:"DryRun,omitempty"`
	Filters          []FiltersOldFormat `json:"Filters,omitempty"`
}

// implements the service definition of ReadClientGatewaysResponse
type ReadClientGatewaysResponse struct {
	ClientGateways  []ClientGateway `json:"ClientGateways,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadConsoleOutputRequest
type ReadConsoleOutputRequest struct {
	DryRun bool   `json:"DryRun,omitempty"`
	VmId   string `json:"VmId,omitempty"`
}

// implements the service definition of ReadConsoleOutputResponse
type ReadConsoleOutputResponse struct {
	ConsoleOutput   string          `json:"ConsoleOutput,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	VmId            string          `json:"VmId,omitempty"`
}

// implements the service definition of ReadDhcpOptionsRequest
type ReadDhcpOptionsRequest struct {
	DryRun  bool               `json:"DryRun,omitempty"`
	Filters FiltersDhcpOptions `json:"Filters,omitempty"`
}

// implements the service definition of ReadDhcpOptionsResponse
type ReadDhcpOptionsResponse struct {
	DhcpOptionsSets []DhcpOptionsSet `json:"DhcpOptionsSets,omitempty"`
	ResponseContext ResponseContext  `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadDirectLinkInterfacesRequest
type ReadDirectLinkInterfacesRequest struct {
	DirectLinkId          string `json:"DirectLinkId,omitempty"`
	DirectLinkInterfaceId string `json:"DirectLinkInterfaceId,omitempty"`
	DryRun                bool   `json:"DryRun,omitempty"`
}

// implements the service definition of ReadDirectLinkInterfacesResponse
type ReadDirectLinkInterfacesResponse struct {
	DirectLinkInterfaces []DirectLinkInterfaces `json:"DirectLinkInterfaces,omitempty"`
	ResponseContext      ResponseContext        `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadDirectLinksRequest
type ReadDirectLinksRequest struct {
	DirectLinkId string `json:"DirectLinkId,omitempty"`
	DryRun       bool   `json:"DryRun,omitempty"`
}

// implements the service definition of ReadDirectLinksResponse
type ReadDirectLinksResponse struct {
	DirectLinks     []DirectLink    `json:"DirectLinks,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadImageExportTasksRequest
type ReadImageExportTasksRequest struct {
	DryRun  bool              `json:"DryRun,omitempty"`
	Filters FiltersExportTask `json:"Filters,omitempty"`
}

// implements the service definition of ReadImageExportTasksResponse
type ReadImageExportTasksResponse struct {
	ImageExportTasks []ImageExportTask `json:"ImageExportTasks,omitempty"`
	ResponseContext  ResponseContext   `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadImagesRequest
type ReadImagesRequest struct {
	DryRun  bool         `json:"DryRun,omitempty"`
	Filters FiltersImage `json:"Filters,omitempty"`
}

// implements the service definition of ReadImagesResponse
type ReadImagesResponse struct {
	Images          []Image         `json:"Images,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadInternetServicesRequest
type ReadInternetServicesRequest struct {
	DryRun  bool                   `json:"DryRun,omitempty"`
	Filters FiltersInternetService `json:"Filters,omitempty"`
}

// implements the service definition of ReadInternetServicesResponse
type ReadInternetServicesResponse struct {
	InternetServices []InternetService `json:"InternetServices,omitempty"`
	ResponseContext  ResponseContext   `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadKeypairsRequest
type ReadKeypairsRequest struct {
	DryRun  bool           `json:"DryRun,omitempty"`
	Filters FiltersKeypair `json:"Filters,omitempty"`
}

// implements the service definition of ReadKeypairsResponse
type ReadKeypairsResponse struct {
	Keypairs        []Keypair       `json:"Keypairs,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadListenerRulesRequest
type ReadListenerRulesRequest struct {
	DryRun            bool     `json:"DryRun,omitempty"`
	ListenerRuleNames []string `json:"ListenerRuleNames,omitempty"`
}

// implements the service definition of ReadListenerRulesResponse
type ReadListenerRulesResponse struct {
	ListenerRules   []ListenerRules `json:"ListenerRules,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadLoadBalancersRequest
type ReadLoadBalancersRequest struct {
	DryRun  bool                `json:"DryRun,omitempty"`
	Filters FiltersLoadBalancer `json:"Filters,omitempty"`
}

// implements the service definition of ReadLoadBalancersResponse
type ReadLoadBalancersResponse struct {
	LoadBalancers   []LoadBalancer  `json:"LoadBalancers,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadLocationsRequest
type ReadLocationsRequest struct {
	DryRun bool `json:"DryRun,omitempty"`
}

// implements the service definition of ReadLocationsResponse
type ReadLocationsResponse struct {
	Locations       []Location      `json:"Locations,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadNatServicesRequest
type ReadNatServicesRequest struct {
	DryRun  bool              `json:"DryRun,omitempty"`
	Filters FiltersNatService `json:"Filters,omitempty"`
}

// implements the service definition of ReadNatServicesResponse
type ReadNatServicesResponse struct {
	NatServices     []NatService    `json:"NatServices,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadNetAccessPointServicesRequest
type ReadNetAccessPointServicesRequest struct {
	DryRun bool `json:"DryRun,omitempty"`
}

// implements the service definition of ReadNetAccessPointServicesResponse
type ReadNetAccessPointServicesResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	ServiceNames    []string        `json:"ServiceNames,omitempty"`
}

// implements the service definition of ReadNetAccessPointsRequest
type ReadNetAccessPointsRequest struct {
	DryRun            bool               `json:"DryRun,omitempty"`
	Filters           []FiltersOldFormat `json:"Filters,omitempty"`
	NetAccessPointIds []string           `json:"NetAccessPointIds,omitempty"`
}

// implements the service definition of ReadNetAccessPointsResponse
type ReadNetAccessPointsResponse struct {
	NetAccessPoints []NetAccessPoint `json:"NetAccessPoints,omitempty"`
	ResponseContext ResponseContext  `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadNetPeeringsRequest
type ReadNetPeeringsRequest struct {
	DryRun  bool              `json:"DryRun,omitempty"`
	Filters FiltersNetPeering `json:"Filters,omitempty"`
}

// implements the service definition of ReadNetPeeringsResponse
type ReadNetPeeringsResponse struct {
	NetPeerings     []NetPeering    `json:"NetPeerings,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadNetsRequest
type ReadNetsRequest struct {
	DryRun  bool       `json:"DryRun,omitempty"`
	Filters FiltersNet `json:"Filters,omitempty"`
}

// implements the service definition of ReadNetsResponse
type ReadNetsResponse struct {
	Nets            []Net           `json:"Nets,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadNicsRequest
type ReadNicsRequest struct {
	DryRun  bool       `json:"DryRun,omitempty"`
	Filters FiltersNic `json:"Filters,omitempty"`
}

// implements the service definition of ReadNicsResponse
type ReadNicsResponse struct {
	Nics            []Nic           `json:"Nics,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadPoliciesRequest
type ReadPoliciesRequest struct {
	DryRun        bool   `json:"DryRun,omitempty"`
	IsLinked      bool   `json:"IsLinked,omitempty"`
	Path          string `json:"Path,omitempty"`
	UserGroupName string `json:"UserGroupName,omitempty"`
	UserName      string `json:"UserName,omitempty"`
}

// implements the service definition of ReadPoliciesResponse
type ReadPoliciesResponse struct {
	Policies        []Policy        `json:"Policies,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadPrefixListsRequest
type ReadPrefixListsRequest struct {
	DryRun        bool               `json:"DryRun,omitempty"`
	Filters       []FiltersOldFormat `json:"Filters,omitempty"`
	PrefixListIds []string           `json:"PrefixListIds,omitempty"`
}

// implements the service definition of ReadPrefixListsResponse
type ReadPrefixListsResponse struct {
	PrefixLists     []PrefixLists   `json:"PrefixLists,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadProductTypesRequest
type ReadProductTypesRequest struct {
	DryRun  bool               `json:"DryRun,omitempty"`
	Filters []FiltersOldFormat `json:"Filters,omitempty"`
}

// implements the service definition of ReadProductTypesResponse
type ReadProductTypesResponse struct {
	ProductTypes    []ProductType   `json:"ProductTypes,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadPublicCatalogRequest
type ReadPublicCatalogRequest struct {
	DryRun bool `json:"DryRun,omitempty"`
}

// implements the service definition of ReadPublicCatalogResponse
type ReadPublicCatalogResponse struct {
	Catalog         Catalog_1       `json:"Catalog,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadPublicIpRangesRequest
type ReadPublicIpRangesRequest struct {
	DryRun bool `json:"DryRun,omitempty"`
}

// implements the service definition of ReadPublicIpRangesResponse
type ReadPublicIpRangesResponse struct {
	PublicIps       []string        `json:"PublicIps,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadPublicIpsRequest
type ReadPublicIpsRequest struct {
	DryRun  bool            `json:"DryRun,omitempty"`
	Filters FiltersPublicIp `json:"Filters,omitempty"`
}

// implements the service definition of ReadPublicIpsResponse
type ReadPublicIpsResponse struct {
	PublicIps       []PublicIp      `json:"PublicIps,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadQuotasRequest
type ReadQuotasRequest struct {
	DryRun     bool               `json:"DryRun,omitempty"`
	Filters    []FiltersOldFormat `json:"Filters,omitempty"`
	QuotaNames []string           `json:"QuotaNames,omitempty"`
}

// implements the service definition of ReadQuotasResponse
type ReadQuotasResponse struct {
	QuotaTypes      []QuotaTypes    `json:"QuotaTypes,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadRegionConfigRequest
type ReadRegionConfigRequest struct {
	DryRun   bool   `json:"DryRun,omitempty"`
	FromDate string `json:"FromDate,omitempty"`
}

// implements the service definition of ReadRegionConfigResponse
type ReadRegionConfigResponse struct {
	RegionConfig    RegionConfig    `json:"RegionConfig,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadRegionsRequest
type ReadRegionsRequest struct {
	DryRun      bool               `json:"DryRun,omitempty"`
	Filters     []FiltersOldFormat `json:"Filters,omitempty"`
	RegionNames []string           `json:"RegionNames,omitempty"`
}

// implements the service definition of ReadRegionsResponse
type ReadRegionsResponse struct {
	Regions         []Region        `json:"Regions,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadReservedVmOffersRequest
type ReadReservedVmOffersRequest struct {
	DryRun              bool               `json:"DryRun,omitempty"`
	Filters             []FiltersOldFormat `json:"Filters,omitempty"`
	OfferingType        string             `json:"OfferingType,omitempty"`
	ProductType         string             `json:"ProductType,omitempty"`
	ReservedVmsOfferIds []string           `json:"ReservedVmsOfferIds,omitempty"`
	SubregionName       string             `json:"SubregionName,omitempty"`
	Tenancy             string             `json:"Tenancy,omitempty"`
	VmType              string             `json:"VmType,omitempty"`
}

// implements the service definition of ReadReservedVmOffersResponse
type ReadReservedVmOffersResponse struct {
	ReservedVmsOffers []ReservedVmsOffer `json:"ReservedVmsOffers,omitempty"`
	ResponseContext   ResponseContext    `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadReservedVmsRequest
type ReadReservedVmsRequest struct {
	DryRun         bool               `json:"DryRun,omitempty"`
	Filters        []FiltersOldFormat `json:"Filters,omitempty"`
	OfferingType   string             `json:"OfferingType,omitempty"`
	ReservedVmsIds []string           `json:"ReservedVmsIds,omitempty"`
	SubregionName  string             `json:"SubregionName,omitempty"`
}

// implements the service definition of ReadReservedVmsResponse
type ReadReservedVmsResponse struct {
	ReservedVms     []ReservedVm    `json:"ReservedVms,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadRouteTablesRequest
type ReadRouteTablesRequest struct {
	DryRun  bool              `json:"DryRun,omitempty"`
	Filters FiltersRouteTable `json:"Filters,omitempty"`
}

// implements the service definition of ReadRouteTablesResponse
type ReadRouteTablesResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	RouteTables     []RouteTable    `json:"RouteTables,omitempty"`
}

// implements the service definition of ReadSecurityGroupsRequest
type ReadSecurityGroupsRequest struct {
	DryRun  bool                 `json:"DryRun,omitempty"`
	Filters FiltersSecurityGroup `json:"Filters,omitempty"`
}

// implements the service definition of ReadSecurityGroupsResponse
type ReadSecurityGroupsResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	SecurityGroups  []SecurityGroup `json:"SecurityGroups,omitempty"`
}

// implements the service definition of ReadServerCertificatesRequest
type ReadServerCertificatesRequest struct {
	DryRun bool   `json:"DryRun,omitempty"`
	Path   string `json:"Path,omitempty"`
}

// implements the service definition of ReadServerCertificatesResponse
type ReadServerCertificatesResponse struct {
	ResponseContext    ResponseContext     `json:"ResponseContext,omitempty"`
	ServerCertificates []ServerCertificate `json:"ServerCertificates,omitempty"`
}

// implements the service definition of ReadSnapshotExportTasksRequest
type ReadSnapshotExportTasksRequest struct {
	DryRun  bool     `json:"DryRun,omitempty"`
	TaskIds []string `json:"TaskIds,omitempty"`
}

// implements the service definition of ReadSnapshotExportTasksResponse
type ReadSnapshotExportTasksResponse struct {
	ResponseContext     ResponseContext      `json:"ResponseContext,omitempty"`
	SnapshotExportTasks []SnapshotExportTask `json:"SnapshotExportTasks,omitempty"`
}

// implements the service definition of ReadSnapshotsRequest
type ReadSnapshotsRequest struct {
	DryRun  bool            `json:"DryRun,omitempty"`
	Filters FiltersSnapshot `json:"Filters,omitempty"`
}

// implements the service definition of ReadSnapshotsResponse
type ReadSnapshotsResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	Snapshots       []Snapshot      `json:"Snapshots,omitempty"`
}

// implements the service definition of ReadSubnetsRequest
type ReadSubnetsRequest struct {
	DryRun  bool          `json:"DryRun,omitempty"`
	Filters FiltersSubnet `json:"Filters,omitempty"`
}

// implements the service definition of ReadSubnetsResponse
type ReadSubnetsResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	Subnets         []Subnet        `json:"Subnets,omitempty"`
}

// implements the service definition of ReadSubregionsRequest
type ReadSubregionsRequest struct {
	DryRun         bool               `json:"DryRun,omitempty"`
	Filters        []FiltersOldFormat `json:"Filters,omitempty"`
	SubregionNames []string           `json:"SubregionNames,omitempty"`
}

// implements the service definition of ReadSubregionsResponse
type ReadSubregionsResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	Subregions      []Subregion     `json:"Subregions,omitempty"`
}

// implements the service definition of ReadTagsRequest
type ReadTagsRequest struct {
	DryRun  bool       `json:"DryRun,omitempty"`
	Filters FiltersTag `json:"Filters,omitempty"`
}

// implements the service definition of ReadTagsResponse
type ReadTagsResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	Tags            []Tag           `json:"Tags,omitempty"`
}

// implements the service definition of ReadUserGroupsRequest
type ReadUserGroupsRequest struct {
	DryRun  bool             `json:"DryRun,omitempty"`
	Filters FiltersUserGroup `json:"Filters,omitempty"`
}

// implements the service definition of ReadUserGroupsResponse
type ReadUserGroupsResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	UserGroups      []UserGroup     `json:"UserGroups,omitempty"`
}

// implements the service definition of ReadUsersRequest
type ReadUsersRequest struct {
	DryRun bool   `json:"DryRun,omitempty"`
	Path   string `json:"Path,omitempty"`
}

// implements the service definition of ReadUsersResponse
type ReadUsersResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	Users           []User          `json:"Users,omitempty"`
}

// implements the service definition of ReadVirtualGatewaysRequest
type ReadVirtualGatewaysRequest struct {
	DryRun  bool               `json:"DryRun,omitempty"`
	Filters []FiltersOldFormat `json:"Filters,omitempty"`
}

// implements the service definition of ReadVirtualGatewaysResponse
type ReadVirtualGatewaysResponse struct {
	ResponseContext ResponseContext  `json:"ResponseContext,omitempty"`
	VirtualGateways []VirtualGateway `json:"VirtualGateways,omitempty"`
}

// implements the service definition of ReadVmTypesRequest
type ReadVmTypesRequest struct {
	DryRun  bool               `json:"DryRun,omitempty"`
	Filters []FiltersOldFormat `json:"Filters,omitempty"`
}

// implements the service definition of ReadVmTypesResponse
type ReadVmTypesResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	VmTypes         []VmType        `json:"VmTypes,omitempty"`
}

// implements the service definition of ReadVmsHealthRequest
type ReadVmsHealthRequest struct {
	BackendVmsIds    []string `json:"BackendVmsIds,omitempty"`
	DryRun           bool     `json:"DryRun,omitempty"`
	LoadBalancerName string   `json:"LoadBalancerName,omitempty"`
}

// implements the service definition of ReadVmsHealthResponse
type ReadVmsHealthResponse struct {
	BackendVmsHealth []BackendVmsHealth `json:"BackendVmsHealth,omitempty"`
	ResponseContext  ResponseContext    `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReadVmsRequest
type ReadVmsRequest struct {
	DryRun  bool      `json:"DryRun,omitempty"`
	Filters FiltersVm `json:"Filters,omitempty"`
}

// implements the service definition of ReadVmsResponse
type ReadVmsResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	Vms             []Vm            `json:"Vms,omitempty"`
}

// implements the service definition of ReadVmsStateRequest
type ReadVmsStateRequest struct {
	AllVms  bool            `json:"AllVms,omitempty"`
	DryRun  bool            `json:"DryRun,omitempty"`
	Filters FiltersVmsState `json:"Filters,omitempty"`
}

// implements the service definition of ReadVmsStateResponse
type ReadVmsStateResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	VmStates        []VmStates      `json:"VmStates,omitempty"`
}

// implements the service definition of ReadVolumesRequest
type ReadVolumesRequest struct {
	DryRun  bool          `json:"DryRun,omitempty"`
	Filters FiltersVolume `json:"Filters,omitempty"`
}

// implements the service definition of ReadVolumesResponse
type ReadVolumesResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	Volumes         []Volume        `json:"Volumes,omitempty"`
}

// implements the service definition of ReadVpnConnectionsRequest
type ReadVpnConnectionsRequest struct {
	DryRun           bool                 `json:"DryRun,omitempty"`
	Filters          FiltersVpnConnection `json:"Filters,omitempty"`
	VpnConnectionIds []string             `json:"VpnConnectionIds,omitempty"`
}

// implements the service definition of ReadVpnConnectionsResponse
type ReadVpnConnectionsResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	VpnConnections  []VpnConnection `json:"VpnConnections,omitempty"`
}

// implements the service definition of RebootVmsRequest
type RebootVmsRequest struct {
	DryRun bool     `json:"DryRun,omitempty"`
	VmIds  []string `json:"VmIds,omitempty"`
}

// implements the service definition of RebootVmsResponse
type RebootVmsResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of RecurringCharge
type RecurringCharge struct {
	Frequency string `json:"Frequency,omitempty"`
}

// implements the service definition of Region
type Region struct {
	RegionEndpoint string `json:"RegionEndpoint,omitempty"`
	RegionName     string `json:"RegionName,omitempty"`
}

// implements the service definition of RegionConfig
type RegionConfig struct {
	FromDate     string              `json:"FromDate,omitempty"`
	Regions      []RegionDescription `json:"Regions,omitempty"`
	TargetRegion TargetRegion        `json:"TargetRegion,omitempty"`
}

// implements the service definition of RegionDescription
type RegionDescription struct {
	Attributes     []Attribute                   `json:"Attributes,omitempty"`
	Continent      string                        `json:"Continent,omitempty"`
	CurrencyCode   string                        `json:"CurrencyCode,omitempty"`
	Entity         string                        `json:"Entity,omitempty"`
	IsPublic       bool                          `json:"IsPublic,omitempty"`
	IsSynchronized bool                          `json:"IsSynchronized,omitempty"`
	Permissions    []RegionDescriptionPermission `json:"Permissions,omitempty"`
	RegionDomain   string                        `json:"RegionDomain,omitempty"`
	RegionId       string                        `json:"RegionId,omitempty"`
	RegionInstance string                        `json:"RegionInstance,omitempty"`
	RegionName     string                        `json:"RegionName,omitempty"`
	SerialFactor   int64                         `json:"SerialFactor,omitempty"`
	Services       []Service                     `json:"Services,omitempty"`
	SubregionNames []string                      `json:"SubregionNames,omitempty"`
}

// implements the service definition of RegionDescriptionPermission
type RegionDescriptionPermission struct {
	Filter         string `json:"Filter,omitempty"`
	PermissionType string `json:"PermissionType,omitempty"`
}

// implements the service definition of RegisterUserInUserGroupRequest
type RegisterUserInUserGroupRequest struct {
	DryRun        bool   `json:"DryRun,omitempty"`
	UserGroupName string `json:"UserGroupName,omitempty"`
	UserName      string `json:"UserName,omitempty"`
}

// implements the service definition of RegisterUserInUserGroupResponse
type RegisterUserInUserGroupResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of RegisterVmsInLoadBalancerRequest
type RegisterVmsInLoadBalancerRequest struct {
	BackendVmsIds    []string `json:"BackendVmsIds,omitempty"`
	DryRun           bool     `json:"DryRun,omitempty"`
	LoadBalancerName string   `json:"LoadBalancerName,omitempty"`
}

// implements the service definition of RegisterVmsInLoadBalancerResponse
type RegisterVmsInLoadBalancerResponse struct {
	BackendVmsIds   []string        `json:"BackendVmsIds,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of RejectNetPeeringRequest
type RejectNetPeeringRequest struct {
	DryRun       bool   `json:"DryRun,omitempty"`
	NetPeeringId string `json:"NetPeeringId,omitempty"`
}

// implements the service definition of RejectNetPeeringResponse
type RejectNetPeeringResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ReservedVm
type ReservedVm struct {
	CurrencyCode     string            `json:"CurrencyCode,omitempty"`
	OfferingType     string            `json:"OfferingType,omitempty"`
	ProductType      string            `json:"ProductType,omitempty"`
	RecurringCharges []RecurringCharge `json:"RecurringCharges,omitempty"`
	ReservedVmsId    string            `json:"ReservedVmsId,omitempty"`
	State            string            `json:"State,omitempty"`
	SubregionName    string            `json:"SubregionName,omitempty"`
	Tenancy          string            `json:"Tenancy,omitempty"`
	VmCount          int64             `json:"VmCount,omitempty"`
	VmType           string            `json:"VmType,omitempty"`
}

// implements the service definition of ReservedVmsOffer
type ReservedVmsOffer struct {
	CurrencyCode       string            `json:"CurrencyCode,omitempty"`
	Duration           int64             `json:"Duration,omitempty"`
	FixedPrice         int               `json:"FixedPrice,omitempty"`
	OfferingType       string            `json:"OfferingType,omitempty"`
	PricingDetails     []PricingDetail   `json:"PricingDetails,omitempty"`
	ProductType        string            `json:"ProductType,omitempty"`
	RecurringCharges   []RecurringCharge `json:"RecurringCharges,omitempty"`
	ReservedVmsOfferId string            `json:"ReservedVmsOfferId,omitempty"`
	SubregionName      string            `json:"SubregionName,omitempty"`
	Tenancy            string            `json:"Tenancy,omitempty"`
	UsagePrice         int               `json:"UsagePrice,omitempty"`
	VmType             string            `json:"VmType,omitempty"`
}

// implements the service definition of ResetAccountPasswordRequest
type ResetAccountPasswordRequest struct {
	DryRun        bool   `json:"DryRun,omitempty"`
	Password      string `json:"Password,omitempty"`
	PasswordToken string `json:"PasswordToken,omitempty"`
}

// implements the service definition of ResetAccountPasswordResponse
type ResetAccountPasswordResponse struct {
	Email           string          `json:"Email,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ResourceTag
type ResourceTag struct {
	Key   string `json:"Key,omitempty"`
	Value string `json:"Value,omitempty"`
}

// implements the service definition of ResponseContext
type ResponseContext struct {
	RequestId string `json:"RequestId,omitempty"`
}

// implements the service definition of Route
type Route struct {
	CreationMethod          string `json:"CreationMethod,omitempty"`
	DestinationIpRange      string `json:"DestinationIpRange,omitempty"`
	DestinationPrefixListId string `json:"DestinationPrefixListId,omitempty"`
	GatewayId               string `json:"GatewayId,omitempty"`
	NatServiceId            string `json:"NatServiceId,omitempty"`
	NetPeeringId            string `json:"NetPeeringId,omitempty"`
	NicId                   string `json:"NicId,omitempty"`
	State                   string `json:"State,omitempty"`
	VmAccountId             string `json:"VmAccountId,omitempty"`
	VmId                    string `json:"VmId,omitempty"`
}

// implements the service definition of RouteLight
type RouteLight struct {
	DestinationIpRange string `json:"DestinationIpRange,omitempty"`
	RouteType          string `json:"RouteType,omitempty"`
	State              string `json:"State,omitempty"`
}

// implements the service definition of RoutePropagatingVirtualGateway
type RoutePropagatingVirtualGateway struct {
	VirtualGatewayId string `json:"VirtualGatewayId,omitempty"`
}

// implements the service definition of RouteTable
type RouteTable struct {
	LinkRouteTables                 []LinkRouteTable                 `json:"LinkRouteTables,omitempty"`
	NetId                           string                           `json:"NetId,omitempty"`
	RoutePropagatingVirtualGateways []RoutePropagatingVirtualGateway `json:"RoutePropagatingVirtualGateways,omitempty"`
	RouteTableId                    string                           `json:"RouteTableId,omitempty"`
	Routes                          []Route                          `json:"Routes,omitempty"`
	Tags                            []ResourceTag                    `json:"Tags,omitempty"`
}

// implements the service definition of SecurityGroup
type SecurityGroup struct {
	AccountId         string              `json:"AccountId,omitempty"`
	Description       string              `json:"Description,omitempty"`
	InboundRules      []SecurityGroupRule `json:"InboundRules,omitempty"`
	NetId             string              `json:"NetId,omitempty"`
	OutboundRules     []SecurityGroupRule `json:"OutboundRules,omitempty"`
	SecurityGroupId   string              `json:"SecurityGroupId,omitempty"`
	SecurityGroupName string              `json:"SecurityGroupName,omitempty"`
	Tags              []ResourceTag       `json:"Tags,omitempty"`
}

// implements the service definition of SecurityGroupLight
type SecurityGroupLight struct {
	SecurityGroupId   string `json:"SecurityGroupId,omitempty"`
	SecurityGroupName string `json:"SecurityGroupName,omitempty"`
}

// implements the service definition of SecurityGroupRule
type SecurityGroupRule struct {
	FromPortRange         int64                  `json:"FromPortRange,omitempty"`
	IpProtocol            string                 `json:"IpProtocol,omitempty"`
	IpRanges              []string               `json:"IpRanges,omitempty"`
	PrefixListIds         []string               `json:"PrefixListIds,omitempty"`
	SecurityGroupsMembers []SecurityGroupsMember `json:"SecurityGroupsMembers,omitempty"`
	ToPortRange           int64                  `json:"ToPortRange,omitempty"`
}

// implements the service definition of SecurityGroupsMember
type SecurityGroupsMember struct {
	AccountId         string `json:"AccountId,omitempty"`
	SecurityGroupId   string `json:"SecurityGroupId,omitempty"`
	SecurityGroupName string `json:"SecurityGroupName,omitempty"`
}

// implements the service definition of SendResetPasswordEmailRequest
type SendResetPasswordEmailRequest struct {
	DryRun bool   `json:"DryRun,omitempty"`
	Email  string `json:"Email,omitempty"`
}

// implements the service definition of SendResetPasswordEmailResponse
type SendResetPasswordEmailResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of ServerCertificate
type ServerCertificate struct {
	Path                  string `json:"Path,omitempty"`
	ServerCertificateId   string `json:"ServerCertificateId,omitempty"`
	ServerCertificateName string `json:"ServerCertificateName,omitempty"`
}

// implements the service definition of Service
type Service struct {
	Filters     []FiltersServices `json:"Filters,omitempty"`
	ServiceName string            `json:"ServiceName,omitempty"`
	ServiceType string            `json:"ServiceType,omitempty"`
}

// implements the service definition of Snapshot
type Snapshot struct {
	AccountAlias              string                `json:"AccountAlias,omitempty"`
	AccountId                 string                `json:"AccountId,omitempty"`
	Description               string                `json:"Description,omitempty"`
	PermissionsToCreateVolume PermissionsOnResource `json:"PermissionsToCreateVolume,omitempty"`
	Progress                  int64                 `json:"Progress,omitempty"`
	SnapshotId                string                `json:"SnapshotId,omitempty"`
	State                     string                `json:"State,omitempty"`
	Tags                      []ResourceTag         `json:"Tags,omitempty"`
	VolumeId                  string                `json:"VolumeId,omitempty"`
	VolumeSize                int64                 `json:"VolumeSize,omitempty"`
}

// implements the service definition of SnapshotExportTask
type SnapshotExportTask struct {
	Comment    string    `json:"Comment,omitempty"`
	OsuExport  OsuExport `json:"OsuExport,omitempty"`
	Progress   int64     `json:"Progress,omitempty"`
	SnapshotId string    `json:"SnapshotId,omitempty"`
	State      string    `json:"State,omitempty"`
	TaskId     string    `json:"TaskId,omitempty"`
}

// implements the service definition of SourceNet
type SourceNet struct {
	AccountId string `json:"AccountId,omitempty"`
	IpRange   string `json:"IpRange,omitempty"`
	NetId     string `json:"NetId,omitempty"`
}

// implements the service definition of SourceSecurityGroup
type SourceSecurityGroup struct {
	SecurityGroupAccountId string `json:"SecurityGroupAccountId,omitempty"`
	SecurityGroupName      string `json:"SecurityGroupName,omitempty"`
}

// implements the service definition of StartVmsRequest
type StartVmsRequest struct {
	DryRun bool     `json:"DryRun,omitempty"`
	VmIds  []string `json:"VmIds,omitempty"`
}

// implements the service definition of StartVmsResponse
type StartVmsResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	Vms             []VmState       `json:"Vms,omitempty"`
}

// implements the service definition of StateComment
type StateComment struct {
	StateCode    string `json:"StateCode,omitempty"`
	StateMessage string `json:"StateMessage,omitempty"`
}

// implements the service definition of StopVmsRequest
type StopVmsRequest struct {
	DryRun    bool     `json:"DryRun,omitempty"`
	ForceStop bool     `json:"ForceStop,omitempty"`
	VmIds     []string `json:"VmIds,omitempty"`
}

// implements the service definition of StopVmsResponse
type StopVmsResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	Vms             []VmState       `json:"Vms,omitempty"`
}

// implements the service definition of Subnet
type Subnet struct {
	AvailableIpsCount int64         `json:"AvailableIpsCount,omitempty"`
	IpRange           string        `json:"IpRange,omitempty"`
	NetId             string        `json:"NetId,omitempty"`
	State             string        `json:"State,omitempty"`
	SubnetId          string        `json:"SubnetId,omitempty"`
	SubregionName     string        `json:"SubregionName,omitempty"`
	Tags              []ResourceTag `json:"Tags,omitempty"`
}

// implements the service definition of Subregion
type Subregion struct {
	RegionName    string `json:"RegionName,omitempty"`
	State         string `json:"State,omitempty"`
	SubregionName string `json:"SubregionName,omitempty"`
}

// implements the service definition of Tag
type Tag struct {
	Key          string `json:"Key,omitempty"`
	ResourceId   string `json:"ResourceId,omitempty"`
	ResourceType string `json:"ResourceType,omitempty"`
	Value        string `json:"Value,omitempty"`
}

// implements the service definition of TargetRegion
type TargetRegion struct {
	RegionDomain string `json:"RegionDomain,omitempty"`
	RegionId     string `json:"RegionId,omitempty"`
	RegionName   string `json:"RegionName,omitempty"`
}

// implements the service definition of UnlinkInternetServiceRequest
type UnlinkInternetServiceRequest struct {
	DryRun            bool   `json:"DryRun,omitempty"`
	InternetServiceId string `json:"InternetServiceId,omitempty"`
	NetId             string `json:"NetId,omitempty"`
}

// implements the service definition of UnlinkInternetServiceResponse
type UnlinkInternetServiceResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of UnlinkNicRequest
type UnlinkNicRequest struct {
	DryRun    bool   `json:"DryRun,omitempty"`
	LinkNicId string `json:"LinkNicId,omitempty"`
}

// implements the service definition of UnlinkNicResponse
type UnlinkNicResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of UnlinkPolicyRequest
type UnlinkPolicyRequest struct {
	DryRun        bool   `json:"DryRun,omitempty"`
	PolicyId      string `json:"PolicyId,omitempty"`
	UserGroupName string `json:"UserGroupName,omitempty"`
	UserName      string `json:"UserName,omitempty"`
}

// implements the service definition of UnlinkPolicyResponse
type UnlinkPolicyResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of UnlinkPrivateIpsRequest
type UnlinkPrivateIpsRequest struct {
	DryRun     bool     `json:"DryRun,omitempty"`
	NicId      string   `json:"NicId,omitempty"`
	PrivateIps []string `json:"PrivateIps,omitempty"`
}

// implements the service definition of UnlinkPrivateIpsResponse
type UnlinkPrivateIpsResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of UnlinkPublicIpRequest
type UnlinkPublicIpRequest struct {
	DryRun         bool   `json:"DryRun,omitempty"`
	LinkPublicIpId string `json:"LinkPublicIpId,omitempty"`
	PublicIp       string `json:"PublicIp,omitempty"`
}

// implements the service definition of UnlinkPublicIpResponse
type UnlinkPublicIpResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of UnlinkRouteTableRequest
type UnlinkRouteTableRequest struct {
	DryRun           bool   `json:"DryRun,omitempty"`
	LinkRouteTableId string `json:"LinkRouteTableId,omitempty"`
}

// implements the service definition of UnlinkRouteTableResponse
type UnlinkRouteTableResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of UnlinkVirtualGatewayRequest
type UnlinkVirtualGatewayRequest struct {
	DryRun           bool   `json:"DryRun,omitempty"`
	NetId            string `json:"NetId,omitempty"`
	VirtualGatewayId string `json:"VirtualGatewayId,omitempty"`
}

// implements the service definition of UnlinkVirtualGatewayResponse
type UnlinkVirtualGatewayResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of UnlinkVolumeRequest
type UnlinkVolumeRequest struct {
	DeviceName  string `json:"DeviceName,omitempty"`
	DryRun      bool   `json:"DryRun,omitempty"`
	ForceUnlink bool   `json:"ForceUnlink,omitempty"`
	VolumeId    string `json:"VolumeId,omitempty"`
}

// implements the service definition of UnlinkVolumeResponse
type UnlinkVolumeResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of UpdateAccountRequest
type UpdateAccountRequest struct {
	City          string `json:"City,omitempty"`
	CompanyName   string `json:"CompanyName,omitempty"`
	Country       string `json:"Country,omitempty"`
	DryRun        bool   `json:"DryRun,omitempty"`
	Email         string `json:"Email,omitempty"`
	FirstName     string `json:"FirstName,omitempty"`
	JobTitle      string `json:"JobTitle,omitempty"`
	LastName      string `json:"LastName,omitempty"`
	Mobile        string `json:"Mobile,omitempty"`
	Password      string `json:"Password,omitempty"`
	Phone         string `json:"Phone,omitempty"`
	StateProvince string `json:"StateProvince,omitempty"`
	VatNumber     string `json:"VatNumber,omitempty"`
	ZipCode       string `json:"ZipCode,omitempty"`
}

// implements the service definition of UpdateAccountResponse
type UpdateAccountResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of UpdateApiKeyRequest
type UpdateApiKeyRequest struct {
	ApiKeyId string `json:"ApiKeyId,omitempty"`
	DryRun   bool   `json:"DryRun,omitempty"`
	State    string `json:"State,omitempty"`
}

// implements the service definition of UpdateApiKeyResponse
type UpdateApiKeyResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of UpdateHealthCheckRequest
type UpdateHealthCheckRequest struct {
	DryRun           bool        `json:"DryRun,omitempty"`
	HealthCheck      HealthCheck `json:"HealthCheck,omitempty"`
	LoadBalancerName string      `json:"LoadBalancerName,omitempty"`
}

// implements the service definition of UpdateHealthCheckResponse
type UpdateHealthCheckResponse struct {
	HealthCheck     HealthCheck     `json:"HealthCheck,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of UpdateImageRequest
type UpdateImageRequest struct {
	DryRun              bool                          `json:"DryRun,omitempty"`
	ImageId             string                        `json:"ImageId,omitempty"`
	PermissionsToLaunch PermissionsOnResourceCreation `json:"PermissionsToLaunch,omitempty"`
}

// implements the service definition of UpdateImageResponse
type UpdateImageResponse struct {
	Image           Image           `json:"Image,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of UpdateKeypairRequest
type UpdateKeypairRequest struct {
	DryRun      bool   `json:"DryRun,omitempty"`
	KeypairName string `json:"KeypairName,omitempty"`
	PublicKey   string `json:"PublicKey,omitempty"`
}

// implements the service definition of UpdateKeypairResponse
type UpdateKeypairResponse struct {
	KeypairFingerprint string          `json:"KeypairFingerprint,omitempty"`
	KeypairName        string          `json:"KeypairName,omitempty"`
	ResponseContext    ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of UpdateListenerRuleRequest
type UpdateListenerRuleRequest struct {
	Attribute        string `json:"Attribute,omitempty"`
	DryRun           bool   `json:"DryRun,omitempty"`
	ListenerRuleName string `json:"ListenerRuleName,omitempty"`
	Value            string `json:"Value,omitempty"`
}

// implements the service definition of UpdateListenerRuleResponse
type UpdateListenerRuleResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of UpdateLoadBalancerRequest
type UpdateLoadBalancerRequest struct {
	AccessLog           AccessLog `json:"AccessLog,omitempty"`
	DryRun              bool      `json:"DryRun,omitempty"`
	LoadBalancerName    string    `json:"LoadBalancerName,omitempty"`
	LoadBalancerPort    int64     `json:"LoadBalancerPort,omitempty"`
	PolicyNames         []string  `json:"PolicyNames,omitempty"`
	ServerCertificateId string    `json:"ServerCertificateId,omitempty"`
}

// implements the service definition of UpdateLoadBalancerResponse
type UpdateLoadBalancerResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of UpdateNetAccessPointRequest
type UpdateNetAccessPointRequest struct {
	AddRouteTableIds    []string `json:"AddRouteTableIds,omitempty"`
	DryRun              bool     `json:"DryRun,omitempty"`
	NetAccessPointId    string   `json:"NetAccessPointId,omitempty"`
	RemoveRouteTableIds []string `json:"RemoveRouteTableIds,omitempty"`
}

// implements the service definition of UpdateNetAccessPointResponse
type UpdateNetAccessPointResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of UpdateNetRequest
type UpdateNetRequest struct {
	DhcpOptionsSetId string `json:"DhcpOptionsSetId,omitempty"`
	DryRun           bool   `json:"DryRun,omitempty"`
	NetId            string `json:"NetId,omitempty"`
}

// implements the service definition of UpdateNetResponse
type UpdateNetResponse struct {
	Net             Net             `json:"Net,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of UpdateNicRequest
type UpdateNicRequest struct {
	Description      string          `json:"Description,omitempty"`
	DryRun           bool            `json:"DryRun,omitempty"`
	LinkNic          LinkNicToUpdate `json:"LinkNic,omitempty"`
	NicId            string          `json:"NicId,omitempty"`
	SecurityGroupIds []string        `json:"SecurityGroupIds,omitempty"`
}

// implements the service definition of UpdateNicResponse
type UpdateNicResponse struct {
	Nic             Nic             `json:"Nic,omitempty"`
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of UpdateRoutePropagationRequest
type UpdateRoutePropagationRequest struct {
	DryRun           bool   `json:"DryRun,omitempty"`
	Enable           bool   `json:"Enable,omitempty"`
	RouteTableId     string `json:"RouteTableId,omitempty"`
	VirtualGatewayId string `json:"VirtualGatewayId,omitempty"`
}

// implements the service definition of UpdateRoutePropagationResponse
type UpdateRoutePropagationResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	RouteTable      RouteTable      `json:"RouteTable,omitempty"`
}

// implements the service definition of UpdateRouteRequest
type UpdateRouteRequest struct {
	DestinationIpRange string `json:"DestinationIpRange,omitempty"`
	DryRun             bool   `json:"DryRun,omitempty"`
	GatewayId          string `json:"GatewayId,omitempty"`
	NatServiceId       string `json:"NatServiceId,omitempty"`
	NetPeeringId       string `json:"NetPeeringId,omitempty"`
	NicId              string `json:"NicId,omitempty"`
	RouteTableId       string `json:"RouteTableId,omitempty"`
	VmId               string `json:"VmId,omitempty"`
}

// implements the service definition of UpdateRouteResponse
type UpdateRouteResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	Route           Route           `json:"Route,omitempty"`
}

// implements the service definition of UpdateServerCertificateRequest
type UpdateServerCertificateRequest struct {
	DryRun                   bool   `json:"DryRun,omitempty"`
	NewPath                  string `json:"NewPath,omitempty"`
	NewServerCertificateName string `json:"NewServerCertificateName,omitempty"`
	ServerCertificateName    string `json:"ServerCertificateName,omitempty"`
}

// implements the service definition of UpdateServerCertificateResponse
type UpdateServerCertificateResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of UpdateSnapshotRequest
type UpdateSnapshotRequest struct {
	DryRun                    bool                          `json:"DryRun,omitempty"`
	PermissionsToCreateVolume PermissionsOnResourceCreation `json:"PermissionsToCreateVolume,omitempty"`
	SnapshotId                string                        `json:"SnapshotId,omitempty"`
}

// implements the service definition of UpdateSnapshotResponse
type UpdateSnapshotResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	Snapshot        Snapshot        `json:"Snapshot,omitempty"`
}

// implements the service definition of UpdateUserGroupRequest
type UpdateUserGroupRequest struct {
	DryRun           bool   `json:"DryRun,omitempty"`
	NewPath          string `json:"NewPath,omitempty"`
	NewUserGroupName string `json:"NewUserGroupName,omitempty"`
	UserGroupName    string `json:"UserGroupName,omitempty"`
}

// implements the service definition of UpdateUserGroupResponse
type UpdateUserGroupResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of UpdateUserRequest
type UpdateUserRequest struct {
	DryRun      bool   `json:"DryRun,omitempty"`
	NewPath     string `json:"NewPath,omitempty"`
	NewUserName string `json:"NewUserName,omitempty"`
	UserName    string `json:"UserName,omitempty"`
}

// implements the service definition of UpdateUserResponse
type UpdateUserResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
}

// implements the service definition of UpdateVmRequest
type UpdateVmRequest struct {
	BlockDeviceMappings         []BlockDeviceMappingVmUpdate `json:"BlockDeviceMappings,omitempty"`
	BsuOptimized                bool                         `json:"BsuOptimized,omitempty"`
	DeletionProtection          bool                         `json:"DeletionProtection,omitempty"`
	DryRun                      bool                         `json:"DryRun,omitempty"`
	IsSourceDestChecked         bool                         `json:"IsSourceDestChecked,omitempty"`
	KeypairName                 string                       `json:"KeypairName,omitempty"`
	SecurityGroupIds            []string                     `json:"SecurityGroupIds,omitempty"`
	UserData                    string                       `json:"UserData,omitempty"`
	VmId                        string                       `json:"VmId,omitempty"`
	VmInitiatedShutdownBehavior string                       `json:"VmInitiatedShutdownBehavior,omitempty"`
	VmType                      string                       `json:"VmType,omitempty"`
}

// implements the service definition of UpdateVmResponse
type UpdateVmResponse struct {
	ResponseContext ResponseContext `json:"ResponseContext,omitempty"`
	Vm              Vm              `json:"Vm,omitempty"`
}

// implements the service definition of User
type User struct {
	Path     string `json:"Path,omitempty"`
	UserId   string `json:"UserId,omitempty"`
	UserName string `json:"UserName,omitempty"`
}

// implements the service definition of UserGroup
type UserGroup struct {
	Path          string `json:"Path,omitempty"`
	UserGroupId   string `json:"UserGroupId,omitempty"`
	UserGroupName string `json:"UserGroupName,omitempty"`
}

// implements the service definition of VirtualGateway
type VirtualGateway struct {
	ConnectionType           string                    `json:"ConnectionType,omitempty"`
	NetToVirtualGatewayLinks []NetToVirtualGatewayLink `json:"NetToVirtualGatewayLinks,omitempty"`
	State                    string                    `json:"State,omitempty"`
	Tags                     []ResourceTag             `json:"Tags,omitempty"`
	VirtualGatewayId         string                    `json:"VirtualGatewayId,omitempty"`
}

// implements the service definition of Vm
type Vm struct {
	Architecture                string                      `json:"Architecture,omitempty"`
	BlockDeviceMappings         []BlockDeviceMappingCreated `json:"BlockDeviceMappings,omitempty"`
	BsuOptimized                bool                        `json:"BsuOptimized,omitempty"`
	ClientToken                 string                      `json:"ClientToken,omitempty"`
	DeletionProtection          bool                        `json:"DeletionProtection,omitempty"`
	Hypervisor                  string                      `json:"Hypervisor,omitempty"`
	ImageId                     string                      `json:"ImageId,omitempty"`
	IsSourceDestChecked         bool                        `json:"IsSourceDestChecked,omitempty"`
	KeypairName                 string                      `json:"KeypairName,omitempty"`
	LaunchNumber                int64                       `json:"LaunchNumber,omitempty"`
	NetId                       string                      `json:"NetId,omitempty"`
	Nics                        []NicLight                  `json:"Nics,omitempty"`
	OsFamily                    string                      `json:"OsFamily,omitempty"`
	Placement                   Placement                   `json:"Placement,omitempty"`
	PrivateDnsName              string                      `json:"PrivateDnsName,omitempty"`
	PrivateIp                   string                      `json:"PrivateIp,omitempty"`
	ProductCodes                []string                    `json:"ProductCodes,omitempty"`
	PublicDnsName               string                      `json:"PublicDnsName,omitempty"`
	PublicIp                    string                      `json:"PublicIp,omitempty"`
	ReservationId               string                      `json:"ReservationId,omitempty"`
	RootDeviceName              string                      `json:"RootDeviceName,omitempty"`
	RootDeviceType              string                      `json:"RootDeviceType,omitempty"`
	SecurityGroups              []SecurityGroupLight        `json:"SecurityGroups,omitempty"`
	State                       string                      `json:"State,omitempty"`
	StateReason                 string                      `json:"StateReason,omitempty"`
	SubnetId                    string                      `json:"SubnetId,omitempty"`
	Tags                        []ResourceTag               `json:"Tags,omitempty"`
	UserData                    string                      `json:"UserData,omitempty"`
	VmId                        string                      `json:"VmId,omitempty"`
	VmInitiatedShutdownBehavior string                      `json:"VmInitiatedShutdownBehavior,omitempty"`
	VmType                      string                      `json:"VmType,omitempty"`
}

// implements the service definition of VmState
type VmState struct {
	CurrentState  string `json:"CurrentState,omitempty"`
	PreviousState string `json:"PreviousState,omitempty"`
	VmId          string `json:"VmId,omitempty"`
}

// implements the service definition of VmStates
type VmStates struct {
	MaintenanceEvents []MaintenanceEvent `json:"MaintenanceEvents,omitempty"`
	SubregionName     string             `json:"SubregionName,omitempty"`
	VmId              string             `json:"VmId,omitempty"`
	VmState           string             `json:"VmState,omitempty"`
}

// implements the service definition of VmType
type VmType struct {
	IsBsuOptimized bool   `json:"IsBsuOptimized,omitempty"`
	MaxPrivateIps  int64  `json:"MaxPrivateIps,omitempty"`
	MemorySize     int64  `json:"MemorySize,omitempty"`
	StorageCount   int64  `json:"StorageCount,omitempty"`
	StorageSize    int64  `json:"StorageSize,omitempty"`
	VcoreCount     int64  `json:"VcoreCount,omitempty"`
	VmTypeName     string `json:"VmTypeName,omitempty"`
}

// implements the service definition of Volume
type Volume struct {
	Iops          int64          `json:"Iops,omitempty"`
	LinkedVolumes []LinkedVolume `json:"LinkedVolumes,omitempty"`
	Size          int64          `json:"Size,omitempty"`
	SnapshotId    string         `json:"SnapshotId,omitempty"`
	State         string         `json:"State,omitempty"`
	SubregionName string         `json:"SubregionName,omitempty"`
	Tags          []ResourceTag  `json:"Tags,omitempty"`
	VolumeId      string         `json:"VolumeId,omitempty"`
	VolumeType    string         `json:"VolumeType,omitempty"`
}

// implements the service definition of VpnConnection
type VpnConnection struct {
	ClientGatewayConfiguration string        `json:"ClientGatewayConfiguration,omitempty"`
	ClientGatewayId            string        `json:"ClientGatewayId,omitempty"`
	ConnectionType             string        `json:"ConnectionType,omitempty"`
	Routes                     []RouteLight  `json:"Routes,omitempty"`
	State                      string        `json:"State,omitempty"`
	StaticRoutesOnly           bool          `json:"StaticRoutesOnly,omitempty"`
	Tags                       []ResourceTag `json:"Tags,omitempty"`
	VirtualGatewayId           string        `json:"VirtualGatewayId,omitempty"`
	VpnConnectionId            string        `json:"VpnConnectionId,omitempty"`
}

// implements the service definition of With
type With struct {
	CallDuration       bool `json:"CallDuration,omitempty"`
	QueryAccessKey     bool `json:"QueryAccessKey,omitempty"`
	QueryApiName       bool `json:"QueryApiName,omitempty"`
	QueryApiVersion    bool `json:"QueryApiVersion,omitempty"`
	QueryCallName      bool `json:"QueryCallName,omitempty"`
	QueryDate          bool `json:"QueryDate,omitempty"`
	QueryIpAddress     bool `json:"QueryIpAddress,omitempty"`
	QueryRaw           bool `json:"QueryRaw,omitempty"`
	QuerySize          bool `json:"QuerySize,omitempty"`
	QueryUserAgent     bool `json:"QueryUserAgent,omitempty"`
	ResponseId         bool `json:"ResponseId,omitempty"`
	ResponseSize       bool `json:"ResponseSize,omitempty"`
	ResponseStatusCode bool `json:"ResponseStatusCode,omitempty"`
}

// POST_AcceptNetPeeringParameters holds parameters to POST_AcceptNetPeering
type POST_AcceptNetPeeringParameters struct {
	Acceptnetpeeringrequest AcceptNetPeeringRequest `json:"acceptnetpeeringrequest,omitempty"`
}

// POST_AcceptNetPeeringResponses holds responses of POST_AcceptNetPeering
type POST_AcceptNetPeeringResponses struct {
	OK      *AcceptNetPeeringResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code409 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_AuthenticateAccountParameters holds parameters to POST_AuthenticateAccount
type POST_AuthenticateAccountParameters struct {
	Authenticateaccountrequest AuthenticateAccountRequest `json:"authenticateaccountrequest,omitempty"`
}

// POST_AuthenticateAccountResponses holds responses of POST_AuthenticateAccount
type POST_AuthenticateAccountResponses struct {
	OK *AuthenticateAccountResponse
}

// POST_CheckSignatureParameters holds parameters to POST_CheckSignature
type POST_CheckSignatureParameters struct {
	Checksignaturerequest CheckSignatureRequest `json:"checksignaturerequest,omitempty"`
}

// POST_CheckSignatureResponses holds responses of POST_CheckSignature
type POST_CheckSignatureResponses struct {
	OK *CheckSignatureResponse
}

// POST_CopyAccountParameters holds parameters to POST_CopyAccount
type POST_CopyAccountParameters struct {
	Copyaccountrequest CopyAccountRequest `json:"copyaccountrequest,omitempty"`
}

// POST_CopyAccountResponses holds responses of POST_CopyAccount
type POST_CopyAccountResponses struct {
	OK *CopyAccountResponse
}

// POST_CreateAccountParameters holds parameters to POST_CreateAccount
type POST_CreateAccountParameters struct {
	Createaccountrequest CreateAccountRequest `json:"createaccountrequest,omitempty"`
}

// POST_CreateAccountResponses holds responses of POST_CreateAccount
type POST_CreateAccountResponses struct {
	OK *CreateAccountResponse
}

// POST_CreateApiKeyParameters holds parameters to POST_CreateApiKey
type POST_CreateApiKeyParameters struct {
	Createapikeyrequest CreateApiKeyRequest `json:"createapikeyrequest,omitempty"`
}

// POST_CreateApiKeyResponses holds responses of POST_CreateApiKey
type POST_CreateApiKeyResponses struct {
	OK *CreateApiKeyResponse
}

// POST_CreateClientGatewayParameters holds parameters to POST_CreateClientGateway
type POST_CreateClientGatewayParameters struct {
	Createclientgatewayrequest CreateClientGatewayRequest `json:"createclientgatewayrequest,omitempty"`
}

// POST_CreateClientGatewayResponses holds responses of POST_CreateClientGateway
type POST_CreateClientGatewayResponses struct {
	OK *CreateClientGatewayResponse
}

// POST_CreateDhcpOptionsParameters holds parameters to POST_CreateDhcpOptions
type POST_CreateDhcpOptionsParameters struct {
	Createdhcpoptionsrequest CreateDhcpOptionsRequest `json:"createdhcpoptionsrequest,omitempty"`
}

// POST_CreateDhcpOptionsResponses holds responses of POST_CreateDhcpOptions
type POST_CreateDhcpOptionsResponses struct {
	OK *CreateDhcpOptionsResponse
}

// POST_CreateDirectLinkParameters holds parameters to POST_CreateDirectLink
type POST_CreateDirectLinkParameters struct {
	Createdirectlinkrequest CreateDirectLinkRequest `json:"createdirectlinkrequest,omitempty"`
}

// POST_CreateDirectLinkResponses holds responses of POST_CreateDirectLink
type POST_CreateDirectLinkResponses struct {
	OK *CreateDirectLinkResponse
}

// POST_CreateDirectLinkInterfaceParameters holds parameters to POST_CreateDirectLinkInterface
type POST_CreateDirectLinkInterfaceParameters struct {
	Createdirectlinkinterfacerequest CreateDirectLinkInterfaceRequest `json:"createdirectlinkinterfacerequest,omitempty"`
}

// POST_CreateDirectLinkInterfaceResponses holds responses of POST_CreateDirectLinkInterface
type POST_CreateDirectLinkInterfaceResponses struct {
	OK *CreateDirectLinkInterfaceResponse
}

// POST_CreateImageParameters holds parameters to POST_CreateImage
type POST_CreateImageParameters struct {
	Createimagerequest CreateImageRequest `json:"createimagerequest,omitempty"`
}

// POST_CreateImageResponses holds responses of POST_CreateImage
type POST_CreateImageResponses struct {
	OK      *CreateImageResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_CreateImageExportTaskParameters holds parameters to POST_CreateImageExportTask
type POST_CreateImageExportTaskParameters struct {
	Createimageexporttaskrequest CreateImageExportTaskRequest `json:"createimageexporttaskrequest,omitempty"`
}

// POST_CreateImageExportTaskResponses holds responses of POST_CreateImageExportTask
type POST_CreateImageExportTaskResponses struct {
	OK *CreateImageExportTaskResponse
}

// POST_CreateInternetServiceParameters holds parameters to POST_CreateInternetService
type POST_CreateInternetServiceParameters struct {
	Createinternetservicerequest CreateInternetServiceRequest `json:"createinternetservicerequest,omitempty"`
}

// POST_CreateInternetServiceResponses holds responses of POST_CreateInternetService
type POST_CreateInternetServiceResponses struct {
	OK      *CreateInternetServiceResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_CreateKeypairParameters holds parameters to POST_CreateKeypair
type POST_CreateKeypairParameters struct {
	Createkeypairrequest CreateKeypairRequest `json:"createkeypairrequest,omitempty"`
}

// POST_CreateKeypairResponses holds responses of POST_CreateKeypair
type POST_CreateKeypairResponses struct {
	OK      *CreateKeypairResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code409 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_CreateListenerRuleParameters holds parameters to POST_CreateListenerRule
type POST_CreateListenerRuleParameters struct {
	Createlistenerrulerequest CreateListenerRuleRequest `json:"createlistenerrulerequest,omitempty"`
}

// POST_CreateListenerRuleResponses holds responses of POST_CreateListenerRule
type POST_CreateListenerRuleResponses struct {
	OK *CreateListenerRuleResponse
}

// POST_CreateLoadBalancerParameters holds parameters to POST_CreateLoadBalancer
type POST_CreateLoadBalancerParameters struct {
	Createloadbalancerrequest CreateLoadBalancerRequest `json:"createloadbalancerrequest,omitempty"`
}

// POST_CreateLoadBalancerResponses holds responses of POST_CreateLoadBalancer
type POST_CreateLoadBalancerResponses struct {
	OK *CreateLoadBalancerResponse
}

// POST_CreateLoadBalancerListenersParameters holds parameters to POST_CreateLoadBalancerListeners
type POST_CreateLoadBalancerListenersParameters struct {
	Createloadbalancerlistenersrequest CreateLoadBalancerListenersRequest `json:"createloadbalancerlistenersrequest,omitempty"`
}

// POST_CreateLoadBalancerListenersResponses holds responses of POST_CreateLoadBalancerListeners
type POST_CreateLoadBalancerListenersResponses struct {
	OK *CreateLoadBalancerListenersResponse
}

// POST_CreateLoadBalancerPolicyParameters holds parameters to POST_CreateLoadBalancerPolicy
type POST_CreateLoadBalancerPolicyParameters struct {
	Createloadbalancerpolicyrequest CreateLoadBalancerPolicyRequest `json:"createloadbalancerpolicyrequest,omitempty"`
}

// POST_CreateLoadBalancerPolicyResponses holds responses of POST_CreateLoadBalancerPolicy
type POST_CreateLoadBalancerPolicyResponses struct {
	OK *CreateLoadBalancerPolicyResponse
}

// POST_CreateNatServiceParameters holds parameters to POST_CreateNatService
type POST_CreateNatServiceParameters struct {
	Createnatservicerequest CreateNatServiceRequest `json:"createnatservicerequest,omitempty"`
}

// POST_CreateNatServiceResponses holds responses of POST_CreateNatService
type POST_CreateNatServiceResponses struct {
	OK      *CreateNatServiceResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_CreateNetParameters holds parameters to POST_CreateNet
type POST_CreateNetParameters struct {
	Createnetrequest CreateNetRequest `json:"createnetrequest,omitempty"`
}

// POST_CreateNetResponses holds responses of POST_CreateNet
type POST_CreateNetResponses struct {
	OK      *CreateNetResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code409 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_CreateNetAccessPointParameters holds parameters to POST_CreateNetAccessPoint
type POST_CreateNetAccessPointParameters struct {
	Createnetaccesspointrequest CreateNetAccessPointRequest `json:"createnetaccesspointrequest,omitempty"`
}

// POST_CreateNetAccessPointResponses holds responses of POST_CreateNetAccessPoint
type POST_CreateNetAccessPointResponses struct {
	OK *CreateNetAccessPointResponse
}

// POST_CreateNetPeeringParameters holds parameters to POST_CreateNetPeering
type POST_CreateNetPeeringParameters struct {
	Createnetpeeringrequest CreateNetPeeringRequest `json:"createnetpeeringrequest,omitempty"`
}

// POST_CreateNetPeeringResponses holds responses of POST_CreateNetPeering
type POST_CreateNetPeeringResponses struct {
	OK      *CreateNetPeeringResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_CreateNicParameters holds parameters to POST_CreateNic
type POST_CreateNicParameters struct {
	Createnicrequest CreateNicRequest `json:"createnicrequest,omitempty"`
}

// POST_CreateNicResponses holds responses of POST_CreateNic
type POST_CreateNicResponses struct {
	OK      *CreateNicResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_CreatePolicyParameters holds parameters to POST_CreatePolicy
type POST_CreatePolicyParameters struct {
	Createpolicyrequest CreatePolicyRequest `json:"createpolicyrequest,omitempty"`
}

// POST_CreatePolicyResponses holds responses of POST_CreatePolicy
type POST_CreatePolicyResponses struct {
	OK *CreatePolicyResponse
}

// POST_CreatePublicIpParameters holds parameters to POST_CreatePublicIp
type POST_CreatePublicIpParameters struct {
	Createpubliciprequest CreatePublicIpRequest `json:"createpubliciprequest,omitempty"`
}

// POST_CreatePublicIpResponses holds responses of POST_CreatePublicIp
type POST_CreatePublicIpResponses struct {
	OK      *CreatePublicIpResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_CreateRouteParameters holds parameters to POST_CreateRoute
type POST_CreateRouteParameters struct {
	Createrouterequest CreateRouteRequest `json:"createrouterequest,omitempty"`
}

// POST_CreateRouteResponses holds responses of POST_CreateRoute
type POST_CreateRouteResponses struct {
	OK      *CreateRouteResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_CreateRouteTableParameters holds parameters to POST_CreateRouteTable
type POST_CreateRouteTableParameters struct {
	Createroutetablerequest CreateRouteTableRequest `json:"createroutetablerequest,omitempty"`
}

// POST_CreateRouteTableResponses holds responses of POST_CreateRouteTable
type POST_CreateRouteTableResponses struct {
	OK      *CreateRouteTableResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_CreateSecurityGroupParameters holds parameters to POST_CreateSecurityGroup
type POST_CreateSecurityGroupParameters struct {
	Createsecuritygrouprequest CreateSecurityGroupRequest `json:"createsecuritygrouprequest,omitempty"`
}

// POST_CreateSecurityGroupResponses holds responses of POST_CreateSecurityGroup
type POST_CreateSecurityGroupResponses struct {
	OK      *CreateSecurityGroupResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_CreateSecurityGroupRuleParameters holds parameters to POST_CreateSecurityGroupRule
type POST_CreateSecurityGroupRuleParameters struct {
	Createsecuritygrouprulerequest CreateSecurityGroupRuleRequest `json:"createsecuritygrouprulerequest,omitempty"`
}

// POST_CreateSecurityGroupRuleResponses holds responses of POST_CreateSecurityGroupRule
type POST_CreateSecurityGroupRuleResponses struct {
	OK      *CreateSecurityGroupRuleResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_CreateServerCertificateParameters holds parameters to POST_CreateServerCertificate
type POST_CreateServerCertificateParameters struct {
	Createservercertificaterequest CreateServerCertificateRequest `json:"createservercertificaterequest,omitempty"`
}

// POST_CreateServerCertificateResponses holds responses of POST_CreateServerCertificate
type POST_CreateServerCertificateResponses struct {
	OK *CreateServerCertificateResponse
}

// POST_CreateSnapshotParameters holds parameters to POST_CreateSnapshot
type POST_CreateSnapshotParameters struct {
	Createsnapshotrequest CreateSnapshotRequest `json:"createsnapshotrequest,omitempty"`
}

// POST_CreateSnapshotResponses holds responses of POST_CreateSnapshot
type POST_CreateSnapshotResponses struct {
	OK      *CreateSnapshotResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_CreateSnapshotExportTaskParameters holds parameters to POST_CreateSnapshotExportTask
type POST_CreateSnapshotExportTaskParameters struct {
	Createsnapshotexporttaskrequest CreateSnapshotExportTaskRequest `json:"createsnapshotexporttaskrequest,omitempty"`
}

// POST_CreateSnapshotExportTaskResponses holds responses of POST_CreateSnapshotExportTask
type POST_CreateSnapshotExportTaskResponses struct {
	OK *CreateSnapshotExportTaskResponse
}

// POST_CreateSubnetParameters holds parameters to POST_CreateSubnet
type POST_CreateSubnetParameters struct {
	Createsubnetrequest CreateSubnetRequest `json:"createsubnetrequest,omitempty"`
}

// POST_CreateSubnetResponses holds responses of POST_CreateSubnet
type POST_CreateSubnetResponses struct {
	OK      *CreateSubnetResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code409 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_CreateTagsParameters holds parameters to POST_CreateTags
type POST_CreateTagsParameters struct {
	Createtagsrequest CreateTagsRequest `json:"createtagsrequest,omitempty"`
}

// POST_CreateTagsResponses holds responses of POST_CreateTags
type POST_CreateTagsResponses struct {
	OK      *CreateTagsResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_CreateUserParameters holds parameters to POST_CreateUser
type POST_CreateUserParameters struct {
	Createuserrequest CreateUserRequest `json:"createuserrequest,omitempty"`
}

// POST_CreateUserResponses holds responses of POST_CreateUser
type POST_CreateUserResponses struct {
	OK *CreateUserResponse
}

// POST_CreateUserGroupParameters holds parameters to POST_CreateUserGroup
type POST_CreateUserGroupParameters struct {
	Createusergrouprequest CreateUserGroupRequest `json:"createusergrouprequest,omitempty"`
}

// POST_CreateUserGroupResponses holds responses of POST_CreateUserGroup
type POST_CreateUserGroupResponses struct {
	OK *CreateUserGroupResponse
}

// POST_CreateVirtualGatewayParameters holds parameters to POST_CreateVirtualGateway
type POST_CreateVirtualGatewayParameters struct {
	Createvirtualgatewayrequest CreateVirtualGatewayRequest `json:"createvirtualgatewayrequest,omitempty"`
}

// POST_CreateVirtualGatewayResponses holds responses of POST_CreateVirtualGateway
type POST_CreateVirtualGatewayResponses struct {
	OK *CreateVirtualGatewayResponse
}

// POST_CreateVmsParameters holds parameters to POST_CreateVms
type POST_CreateVmsParameters struct {
	Createvmsrequest CreateVmsRequest `json:"createvmsrequest,omitempty"`
}

// POST_CreateVmsResponses holds responses of POST_CreateVms
type POST_CreateVmsResponses struct {
	OK      *CreateVmsResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_CreateVolumeParameters holds parameters to POST_CreateVolume
type POST_CreateVolumeParameters struct {
	Createvolumerequest CreateVolumeRequest `json:"createvolumerequest,omitempty"`
}

// POST_CreateVolumeResponses holds responses of POST_CreateVolume
type POST_CreateVolumeResponses struct {
	OK      *CreateVolumeResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_CreateVpnConnectionParameters holds parameters to POST_CreateVpnConnection
type POST_CreateVpnConnectionParameters struct {
	Createvpnconnectionrequest CreateVpnConnectionRequest `json:"createvpnconnectionrequest,omitempty"`
}

// POST_CreateVpnConnectionResponses holds responses of POST_CreateVpnConnection
type POST_CreateVpnConnectionResponses struct {
	OK *CreateVpnConnectionResponse
}

// POST_CreateVpnConnectionRouteParameters holds parameters to POST_CreateVpnConnectionRoute
type POST_CreateVpnConnectionRouteParameters struct {
	Createvpnconnectionrouterequest CreateVpnConnectionRouteRequest `json:"createvpnconnectionrouterequest,omitempty"`
}

// POST_CreateVpnConnectionRouteResponses holds responses of POST_CreateVpnConnectionRoute
type POST_CreateVpnConnectionRouteResponses struct {
	OK *CreateVpnConnectionRouteResponse
}

// POST_DeleteApiKeyParameters holds parameters to POST_DeleteApiKey
type POST_DeleteApiKeyParameters struct {
	Deleteapikeyrequest DeleteApiKeyRequest `json:"deleteapikeyrequest,omitempty"`
}

// POST_DeleteApiKeyResponses holds responses of POST_DeleteApiKey
type POST_DeleteApiKeyResponses struct {
	OK *DeleteApiKeyResponse
}

// POST_DeleteClientGatewayParameters holds parameters to POST_DeleteClientGateway
type POST_DeleteClientGatewayParameters struct {
	Deleteclientgatewayrequest DeleteClientGatewayRequest `json:"deleteclientgatewayrequest,omitempty"`
}

// POST_DeleteClientGatewayResponses holds responses of POST_DeleteClientGateway
type POST_DeleteClientGatewayResponses struct {
	OK *DeleteClientGatewayResponse
}

// POST_DeleteDhcpOptionsParameters holds parameters to POST_DeleteDhcpOptions
type POST_DeleteDhcpOptionsParameters struct {
	Deletedhcpoptionsrequest DeleteDhcpOptionsRequest `json:"deletedhcpoptionsrequest,omitempty"`
}

// POST_DeleteDhcpOptionsResponses holds responses of POST_DeleteDhcpOptions
type POST_DeleteDhcpOptionsResponses struct {
	OK *DeleteDhcpOptionsResponse
}

// POST_DeleteDirectLinkParameters holds parameters to POST_DeleteDirectLink
type POST_DeleteDirectLinkParameters struct {
	Deletedirectlinkrequest DeleteDirectLinkRequest `json:"deletedirectlinkrequest,omitempty"`
}

// POST_DeleteDirectLinkResponses holds responses of POST_DeleteDirectLink
type POST_DeleteDirectLinkResponses struct {
	OK *DeleteDirectLinkResponse
}

// POST_DeleteDirectLinkInterfaceParameters holds parameters to POST_DeleteDirectLinkInterface
type POST_DeleteDirectLinkInterfaceParameters struct {
	Deletedirectlinkinterfacerequest DeleteDirectLinkInterfaceRequest `json:"deletedirectlinkinterfacerequest,omitempty"`
}

// POST_DeleteDirectLinkInterfaceResponses holds responses of POST_DeleteDirectLinkInterface
type POST_DeleteDirectLinkInterfaceResponses struct {
	OK *DeleteDirectLinkInterfaceResponse
}

// POST_DeleteExportTaskParameters holds parameters to POST_DeleteExportTask
type POST_DeleteExportTaskParameters struct {
	Deleteexporttaskrequest DeleteExportTaskRequest `json:"deleteexporttaskrequest,omitempty"`
}

// POST_DeleteExportTaskResponses holds responses of POST_DeleteExportTask
type POST_DeleteExportTaskResponses struct {
	OK *DeleteExportTaskResponse
}

// POST_DeleteImageParameters holds parameters to POST_DeleteImage
type POST_DeleteImageParameters struct {
	Deleteimagerequest DeleteImageRequest `json:"deleteimagerequest,omitempty"`
}

// POST_DeleteImageResponses holds responses of POST_DeleteImage
type POST_DeleteImageResponses struct {
	OK      *DeleteImageResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_DeleteInternetServiceParameters holds parameters to POST_DeleteInternetService
type POST_DeleteInternetServiceParameters struct {
	Deleteinternetservicerequest DeleteInternetServiceRequest `json:"deleteinternetservicerequest,omitempty"`
}

// POST_DeleteInternetServiceResponses holds responses of POST_DeleteInternetService
type POST_DeleteInternetServiceResponses struct {
	OK      *DeleteInternetServiceResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_DeleteKeypairParameters holds parameters to POST_DeleteKeypair
type POST_DeleteKeypairParameters struct {
	Deletekeypairrequest DeleteKeypairRequest `json:"deletekeypairrequest,omitempty"`
}

// POST_DeleteKeypairResponses holds responses of POST_DeleteKeypair
type POST_DeleteKeypairResponses struct {
	OK      *DeleteKeypairResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_DeleteListenerRuleParameters holds parameters to POST_DeleteListenerRule
type POST_DeleteListenerRuleParameters struct {
	Deletelistenerrulerequest DeleteListenerRuleRequest `json:"deletelistenerrulerequest,omitempty"`
}

// POST_DeleteListenerRuleResponses holds responses of POST_DeleteListenerRule
type POST_DeleteListenerRuleResponses struct {
	OK *DeleteListenerRuleResponse
}

// POST_DeleteLoadBalancerParameters holds parameters to POST_DeleteLoadBalancer
type POST_DeleteLoadBalancerParameters struct {
	Deleteloadbalancerrequest DeleteLoadBalancerRequest `json:"deleteloadbalancerrequest,omitempty"`
}

// POST_DeleteLoadBalancerResponses holds responses of POST_DeleteLoadBalancer
type POST_DeleteLoadBalancerResponses struct {
	OK *DeleteLoadBalancerResponse
}

// POST_DeleteLoadBalancerListenersParameters holds parameters to POST_DeleteLoadBalancerListeners
type POST_DeleteLoadBalancerListenersParameters struct {
	Deleteloadbalancerlistenersrequest DeleteLoadBalancerListenersRequest `json:"deleteloadbalancerlistenersrequest,omitempty"`
}

// POST_DeleteLoadBalancerListenersResponses holds responses of POST_DeleteLoadBalancerListeners
type POST_DeleteLoadBalancerListenersResponses struct {
	OK *DeleteLoadBalancerListenersResponse
}

// POST_DeleteLoadBalancerPolicyParameters holds parameters to POST_DeleteLoadBalancerPolicy
type POST_DeleteLoadBalancerPolicyParameters struct {
	Deleteloadbalancerpolicyrequest DeleteLoadBalancerPolicyRequest `json:"deleteloadbalancerpolicyrequest,omitempty"`
}

// POST_DeleteLoadBalancerPolicyResponses holds responses of POST_DeleteLoadBalancerPolicy
type POST_DeleteLoadBalancerPolicyResponses struct {
	OK *DeleteLoadBalancerPolicyResponse
}

// POST_DeleteNatServiceParameters holds parameters to POST_DeleteNatService
type POST_DeleteNatServiceParameters struct {
	Deletenatservicerequest DeleteNatServiceRequest `json:"deletenatservicerequest,omitempty"`
}

// POST_DeleteNatServiceResponses holds responses of POST_DeleteNatService
type POST_DeleteNatServiceResponses struct {
	OK      *DeleteNatServiceResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_DeleteNetParameters holds parameters to POST_DeleteNet
type POST_DeleteNetParameters struct {
	Deletenetrequest DeleteNetRequest `json:"deletenetrequest,omitempty"`
}

// POST_DeleteNetResponses holds responses of POST_DeleteNet
type POST_DeleteNetResponses struct {
	OK      *DeleteNetResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_DeleteNetAccessPointsParameters holds parameters to POST_DeleteNetAccessPoints
type POST_DeleteNetAccessPointsParameters struct {
	Deletenetaccesspointsrequest DeleteNetAccessPointsRequest `json:"deletenetaccesspointsrequest,omitempty"`
}

// POST_DeleteNetAccessPointsResponses holds responses of POST_DeleteNetAccessPoints
type POST_DeleteNetAccessPointsResponses struct {
	OK *DeleteNetAccessPointsResponse
}

// POST_DeleteNetPeeringParameters holds parameters to POST_DeleteNetPeering
type POST_DeleteNetPeeringParameters struct {
	Deletenetpeeringrequest DeleteNetPeeringRequest `json:"deletenetpeeringrequest,omitempty"`
}

// POST_DeleteNetPeeringResponses holds responses of POST_DeleteNetPeering
type POST_DeleteNetPeeringResponses struct {
	OK      *DeleteNetPeeringResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code409 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_DeleteNicParameters holds parameters to POST_DeleteNic
type POST_DeleteNicParameters struct {
	Deletenicrequest DeleteNicRequest `json:"deletenicrequest,omitempty"`
}

// POST_DeleteNicResponses holds responses of POST_DeleteNic
type POST_DeleteNicResponses struct {
	OK      *DeleteNicResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_DeletePolicyParameters holds parameters to POST_DeletePolicy
type POST_DeletePolicyParameters struct {
	Deletepolicyrequest DeletePolicyRequest `json:"deletepolicyrequest,omitempty"`
}

// POST_DeletePolicyResponses holds responses of POST_DeletePolicy
type POST_DeletePolicyResponses struct {
	OK *DeletePolicyResponse
}

// POST_DeletePublicIpParameters holds parameters to POST_DeletePublicIp
type POST_DeletePublicIpParameters struct {
	Deletepubliciprequest DeletePublicIpRequest `json:"deletepubliciprequest,omitempty"`
}

// POST_DeletePublicIpResponses holds responses of POST_DeletePublicIp
type POST_DeletePublicIpResponses struct {
	OK      *DeletePublicIpResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_DeleteRouteParameters holds parameters to POST_DeleteRoute
type POST_DeleteRouteParameters struct {
	Deleterouterequest DeleteRouteRequest `json:"deleterouterequest,omitempty"`
}

// POST_DeleteRouteResponses holds responses of POST_DeleteRoute
type POST_DeleteRouteResponses struct {
	OK      *DeleteRouteResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_DeleteRouteTableParameters holds parameters to POST_DeleteRouteTable
type POST_DeleteRouteTableParameters struct {
	Deleteroutetablerequest DeleteRouteTableRequest `json:"deleteroutetablerequest,omitempty"`
}

// POST_DeleteRouteTableResponses holds responses of POST_DeleteRouteTable
type POST_DeleteRouteTableResponses struct {
	OK      *DeleteRouteTableResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_DeleteSecurityGroupParameters holds parameters to POST_DeleteSecurityGroup
type POST_DeleteSecurityGroupParameters struct {
	Deletesecuritygrouprequest DeleteSecurityGroupRequest `json:"deletesecuritygrouprequest,omitempty"`
}

// POST_DeleteSecurityGroupResponses holds responses of POST_DeleteSecurityGroup
type POST_DeleteSecurityGroupResponses struct {
	OK      *DeleteSecurityGroupResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_DeleteSecurityGroupRuleParameters holds parameters to POST_DeleteSecurityGroupRule
type POST_DeleteSecurityGroupRuleParameters struct {
	Deletesecuritygrouprulerequest DeleteSecurityGroupRuleRequest `json:"deletesecuritygrouprulerequest,omitempty"`
}

// POST_DeleteSecurityGroupRuleResponses holds responses of POST_DeleteSecurityGroupRule
type POST_DeleteSecurityGroupRuleResponses struct {
	OK      *DeleteSecurityGroupRuleResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_DeleteServerCertificateParameters holds parameters to POST_DeleteServerCertificate
type POST_DeleteServerCertificateParameters struct {
	Deleteservercertificaterequest DeleteServerCertificateRequest `json:"deleteservercertificaterequest,omitempty"`
}

// POST_DeleteServerCertificateResponses holds responses of POST_DeleteServerCertificate
type POST_DeleteServerCertificateResponses struct {
	OK *DeleteServerCertificateResponse
}

// POST_DeleteSnapshotParameters holds parameters to POST_DeleteSnapshot
type POST_DeleteSnapshotParameters struct {
	Deletesnapshotrequest DeleteSnapshotRequest `json:"deletesnapshotrequest,omitempty"`
}

// POST_DeleteSnapshotResponses holds responses of POST_DeleteSnapshot
type POST_DeleteSnapshotResponses struct {
	OK      *DeleteSnapshotResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_DeleteSubnetParameters holds parameters to POST_DeleteSubnet
type POST_DeleteSubnetParameters struct {
	Deletesubnetrequest DeleteSubnetRequest `json:"deletesubnetrequest,omitempty"`
}

// POST_DeleteSubnetResponses holds responses of POST_DeleteSubnet
type POST_DeleteSubnetResponses struct {
	OK      *DeleteSubnetResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_DeleteTagsParameters holds parameters to POST_DeleteTags
type POST_DeleteTagsParameters struct {
	Deletetagsrequest DeleteTagsRequest `json:"deletetagsrequest,omitempty"`
}

// POST_DeleteTagsResponses holds responses of POST_DeleteTags
type POST_DeleteTagsResponses struct {
	OK      *DeleteTagsResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_DeleteUserParameters holds parameters to POST_DeleteUser
type POST_DeleteUserParameters struct {
	Deleteuserrequest DeleteUserRequest `json:"deleteuserrequest,omitempty"`
}

// POST_DeleteUserResponses holds responses of POST_DeleteUser
type POST_DeleteUserResponses struct {
	OK *DeleteUserResponse
}

// POST_DeleteUserGroupParameters holds parameters to POST_DeleteUserGroup
type POST_DeleteUserGroupParameters struct {
	Deleteusergrouprequest DeleteUserGroupRequest `json:"deleteusergrouprequest,omitempty"`
}

// POST_DeleteUserGroupResponses holds responses of POST_DeleteUserGroup
type POST_DeleteUserGroupResponses struct {
	OK *DeleteUserGroupResponse
}

// POST_DeleteVirtualGatewayParameters holds parameters to POST_DeleteVirtualGateway
type POST_DeleteVirtualGatewayParameters struct {
	Deletevirtualgatewayrequest DeleteVirtualGatewayRequest `json:"deletevirtualgatewayrequest,omitempty"`
}

// POST_DeleteVirtualGatewayResponses holds responses of POST_DeleteVirtualGateway
type POST_DeleteVirtualGatewayResponses struct {
	OK *DeleteVirtualGatewayResponse
}

// POST_DeleteVmsParameters holds parameters to POST_DeleteVms
type POST_DeleteVmsParameters struct {
	Deletevmsrequest DeleteVmsRequest `json:"deletevmsrequest,omitempty"`
}

// POST_DeleteVmsResponses holds responses of POST_DeleteVms
type POST_DeleteVmsResponses struct {
	OK      *DeleteVmsResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_DeleteVolumeParameters holds parameters to POST_DeleteVolume
type POST_DeleteVolumeParameters struct {
	Deletevolumerequest DeleteVolumeRequest `json:"deletevolumerequest,omitempty"`
}

// POST_DeleteVolumeResponses holds responses of POST_DeleteVolume
type POST_DeleteVolumeResponses struct {
	OK      *DeleteVolumeResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_DeleteVpnConnectionParameters holds parameters to POST_DeleteVpnConnection
type POST_DeleteVpnConnectionParameters struct {
	Deletevpnconnectionrequest DeleteVpnConnectionRequest `json:"deletevpnconnectionrequest,omitempty"`
}

// POST_DeleteVpnConnectionResponses holds responses of POST_DeleteVpnConnection
type POST_DeleteVpnConnectionResponses struct {
	OK *DeleteVpnConnectionResponse
}

// POST_DeleteVpnConnectionRouteParameters holds parameters to POST_DeleteVpnConnectionRoute
type POST_DeleteVpnConnectionRouteParameters struct {
	Deletevpnconnectionrouterequest DeleteVpnConnectionRouteRequest `json:"deletevpnconnectionrouterequest,omitempty"`
}

// POST_DeleteVpnConnectionRouteResponses holds responses of POST_DeleteVpnConnectionRoute
type POST_DeleteVpnConnectionRouteResponses struct {
	OK *DeleteVpnConnectionRouteResponse
}

// POST_DeregisterUserInUserGroupParameters holds parameters to POST_DeregisterUserInUserGroup
type POST_DeregisterUserInUserGroupParameters struct {
	Deregisteruserinusergrouprequest DeregisterUserInUserGroupRequest `json:"deregisteruserinusergrouprequest,omitempty"`
}

// POST_DeregisterUserInUserGroupResponses holds responses of POST_DeregisterUserInUserGroup
type POST_DeregisterUserInUserGroupResponses struct {
	OK *DeregisterUserInUserGroupResponse
}

// POST_DeregisterVmsInLoadBalancerParameters holds parameters to POST_DeregisterVmsInLoadBalancer
type POST_DeregisterVmsInLoadBalancerParameters struct {
	Deregistervmsinloadbalancerrequest DeregisterVmsInLoadBalancerRequest `json:"deregistervmsinloadbalancerrequest,omitempty"`
}

// POST_DeregisterVmsInLoadBalancerResponses holds responses of POST_DeregisterVmsInLoadBalancer
type POST_DeregisterVmsInLoadBalancerResponses struct {
	OK *DeregisterVmsInLoadBalancerResponse
}

// POST_LinkInternetServiceParameters holds parameters to POST_LinkInternetService
type POST_LinkInternetServiceParameters struct {
	Linkinternetservicerequest LinkInternetServiceRequest `json:"linkinternetservicerequest,omitempty"`
}

// POST_LinkInternetServiceResponses holds responses of POST_LinkInternetService
type POST_LinkInternetServiceResponses struct {
	OK      *LinkInternetServiceResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_LinkNicParameters holds parameters to POST_LinkNic
type POST_LinkNicParameters struct {
	Linknicrequest LinkNicRequest `json:"linknicrequest,omitempty"`
}

// POST_LinkNicResponses holds responses of POST_LinkNic
type POST_LinkNicResponses struct {
	OK      *LinkNicResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_LinkPolicyParameters holds parameters to POST_LinkPolicy
type POST_LinkPolicyParameters struct {
	Linkpolicyrequest LinkPolicyRequest `json:"linkpolicyrequest,omitempty"`
}

// POST_LinkPolicyResponses holds responses of POST_LinkPolicy
type POST_LinkPolicyResponses struct {
	OK *LinkPolicyResponse
}

// POST_LinkPrivateIpsParameters holds parameters to POST_LinkPrivateIps
type POST_LinkPrivateIpsParameters struct {
	Linkprivateipsrequest LinkPrivateIpsRequest `json:"linkprivateipsrequest,omitempty"`
}

// POST_LinkPrivateIpsResponses holds responses of POST_LinkPrivateIps
type POST_LinkPrivateIpsResponses struct {
	OK      *LinkPrivateIpsResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_LinkPublicIpParameters holds parameters to POST_LinkPublicIp
type POST_LinkPublicIpParameters struct {
	Linkpubliciprequest LinkPublicIpRequest `json:"linkpubliciprequest,omitempty"`
}

// POST_LinkPublicIpResponses holds responses of POST_LinkPublicIp
type POST_LinkPublicIpResponses struct {
	OK      *LinkPublicIpResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_LinkRouteTableParameters holds parameters to POST_LinkRouteTable
type POST_LinkRouteTableParameters struct {
	Linkroutetablerequest LinkRouteTableRequest `json:"linkroutetablerequest,omitempty"`
}

// POST_LinkRouteTableResponses holds responses of POST_LinkRouteTable
type POST_LinkRouteTableResponses struct {
	OK      *LinkRouteTableResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_LinkVirtualGatewayParameters holds parameters to POST_LinkVirtualGateway
type POST_LinkVirtualGatewayParameters struct {
	Linkvirtualgatewayrequest LinkVirtualGatewayRequest `json:"linkvirtualgatewayrequest,omitempty"`
}

// POST_LinkVirtualGatewayResponses holds responses of POST_LinkVirtualGateway
type POST_LinkVirtualGatewayResponses struct {
	OK *LinkVirtualGatewayResponse
}

// POST_LinkVolumeParameters holds parameters to POST_LinkVolume
type POST_LinkVolumeParameters struct {
	Linkvolumerequest LinkVolumeRequest `json:"linkvolumerequest,omitempty"`
}

// POST_LinkVolumeResponses holds responses of POST_LinkVolume
type POST_LinkVolumeResponses struct {
	OK      *LinkVolumeResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_PurchaseReservedVmsOfferParameters holds parameters to POST_PurchaseReservedVmsOffer
type POST_PurchaseReservedVmsOfferParameters struct {
	Purchasereservedvmsofferrequest PurchaseReservedVmsOfferRequest `json:"purchasereservedvmsofferrequest,omitempty"`
}

// POST_PurchaseReservedVmsOfferResponses holds responses of POST_PurchaseReservedVmsOffer
type POST_PurchaseReservedVmsOfferResponses struct {
	OK *PurchaseReservedVmsOfferResponse
}

// POST_ReadAccountParameters holds parameters to POST_ReadAccount
type POST_ReadAccountParameters struct {
	Readaccountrequest ReadAccountRequest `json:"readaccountrequest,omitempty"`
}

// POST_ReadAccountResponses holds responses of POST_ReadAccount
type POST_ReadAccountResponses struct {
	OK *ReadAccountResponse
}

// POST_ReadAccountConsumptionParameters holds parameters to POST_ReadAccountConsumption
type POST_ReadAccountConsumptionParameters struct {
	Readaccountconsumptionrequest ReadAccountConsumptionRequest `json:"readaccountconsumptionrequest,omitempty"`
}

// POST_ReadAccountConsumptionResponses holds responses of POST_ReadAccountConsumption
type POST_ReadAccountConsumptionResponses struct {
	OK *ReadAccountConsumptionResponse
}

// POST_ReadAdminPasswordParameters holds parameters to POST_ReadAdminPassword
type POST_ReadAdminPasswordParameters struct {
	Readadminpasswordrequest ReadAdminPasswordRequest `json:"readadminpasswordrequest,omitempty"`
}

// POST_ReadAdminPasswordResponses holds responses of POST_ReadAdminPassword
type POST_ReadAdminPasswordResponses struct {
	OK      *ReadAdminPasswordResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_ReadApiKeysParameters holds parameters to POST_ReadApiKeys
type POST_ReadApiKeysParameters struct {
	Readapikeysrequest ReadApiKeysRequest `json:"readapikeysrequest,omitempty"`
}

// POST_ReadApiKeysResponses holds responses of POST_ReadApiKeys
type POST_ReadApiKeysResponses struct {
	OK *ReadApiKeysResponse
}

// POST_ReadApiLogsParameters holds parameters to POST_ReadApiLogs
type POST_ReadApiLogsParameters struct {
	Readapilogsrequest ReadApiLogsRequest `json:"readapilogsrequest,omitempty"`
}

// POST_ReadApiLogsResponses holds responses of POST_ReadApiLogs
type POST_ReadApiLogsResponses struct {
	OK *ReadApiLogsResponse
}

// POST_ReadBillableDigestParameters holds parameters to POST_ReadBillableDigest
type POST_ReadBillableDigestParameters struct {
	Readbillabledigestrequest ReadBillableDigestRequest `json:"readbillabledigestrequest,omitempty"`
}

// POST_ReadBillableDigestResponses holds responses of POST_ReadBillableDigest
type POST_ReadBillableDigestResponses struct {
	OK *ReadBillableDigestResponse
}

// POST_ReadCatalogParameters holds parameters to POST_ReadCatalog
type POST_ReadCatalogParameters struct {
	Readcatalogrequest ReadCatalogRequest `json:"readcatalogrequest,omitempty"`
}

// POST_ReadCatalogResponses holds responses of POST_ReadCatalog
type POST_ReadCatalogResponses struct {
	OK *ReadCatalogResponse
}

// POST_ReadClientGatewaysParameters holds parameters to POST_ReadClientGateways
type POST_ReadClientGatewaysParameters struct {
	Readclientgatewaysrequest ReadClientGatewaysRequest `json:"readclientgatewaysrequest,omitempty"`
}

// POST_ReadClientGatewaysResponses holds responses of POST_ReadClientGateways
type POST_ReadClientGatewaysResponses struct {
	OK *ReadClientGatewaysResponse
}

// POST_ReadConsoleOutputParameters holds parameters to POST_ReadConsoleOutput
type POST_ReadConsoleOutputParameters struct {
	Readconsoleoutputrequest ReadConsoleOutputRequest `json:"readconsoleoutputrequest,omitempty"`
}

// POST_ReadConsoleOutputResponses holds responses of POST_ReadConsoleOutput
type POST_ReadConsoleOutputResponses struct {
	OK      *ReadConsoleOutputResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_ReadDhcpOptionsParameters holds parameters to POST_ReadDhcpOptions
type POST_ReadDhcpOptionsParameters struct {
	Readdhcpoptionsrequest ReadDhcpOptionsRequest `json:"readdhcpoptionsrequest,omitempty"`
}

// POST_ReadDhcpOptionsResponses holds responses of POST_ReadDhcpOptions
type POST_ReadDhcpOptionsResponses struct {
	OK *ReadDhcpOptionsResponse
}

// POST_ReadDirectLinkInterfacesParameters holds parameters to POST_ReadDirectLinkInterfaces
type POST_ReadDirectLinkInterfacesParameters struct {
	Readdirectlinkinterfacesrequest ReadDirectLinkInterfacesRequest `json:"readdirectlinkinterfacesrequest,omitempty"`
}

// POST_ReadDirectLinkInterfacesResponses holds responses of POST_ReadDirectLinkInterfaces
type POST_ReadDirectLinkInterfacesResponses struct {
	OK *ReadDirectLinkInterfacesResponse
}

// POST_ReadDirectLinksParameters holds parameters to POST_ReadDirectLinks
type POST_ReadDirectLinksParameters struct {
	Readdirectlinksrequest ReadDirectLinksRequest `json:"readdirectlinksrequest,omitempty"`
}

// POST_ReadDirectLinksResponses holds responses of POST_ReadDirectLinks
type POST_ReadDirectLinksResponses struct {
	OK *ReadDirectLinksResponse
}

// POST_ReadImageExportTasksParameters holds parameters to POST_ReadImageExportTasks
type POST_ReadImageExportTasksParameters struct {
	Readimageexporttasksrequest ReadImageExportTasksRequest `json:"readimageexporttasksrequest,omitempty"`
}

// POST_ReadImageExportTasksResponses holds responses of POST_ReadImageExportTasks
type POST_ReadImageExportTasksResponses struct {
	OK *ReadImageExportTasksResponse
}

// POST_ReadImagesParameters holds parameters to POST_ReadImages
type POST_ReadImagesParameters struct {
	Readimagesrequest ReadImagesRequest `json:"readimagesrequest,omitempty"`
}

// POST_ReadImagesResponses holds responses of POST_ReadImages
type POST_ReadImagesResponses struct {
	OK      *ReadImagesResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_ReadInternetServicesParameters holds parameters to POST_ReadInternetServices
type POST_ReadInternetServicesParameters struct {
	Readinternetservicesrequest ReadInternetServicesRequest `json:"readinternetservicesrequest,omitempty"`
}

// POST_ReadInternetServicesResponses holds responses of POST_ReadInternetServices
type POST_ReadInternetServicesResponses struct {
	OK      *ReadInternetServicesResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_ReadKeypairsParameters holds parameters to POST_ReadKeypairs
type POST_ReadKeypairsParameters struct {
	Readkeypairsrequest ReadKeypairsRequest `json:"readkeypairsrequest,omitempty"`
}

// POST_ReadKeypairsResponses holds responses of POST_ReadKeypairs
type POST_ReadKeypairsResponses struct {
	OK      *ReadKeypairsResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_ReadListenerRulesParameters holds parameters to POST_ReadListenerRules
type POST_ReadListenerRulesParameters struct {
	Readlistenerrulesrequest ReadListenerRulesRequest `json:"readlistenerrulesrequest,omitempty"`
}

// POST_ReadListenerRulesResponses holds responses of POST_ReadListenerRules
type POST_ReadListenerRulesResponses struct {
	OK *ReadListenerRulesResponse
}

// POST_ReadLoadBalancersParameters holds parameters to POST_ReadLoadBalancers
type POST_ReadLoadBalancersParameters struct {
	Readloadbalancersrequest ReadLoadBalancersRequest `json:"readloadbalancersrequest,omitempty"`
}

// POST_ReadLoadBalancersResponses holds responses of POST_ReadLoadBalancers
type POST_ReadLoadBalancersResponses struct {
	OK *ReadLoadBalancersResponse
}

// POST_ReadLocationsParameters holds parameters to POST_ReadLocations
type POST_ReadLocationsParameters struct {
	Readlocationsrequest ReadLocationsRequest `json:"readlocationsrequest,omitempty"`
}

// POST_ReadLocationsResponses holds responses of POST_ReadLocations
type POST_ReadLocationsResponses struct {
	OK *ReadLocationsResponse
}

// POST_ReadNatServicesParameters holds parameters to POST_ReadNatServices
type POST_ReadNatServicesParameters struct {
	Readnatservicesrequest ReadNatServicesRequest `json:"readnatservicesrequest,omitempty"`
}

// POST_ReadNatServicesResponses holds responses of POST_ReadNatServices
type POST_ReadNatServicesResponses struct {
	OK      *ReadNatServicesResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_ReadNetAccessPointServicesParameters holds parameters to POST_ReadNetAccessPointServices
type POST_ReadNetAccessPointServicesParameters struct {
	Readnetaccesspointservicesrequest ReadNetAccessPointServicesRequest `json:"readnetaccesspointservicesrequest,omitempty"`
}

// POST_ReadNetAccessPointServicesResponses holds responses of POST_ReadNetAccessPointServices
type POST_ReadNetAccessPointServicesResponses struct {
	OK *ReadNetAccessPointServicesResponse
}

// POST_ReadNetAccessPointsParameters holds parameters to POST_ReadNetAccessPoints
type POST_ReadNetAccessPointsParameters struct {
	Readnetaccesspointsrequest ReadNetAccessPointsRequest `json:"readnetaccesspointsrequest,omitempty"`
}

// POST_ReadNetAccessPointsResponses holds responses of POST_ReadNetAccessPoints
type POST_ReadNetAccessPointsResponses struct {
	OK *ReadNetAccessPointsResponse
}

// POST_ReadNetPeeringsParameters holds parameters to POST_ReadNetPeerings
type POST_ReadNetPeeringsParameters struct {
	Readnetpeeringsrequest ReadNetPeeringsRequest `json:"readnetpeeringsrequest,omitempty"`
}

// POST_ReadNetPeeringsResponses holds responses of POST_ReadNetPeerings
type POST_ReadNetPeeringsResponses struct {
	OK      *ReadNetPeeringsResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_ReadNetsParameters holds parameters to POST_ReadNets
type POST_ReadNetsParameters struct {
	Readnetsrequest ReadNetsRequest `json:"readnetsrequest,omitempty"`
}

// POST_ReadNetsResponses holds responses of POST_ReadNets
type POST_ReadNetsResponses struct {
	OK      *ReadNetsResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_ReadNicsParameters holds parameters to POST_ReadNics
type POST_ReadNicsParameters struct {
	Readnicsrequest ReadNicsRequest `json:"readnicsrequest,omitempty"`
}

// POST_ReadNicsResponses holds responses of POST_ReadNics
type POST_ReadNicsResponses struct {
	OK      *ReadNicsResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_ReadPoliciesParameters holds parameters to POST_ReadPolicies
type POST_ReadPoliciesParameters struct {
	Readpoliciesrequest ReadPoliciesRequest `json:"readpoliciesrequest,omitempty"`
}

// POST_ReadPoliciesResponses holds responses of POST_ReadPolicies
type POST_ReadPoliciesResponses struct {
	OK *ReadPoliciesResponse
}

// POST_ReadPrefixListsParameters holds parameters to POST_ReadPrefixLists
type POST_ReadPrefixListsParameters struct {
	Readprefixlistsrequest ReadPrefixListsRequest `json:"readprefixlistsrequest,omitempty"`
}

// POST_ReadPrefixListsResponses holds responses of POST_ReadPrefixLists
type POST_ReadPrefixListsResponses struct {
	OK *ReadPrefixListsResponse
}

// POST_ReadProductTypesParameters holds parameters to POST_ReadProductTypes
type POST_ReadProductTypesParameters struct {
	Readproducttypesrequest ReadProductTypesRequest `json:"readproducttypesrequest,omitempty"`
}

// POST_ReadProductTypesResponses holds responses of POST_ReadProductTypes
type POST_ReadProductTypesResponses struct {
	OK *ReadProductTypesResponse
}

// POST_ReadPublicCatalogParameters holds parameters to POST_ReadPublicCatalog
type POST_ReadPublicCatalogParameters struct {
	Readpubliccatalogrequest ReadPublicCatalogRequest `json:"readpubliccatalogrequest,omitempty"`
}

// POST_ReadPublicCatalogResponses holds responses of POST_ReadPublicCatalog
type POST_ReadPublicCatalogResponses struct {
	OK *ReadPublicCatalogResponse
}

// POST_ReadPublicIpRangesParameters holds parameters to POST_ReadPublicIpRanges
type POST_ReadPublicIpRangesParameters struct {
	Readpubliciprangesrequest ReadPublicIpRangesRequest `json:"readpubliciprangesrequest,omitempty"`
}

// POST_ReadPublicIpRangesResponses holds responses of POST_ReadPublicIpRanges
type POST_ReadPublicIpRangesResponses struct {
	OK *ReadPublicIpRangesResponse
}

// POST_ReadPublicIpsParameters holds parameters to POST_ReadPublicIps
type POST_ReadPublicIpsParameters struct {
	Readpublicipsrequest ReadPublicIpsRequest `json:"readpublicipsrequest,omitempty"`
}

// POST_ReadPublicIpsResponses holds responses of POST_ReadPublicIps
type POST_ReadPublicIpsResponses struct {
	OK      *ReadPublicIpsResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_ReadQuotasParameters holds parameters to POST_ReadQuotas
type POST_ReadQuotasParameters struct {
	Readquotasrequest ReadQuotasRequest `json:"readquotasrequest,omitempty"`
}

// POST_ReadQuotasResponses holds responses of POST_ReadQuotas
type POST_ReadQuotasResponses struct {
	OK *ReadQuotasResponse
}

// POST_ReadRegionConfigParameters holds parameters to POST_ReadRegionConfig
type POST_ReadRegionConfigParameters struct {
	Readregionconfigrequest ReadRegionConfigRequest `json:"readregionconfigrequest,omitempty"`
}

// POST_ReadRegionConfigResponses holds responses of POST_ReadRegionConfig
type POST_ReadRegionConfigResponses struct {
	OK *ReadRegionConfigResponse
}

// POST_ReadRegionsParameters holds parameters to POST_ReadRegions
type POST_ReadRegionsParameters struct {
	Readregionsrequest ReadRegionsRequest `json:"readregionsrequest,omitempty"`
}

// POST_ReadRegionsResponses holds responses of POST_ReadRegions
type POST_ReadRegionsResponses struct {
	OK *ReadRegionsResponse
}

// POST_ReadReservedVmOffersParameters holds parameters to POST_ReadReservedVmOffers
type POST_ReadReservedVmOffersParameters struct {
	Readreservedvmoffersrequest ReadReservedVmOffersRequest `json:"readreservedvmoffersrequest,omitempty"`
}

// POST_ReadReservedVmOffersResponses holds responses of POST_ReadReservedVmOffers
type POST_ReadReservedVmOffersResponses struct {
	OK *ReadReservedVmOffersResponse
}

// POST_ReadReservedVmsParameters holds parameters to POST_ReadReservedVms
type POST_ReadReservedVmsParameters struct {
	Readreservedvmsrequest ReadReservedVmsRequest `json:"readreservedvmsrequest,omitempty"`
}

// POST_ReadReservedVmsResponses holds responses of POST_ReadReservedVms
type POST_ReadReservedVmsResponses struct {
	OK *ReadReservedVmsResponse
}

// POST_ReadRouteTablesParameters holds parameters to POST_ReadRouteTables
type POST_ReadRouteTablesParameters struct {
	Readroutetablesrequest ReadRouteTablesRequest `json:"readroutetablesrequest,omitempty"`
}

// POST_ReadRouteTablesResponses holds responses of POST_ReadRouteTables
type POST_ReadRouteTablesResponses struct {
	OK      *ReadRouteTablesResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_ReadSecurityGroupsParameters holds parameters to POST_ReadSecurityGroups
type POST_ReadSecurityGroupsParameters struct {
	Readsecuritygroupsrequest ReadSecurityGroupsRequest `json:"readsecuritygroupsrequest,omitempty"`
}

// POST_ReadSecurityGroupsResponses holds responses of POST_ReadSecurityGroups
type POST_ReadSecurityGroupsResponses struct {
	OK      *ReadSecurityGroupsResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_ReadServerCertificatesParameters holds parameters to POST_ReadServerCertificates
type POST_ReadServerCertificatesParameters struct {
	Readservercertificatesrequest ReadServerCertificatesRequest `json:"readservercertificatesrequest,omitempty"`
}

// POST_ReadServerCertificatesResponses holds responses of POST_ReadServerCertificates
type POST_ReadServerCertificatesResponses struct {
	OK *ReadServerCertificatesResponse
}

// POST_ReadSnapshotExportTasksParameters holds parameters to POST_ReadSnapshotExportTasks
type POST_ReadSnapshotExportTasksParameters struct {
	Readsnapshotexporttasksrequest ReadSnapshotExportTasksRequest `json:"readsnapshotexporttasksrequest,omitempty"`
}

// POST_ReadSnapshotExportTasksResponses holds responses of POST_ReadSnapshotExportTasks
type POST_ReadSnapshotExportTasksResponses struct {
	OK *ReadSnapshotExportTasksResponse
}

// POST_ReadSnapshotsParameters holds parameters to POST_ReadSnapshots
type POST_ReadSnapshotsParameters struct {
	Readsnapshotsrequest ReadSnapshotsRequest `json:"readsnapshotsrequest,omitempty"`
}

// POST_ReadSnapshotsResponses holds responses of POST_ReadSnapshots
type POST_ReadSnapshotsResponses struct {
	OK      *ReadSnapshotsResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_ReadSubnetsParameters holds parameters to POST_ReadSubnets
type POST_ReadSubnetsParameters struct {
	Readsubnetsrequest ReadSubnetsRequest `json:"readsubnetsrequest,omitempty"`
}

// POST_ReadSubnetsResponses holds responses of POST_ReadSubnets
type POST_ReadSubnetsResponses struct {
	OK      *ReadSubnetsResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_ReadSubregionsParameters holds parameters to POST_ReadSubregions
type POST_ReadSubregionsParameters struct {
	Readsubregionsrequest ReadSubregionsRequest `json:"readsubregionsrequest,omitempty"`
}

// POST_ReadSubregionsResponses holds responses of POST_ReadSubregions
type POST_ReadSubregionsResponses struct {
	OK *ReadSubregionsResponse
}

// POST_ReadTagsParameters holds parameters to POST_ReadTags
type POST_ReadTagsParameters struct {
	Readtagsrequest ReadTagsRequest `json:"readtagsrequest,omitempty"`
}

// POST_ReadTagsResponses holds responses of POST_ReadTags
type POST_ReadTagsResponses struct {
	OK      *ReadTagsResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_ReadUserGroupsParameters holds parameters to POST_ReadUserGroups
type POST_ReadUserGroupsParameters struct {
	Readusergroupsrequest ReadUserGroupsRequest `json:"readusergroupsrequest,omitempty"`
}

// POST_ReadUserGroupsResponses holds responses of POST_ReadUserGroups
type POST_ReadUserGroupsResponses struct {
	OK *ReadUserGroupsResponse
}

// POST_ReadUsersParameters holds parameters to POST_ReadUsers
type POST_ReadUsersParameters struct {
	Readusersrequest ReadUsersRequest `json:"readusersrequest,omitempty"`
}

// POST_ReadUsersResponses holds responses of POST_ReadUsers
type POST_ReadUsersResponses struct {
	OK *ReadUsersResponse
}

// POST_ReadVirtualGatewaysParameters holds parameters to POST_ReadVirtualGateways
type POST_ReadVirtualGatewaysParameters struct {
	Readvirtualgatewaysrequest ReadVirtualGatewaysRequest `json:"readvirtualgatewaysrequest,omitempty"`
}

// POST_ReadVirtualGatewaysResponses holds responses of POST_ReadVirtualGateways
type POST_ReadVirtualGatewaysResponses struct {
	OK *ReadVirtualGatewaysResponse
}

// POST_ReadVmTypesParameters holds parameters to POST_ReadVmTypes
type POST_ReadVmTypesParameters struct {
	Readvmtypesrequest ReadVmTypesRequest `json:"readvmtypesrequest,omitempty"`
}

// POST_ReadVmTypesResponses holds responses of POST_ReadVmTypes
type POST_ReadVmTypesResponses struct {
	OK *ReadVmTypesResponse
}

// POST_ReadVmsParameters holds parameters to POST_ReadVms
type POST_ReadVmsParameters struct {
	Readvmsrequest ReadVmsRequest `json:"readvmsrequest,omitempty"`
}

// POST_ReadVmsResponses holds responses of POST_ReadVms
type POST_ReadVmsResponses struct {
	OK      *ReadVmsResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_ReadVmsHealthParameters holds parameters to POST_ReadVmsHealth
type POST_ReadVmsHealthParameters struct {
	Readvmshealthrequest ReadVmsHealthRequest `json:"readvmshealthrequest,omitempty"`
}

// POST_ReadVmsHealthResponses holds responses of POST_ReadVmsHealth
type POST_ReadVmsHealthResponses struct {
	OK *ReadVmsHealthResponse
}

// POST_ReadVmsStateParameters holds parameters to POST_ReadVmsState
type POST_ReadVmsStateParameters struct {
	Readvmsstaterequest ReadVmsStateRequest `json:"readvmsstaterequest,omitempty"`
}

// POST_ReadVmsStateResponses holds responses of POST_ReadVmsState
type POST_ReadVmsStateResponses struct {
	OK      *ReadVmsStateResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_ReadVolumesParameters holds parameters to POST_ReadVolumes
type POST_ReadVolumesParameters struct {
	Readvolumesrequest ReadVolumesRequest `json:"readvolumesrequest,omitempty"`
}

// POST_ReadVolumesResponses holds responses of POST_ReadVolumes
type POST_ReadVolumesResponses struct {
	OK      *ReadVolumesResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_ReadVpnConnectionsParameters holds parameters to POST_ReadVpnConnections
type POST_ReadVpnConnectionsParameters struct {
	Readvpnconnectionsrequest ReadVpnConnectionsRequest `json:"readvpnconnectionsrequest,omitempty"`
}

// POST_ReadVpnConnectionsResponses holds responses of POST_ReadVpnConnections
type POST_ReadVpnConnectionsResponses struct {
	OK *ReadVpnConnectionsResponse
}

// POST_RebootVmsParameters holds parameters to POST_RebootVms
type POST_RebootVmsParameters struct {
	Rebootvmsrequest RebootVmsRequest `json:"rebootvmsrequest,omitempty"`
}

// POST_RebootVmsResponses holds responses of POST_RebootVms
type POST_RebootVmsResponses struct {
	OK      *RebootVmsResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_RegisterUserInUserGroupParameters holds parameters to POST_RegisterUserInUserGroup
type POST_RegisterUserInUserGroupParameters struct {
	Registeruserinusergrouprequest RegisterUserInUserGroupRequest `json:"registeruserinusergrouprequest,omitempty"`
}

// POST_RegisterUserInUserGroupResponses holds responses of POST_RegisterUserInUserGroup
type POST_RegisterUserInUserGroupResponses struct {
	OK *RegisterUserInUserGroupResponse
}

// POST_RegisterVmsInLoadBalancerParameters holds parameters to POST_RegisterVmsInLoadBalancer
type POST_RegisterVmsInLoadBalancerParameters struct {
	Registervmsinloadbalancerrequest RegisterVmsInLoadBalancerRequest `json:"registervmsinloadbalancerrequest,omitempty"`
}

// POST_RegisterVmsInLoadBalancerResponses holds responses of POST_RegisterVmsInLoadBalancer
type POST_RegisterVmsInLoadBalancerResponses struct {
	OK *RegisterVmsInLoadBalancerResponse
}

// POST_RejectNetPeeringParameters holds parameters to POST_RejectNetPeering
type POST_RejectNetPeeringParameters struct {
	Rejectnetpeeringrequest RejectNetPeeringRequest `json:"rejectnetpeeringrequest,omitempty"`
}

// POST_RejectNetPeeringResponses holds responses of POST_RejectNetPeering
type POST_RejectNetPeeringResponses struct {
	OK      *RejectNetPeeringResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code409 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_ResetAccountPasswordParameters holds parameters to POST_ResetAccountPassword
type POST_ResetAccountPasswordParameters struct {
	Resetaccountpasswordrequest ResetAccountPasswordRequest `json:"resetaccountpasswordrequest,omitempty"`
}

// POST_ResetAccountPasswordResponses holds responses of POST_ResetAccountPassword
type POST_ResetAccountPasswordResponses struct {
	OK *ResetAccountPasswordResponse
}

// POST_SendResetPasswordEmailParameters holds parameters to POST_SendResetPasswordEmail
type POST_SendResetPasswordEmailParameters struct {
	Sendresetpasswordemailrequest SendResetPasswordEmailRequest `json:"sendresetpasswordemailrequest,omitempty"`
}

// POST_SendResetPasswordEmailResponses holds responses of POST_SendResetPasswordEmail
type POST_SendResetPasswordEmailResponses struct {
	OK *SendResetPasswordEmailResponse
}

// POST_StartVmsParameters holds parameters to POST_StartVms
type POST_StartVmsParameters struct {
	Startvmsrequest StartVmsRequest `json:"startvmsrequest,omitempty"`
}

// POST_StartVmsResponses holds responses of POST_StartVms
type POST_StartVmsResponses struct {
	OK      *StartVmsResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_StopVmsParameters holds parameters to POST_StopVms
type POST_StopVmsParameters struct {
	Stopvmsrequest StopVmsRequest `json:"stopvmsrequest,omitempty"`
}

// POST_StopVmsResponses holds responses of POST_StopVms
type POST_StopVmsResponses struct {
	OK      *StopVmsResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_UnlinkInternetServiceParameters holds parameters to POST_UnlinkInternetService
type POST_UnlinkInternetServiceParameters struct {
	Unlinkinternetservicerequest UnlinkInternetServiceRequest `json:"unlinkinternetservicerequest,omitempty"`
}

// POST_UnlinkInternetServiceResponses holds responses of POST_UnlinkInternetService
type POST_UnlinkInternetServiceResponses struct {
	OK      *UnlinkInternetServiceResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_UnlinkNicParameters holds parameters to POST_UnlinkNic
type POST_UnlinkNicParameters struct {
	Unlinknicrequest UnlinkNicRequest `json:"unlinknicrequest,omitempty"`
}

// POST_UnlinkNicResponses holds responses of POST_UnlinkNic
type POST_UnlinkNicResponses struct {
	OK      *UnlinkNicResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_UnlinkPolicyParameters holds parameters to POST_UnlinkPolicy
type POST_UnlinkPolicyParameters struct {
	Unlinkpolicyrequest UnlinkPolicyRequest `json:"unlinkpolicyrequest,omitempty"`
}

// POST_UnlinkPolicyResponses holds responses of POST_UnlinkPolicy
type POST_UnlinkPolicyResponses struct {
	OK *UnlinkPolicyResponse
}

// POST_UnlinkPrivateIpsParameters holds parameters to POST_UnlinkPrivateIps
type POST_UnlinkPrivateIpsParameters struct {
	Unlinkprivateipsrequest UnlinkPrivateIpsRequest `json:"unlinkprivateipsrequest,omitempty"`
}

// POST_UnlinkPrivateIpsResponses holds responses of POST_UnlinkPrivateIps
type POST_UnlinkPrivateIpsResponses struct {
	OK      *UnlinkPrivateIpsResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_UnlinkPublicIpParameters holds parameters to POST_UnlinkPublicIp
type POST_UnlinkPublicIpParameters struct {
	Unlinkpubliciprequest UnlinkPublicIpRequest `json:"unlinkpubliciprequest,omitempty"`
}

// POST_UnlinkPublicIpResponses holds responses of POST_UnlinkPublicIp
type POST_UnlinkPublicIpResponses struct {
	OK      *UnlinkPublicIpResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_UnlinkRouteTableParameters holds parameters to POST_UnlinkRouteTable
type POST_UnlinkRouteTableParameters struct {
	Unlinkroutetablerequest UnlinkRouteTableRequest `json:"unlinkroutetablerequest,omitempty"`
}

// POST_UnlinkRouteTableResponses holds responses of POST_UnlinkRouteTable
type POST_UnlinkRouteTableResponses struct {
	OK      *UnlinkRouteTableResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_UnlinkVirtualGatewayParameters holds parameters to POST_UnlinkVirtualGateway
type POST_UnlinkVirtualGatewayParameters struct {
	Unlinkvirtualgatewayrequest UnlinkVirtualGatewayRequest `json:"unlinkvirtualgatewayrequest,omitempty"`
}

// POST_UnlinkVirtualGatewayResponses holds responses of POST_UnlinkVirtualGateway
type POST_UnlinkVirtualGatewayResponses struct {
	OK *UnlinkVirtualGatewayResponse
}

// POST_UnlinkVolumeParameters holds parameters to POST_UnlinkVolume
type POST_UnlinkVolumeParameters struct {
	Unlinkvolumerequest UnlinkVolumeRequest `json:"unlinkvolumerequest,omitempty"`
}

// POST_UnlinkVolumeResponses holds responses of POST_UnlinkVolume
type POST_UnlinkVolumeResponses struct {
	OK      *UnlinkVolumeResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_UpdateAccountParameters holds parameters to POST_UpdateAccount
type POST_UpdateAccountParameters struct {
	Updateaccountrequest UpdateAccountRequest `json:"updateaccountrequest,omitempty"`
}

// POST_UpdateAccountResponses holds responses of POST_UpdateAccount
type POST_UpdateAccountResponses struct {
	OK *UpdateAccountResponse
}

// POST_UpdateApiKeyParameters holds parameters to POST_UpdateApiKey
type POST_UpdateApiKeyParameters struct {
	Updateapikeyrequest UpdateApiKeyRequest `json:"updateapikeyrequest,omitempty"`
}

// POST_UpdateApiKeyResponses holds responses of POST_UpdateApiKey
type POST_UpdateApiKeyResponses struct {
	OK *UpdateApiKeyResponse
}

// POST_UpdateHealthCheckParameters holds parameters to POST_UpdateHealthCheck
type POST_UpdateHealthCheckParameters struct {
	Updatehealthcheckrequest UpdateHealthCheckRequest `json:"updatehealthcheckrequest,omitempty"`
}

// POST_UpdateHealthCheckResponses holds responses of POST_UpdateHealthCheck
type POST_UpdateHealthCheckResponses struct {
	OK *UpdateHealthCheckResponse
}

// POST_UpdateImageParameters holds parameters to POST_UpdateImage
type POST_UpdateImageParameters struct {
	Updateimagerequest UpdateImageRequest `json:"updateimagerequest,omitempty"`
}

// POST_UpdateImageResponses holds responses of POST_UpdateImage
type POST_UpdateImageResponses struct {
	OK      *UpdateImageResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_UpdateKeypairParameters holds parameters to POST_UpdateKeypair
type POST_UpdateKeypairParameters struct {
	Updatekeypairrequest UpdateKeypairRequest `json:"updatekeypairrequest,omitempty"`
}

// POST_UpdateKeypairResponses holds responses of POST_UpdateKeypair
type POST_UpdateKeypairResponses struct {
	OK *UpdateKeypairResponse
}

// POST_UpdateListenerRuleParameters holds parameters to POST_UpdateListenerRule
type POST_UpdateListenerRuleParameters struct {
	Updatelistenerrulerequest UpdateListenerRuleRequest `json:"updatelistenerrulerequest,omitempty"`
}

// POST_UpdateListenerRuleResponses holds responses of POST_UpdateListenerRule
type POST_UpdateListenerRuleResponses struct {
	OK *UpdateListenerRuleResponse
}

// POST_UpdateLoadBalancerParameters holds parameters to POST_UpdateLoadBalancer
type POST_UpdateLoadBalancerParameters struct {
	Updateloadbalancerrequest UpdateLoadBalancerRequest `json:"updateloadbalancerrequest,omitempty"`
}

// POST_UpdateLoadBalancerResponses holds responses of POST_UpdateLoadBalancer
type POST_UpdateLoadBalancerResponses struct {
	OK *UpdateLoadBalancerResponse
}

// POST_UpdateNetParameters holds parameters to POST_UpdateNet
type POST_UpdateNetParameters struct {
	Updatenetrequest UpdateNetRequest `json:"updatenetrequest,omitempty"`
}

// POST_UpdateNetResponses holds responses of POST_UpdateNet
type POST_UpdateNetResponses struct {
	OK      *UpdateNetResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_UpdateNetAccessPointParameters holds parameters to POST_UpdateNetAccessPoint
type POST_UpdateNetAccessPointParameters struct {
	Updatenetaccesspointrequest UpdateNetAccessPointRequest `json:"updatenetaccesspointrequest,omitempty"`
}

// POST_UpdateNetAccessPointResponses holds responses of POST_UpdateNetAccessPoint
type POST_UpdateNetAccessPointResponses struct {
	OK *UpdateNetAccessPointResponse
}

// POST_UpdateNicParameters holds parameters to POST_UpdateNic
type POST_UpdateNicParameters struct {
	Updatenicrequest UpdateNicRequest `json:"updatenicrequest,omitempty"`
}

// POST_UpdateNicResponses holds responses of POST_UpdateNic
type POST_UpdateNicResponses struct {
	OK      *UpdateNicResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_UpdateRouteParameters holds parameters to POST_UpdateRoute
type POST_UpdateRouteParameters struct {
	Updaterouterequest UpdateRouteRequest `json:"updaterouterequest,omitempty"`
}

// POST_UpdateRouteResponses holds responses of POST_UpdateRoute
type POST_UpdateRouteResponses struct {
	OK      *UpdateRouteResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_UpdateRoutePropagationParameters holds parameters to POST_UpdateRoutePropagation
type POST_UpdateRoutePropagationParameters struct {
	Updateroutepropagationrequest UpdateRoutePropagationRequest `json:"updateroutepropagationrequest,omitempty"`
}

// POST_UpdateRoutePropagationResponses holds responses of POST_UpdateRoutePropagation
type POST_UpdateRoutePropagationResponses struct {
	OK *UpdateRoutePropagationResponse
}

// POST_UpdateServerCertificateParameters holds parameters to POST_UpdateServerCertificate
type POST_UpdateServerCertificateParameters struct {
	Updateservercertificaterequest UpdateServerCertificateRequest `json:"updateservercertificaterequest,omitempty"`
}

// POST_UpdateServerCertificateResponses holds responses of POST_UpdateServerCertificate
type POST_UpdateServerCertificateResponses struct {
	OK *UpdateServerCertificateResponse
}

// POST_UpdateSnapshotParameters holds parameters to POST_UpdateSnapshot
type POST_UpdateSnapshotParameters struct {
	Updatesnapshotrequest UpdateSnapshotRequest `json:"updatesnapshotrequest,omitempty"`
}

// POST_UpdateSnapshotResponses holds responses of POST_UpdateSnapshot
type POST_UpdateSnapshotResponses struct {
	OK      *UpdateSnapshotResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}

// POST_UpdateUserParameters holds parameters to POST_UpdateUser
type POST_UpdateUserParameters struct {
	Updateuserrequest UpdateUserRequest `json:"updateuserrequest,omitempty"`
}

// POST_UpdateUserResponses holds responses of POST_UpdateUser
type POST_UpdateUserResponses struct {
	OK *UpdateUserResponse
}

// POST_UpdateUserGroupParameters holds parameters to POST_UpdateUserGroup
type POST_UpdateUserGroupParameters struct {
	Updateusergrouprequest UpdateUserGroupRequest `json:"updateusergrouprequest,omitempty"`
}

// POST_UpdateUserGroupResponses holds responses of POST_UpdateUserGroup
type POST_UpdateUserGroupResponses struct {
	OK *UpdateUserGroupResponse
}

// POST_UpdateVmParameters holds parameters to POST_UpdateVm
type POST_UpdateVmParameters struct {
	Updatevmrequest UpdateVmRequest `json:"updatevmrequest,omitempty"`
}

// POST_UpdateVmResponses holds responses of POST_UpdateVm
type POST_UpdateVmResponses struct {
	OK      *UpdateVmResponse
	Code400 *ErrorResponse
	Code401 *ErrorResponse
	Code500 *ErrorResponse
}
