package sdk

import (
	common "github.com/NaverCloudPlatform/ncloud-sdk-go/common"
)

type Conn struct {
	accessKey string
	secretKey string
	apiURL    string
}

// ServerImage structures
type ServerImage struct {
	MemberServerImageNo                    string            `xml:"memberServerImageNo"`
	MemberServerImageName                  string            `xml:"memberServerImageName"`
	MemberServerImageDescription           string            `xml:"memberServerImageDescription"`
	OriginalServerInstanceNo               string            `xml:"originalServerInstanceNo"`
	OriginalServerProductCode              string            `xml:"originalServerProductCode"`
	OriginalServerName                     string            `xml:"originalServerName"`
	OriginalBaseBlockStorageDiskType       common.CommonCode `xml:"originalBaseBlockStorageDiskType"`
	OriginalServerImageProductCode         string            `xml:"originalServerImageProductCode"`
	OriginalOsInformation                  string            `xml:"originalOsInformation"`
	OriginalServerImageName                string            `xml:"originalServerImageName"`
	MemberServerImageStatusName            string            `xml:"memberServerImageStatusName"`
	MemberServerImageStatus                common.CommonCode `xml:"memberServerImageStatus"`
	MemberServerImageOperation             common.CommonCode `xml:"memberServerImageOperation"`
	MemberServerImagePlatformType          common.CommonCode `xml:"memberServerImagePlatformType"`
	CreateDate                             string            `xml:"createDate"`
	Zone                                   common.Zone       `xml:"zone"`
	Region                                 common.Region     `xml:"region"`
	MemberServerImageBlockStorageTotalRows int               `xml:"memberServerImageBlockStorageTotalRows"`
	MemberServerImageBlockStorageTotalSize int               `xml:"memberServerImageBlockStorageTotalSize"`
}

type MemberServerImageList struct {
	common.CommonResponse
	TotalRows             int           `xml:"totalRows"`
	MemberServerImageList []ServerImage `xml:"memberServerImageList>memberServerImage,omitempty"`
}

type RequestServerImageList struct {
	MemberServerImageNoList []string
	PlatformTypeCodeList    []string
	PageNo                  int
	PageSize                int
	RegionNo                string
	SortedBy                string
	SortingOrder            string
}

type RequestCreateServerImage struct {
	MemberServerImageName        string
	MemberServerImageDescription string
	ServerInstanceNo             string
}

type RequestGetServerImageProductList struct {
	ExclusionProductCode string
	ProductCode          string
	PlatformTypeCodeList []string
	BlockStorageSize     int
	RegionNo             string
}

// ProductList : Response of server product list
type ProductList struct {
	common.CommonResponse
	TotalRows int       `xml:"totalRows"`
	Product   []Product `xml:"productList>product,omitempty"`
}

// Product : Product information of Server
type Product struct {
	ProductCode          string            `xml:"productCode"`
	ProductName          string            `xml:"productName"`
	ProductType          common.CommonCode `xml:"productType"`
	ProductDescription   string            `xml:"productDescription"`
	InfraResourceType    common.CommonCode `xml:"infraResourceType"`
	CPUCount             int               `xml:"cpuCount"`
	MemorySize           int               `xml:"memorySize"`
	BaseBlockStorageSize int               `xml:"baseBlockStorageSize"`
	PlatformType         common.CommonCode `xml:"platformType"`
	OsInformation        string            `xml:"osInformation"`
	AddBlockStroageSize  int               `xml:"addBlockStroageSize"`
}

// RequestCreateServerInstance is Server Instances structures
type RequestCreateServerInstance struct {
	ServerImageProductCode                string
	ServerProductCode                     string
	MemberServerImageNo                   string
	ServerName                            string
	ServerDescription                     string
	LoginKeyName                          string
	IsProtectServerTermination            bool
	ServerCreateCount                     int
	ServerCreateStartNo                   int
	InternetLineTypeCode                  string
	FeeSystemTypeCode                     string
	UserData                              string
	ZoneNo                                string
	AccessControlGroupConfigurationNoList []string
}

type ServerInstanceList struct {
	common.CommonResponse
	TotalRows          int              `xml:"totalRows"`
	ServerInstanceList []ServerInstance `xml:"serverInstanceList>serverInstance,omitempty"`
}

type ServerInstance struct {
	ServerInstanceNo               string               `xml:"serverInstanceNo"`
	ServerName                     string               `xml:"serverName"`
	ServerDescription              string               `xml:"serverDescription"`
	CPUCount                       int                  `xml:"cpuCount"`
	MemorySize                     int                  `xml:"memorySize"`
	BaseBlockStorageSize           int                  `xml:"baseBlockStorageSize"`
	PlatformType                   common.CommonCode    `xml:"platformType"`
	LoginKeyName                   string               `xml:"loginKeyName"`
	IsFeeChargingMonitoring        bool                 `xml:"isFeeChargingMonitoring"`
	PublicIP                       string               `xml:"publicIp"`
	PrivateIP                      string               `xml:"privateIp"`
	ServerImageName                string               `xml:"serverImageName"`
	ServerInstanceStatus           common.CommonCode    `xml:"serverInstanceStatus"`
	ServerInstanceOperation        common.CommonCode    `xml:"serverInstanceOperation"`
	ServerInstanceStatusName       string               `xml:"serverInstanceStatusName"`
	CreateDate                     string               `xml:"createDate"`
	Uptime                         string               `xml:"uptime"`
	ServerImageProductCode         string               `xml:"serverImageProductCode"`
	ServerProductCode              string               `xml:"serverProductCode"`
	IsProtectServerTermination     bool                 `xml:"isProtectServerTermination"`
	PortForwardingPublicIP         string               `xml:"portForwardingPublicIp"`
	PortForwardingExternalPort     int                  `xml:"portForwardingExternalPort"`
	PortForwardingInternalPort     int                  `xml:"portForwardingInternalPort"`
	Zone                           common.Zone          `xml:"zone"`
	Region                         common.Region        `xml:"region"`
	BaseBlockStorageDiskType       common.CommonCode    `xml:"baseBlockStorageDiskType"`
	BaseBlockStroageDiskDetailType common.CommonCode    `xml:"baseBlockStroageDiskDetailType"`
	InternetLineType               common.CommonCode    `xml:"internetLineType"`
	UserData                       string               `xml:"userData"`
	AccessControlGroupList         []AccessControlGroup `xml:"accessControlGroupList>accessControlGroup"`
}

type AccessControlGroup struct {
	AccessControlGroupConfigurationNo string `xml:"accessControlGroupConfigurationNo"`
	AccessControlGroupName            string `xml:"accessControlGroupName"`
	AccessControlGroupDescription     string `xml:"accessControlGroupDescription"`
	IsDefault                         bool   `xml:"isDefault"`
	CreateDate                        string `xml:"createDate"`
}

// RequestGetLoginKeyList is Login Key structures
type RequestGetLoginKeyList struct {
	KeyName  string
	PageNo   int
	PageSize int
}

type LoginKeyList struct {
	common.CommonResponse
	TotalRows    int        `xml:"totalRows"`
	LoginKeyList []LoginKey `xml:"loginKeyList>loginKey,omitempty"`
}

type LoginKey struct {
	Fingerprint string `xml:"fingerprint"`
	KeyName     string `xml:"keyName"`
	CreateDate  string `xml:"createDate"`
}

type PrivateKey struct {
	common.CommonResponse
	PrivateKey string `xml:"privateKey"`
}
type RequestCreatePublicIPInstance struct {
	ServerInstanceNo     string
	PublicIPDescription  string
	InternetLineTypeCode string
	RegionNo             string
}

type RequestPublicIPInstanceList struct {
	IsAssociated           bool
	PublicIPInstanceNoList []string
	PublicIPList           []string
	SearchFilterName       string
	SearchFilterValue      string
	InternetLineTypeCode   string
	RegionNo               string
	PageNo                 int
	PageSize               int
	SortedBy               string
	SortingOrder           string
}

type PublicIPInstanceList struct {
	common.CommonResponse
	TotalRows            int                `xml:"totalRows"`
	PublicIPInstanceList []PublicIPInstance `xml:"publicIpInstanceList>publicIpInstance,omitempty"`
}

type PublicIPInstance struct {
	PublicIPInstanceNo         string            `xml:"publicIpInstanceNo"`
	PublicIP                   string            `xml:"publicIp"`
	PublicIPDescription        string            `xml:"publicIpDescription"`
	CreateDate                 string            `xml:"createDate"`
	InternetLineType           common.CommonCode `xml:"internetLineType"`
	PublicIPInstanceStatusName string            `xml:"publicIpInstanceStatusName"`
	PublicIPInstanceStatus     common.CommonCode `xml:"publicIpInstanceStatus"`
	PublicIPInstanceOperation  common.CommonCode `xml:"publicIpInstanceOperation"`
	PublicIPKindType           common.CommonCode `xml:"publicIpKindType"`
	ServerInstance             ServerInstance    `xml:"serverInstanceAssociatedWithPublicIp"`
}

type RequestDeletePublicIPInstances struct {
	PublicIPInstanceNoList []string
}

// RequestGetServerInstanceList : Get Server Instance List
type RequestGetServerInstanceList struct {
	ServerInstanceNoList               []string
	SearchFilterName                   string
	SearchFilterValue                  string
	PageNo                             int
	PageSize                           int
	ServerInstanceStatusCode           string
	InternetLineTypeCode               string
	RegionNo                           string
	BaseBlockStorageDiskTypeCode       string
	BaseBlockStorageDiskDetailTypeCode string
	SortedBy                           string
	SortingOrder                       string
}

type RequestStopServerInstances struct {
	ServerInstanceNoList []string
}

type RequestTerminateServerInstances struct {
	ServerInstanceNoList []string
}

// RequestGetRootPassword : Request to get root password of the server
type RequestGetRootPassword struct {
	ServerInstanceNo string
	PrivateKey       string
}

// RootPassword : Response of getting root password of the server
type RootPassword struct {
	common.CommonResponse
	TotalRows    int    `xml:"totalRows"`
	RootPassword string `xml:"rootPassword"`
}

// RequestGetZoneList : Request to get zone list
type RequestGetZoneList struct {
	regionNo string
}

// ZoneList : Response of getting zone list
type ZoneList struct {
	common.CommonResponse
	TotalRows int           `xml:"totalRows"`
	Zone      []common.Zone `xml:"zoneList>zone"`
}

// RegionList : Response of getting region list
type RegionList struct {
	common.CommonResponse
	TotalRows  int             `xml:"totalRows"`
	RegionList []common.Region `xml:"regionList>region,omitempty"`
}

type RequestBlockStorageInstance struct {
	BlockStorageName        string
	BlockStorageSize        int
	BlockStorageDescription string
	ServerInstanceNo        string
}

type RequestBlockStorageInstanceList struct {
	ServerInstanceNo               string
	BlockStorageInstanceNoList     []string
	SearchFilterName               string
	SearchFilterValue              string
	BlockStorageTypeCodeList       []string
	PageNo                         int
	PageSize                       int
	BlockStorageInstanceStatusCode string
	DiskTypeCode                   string
	DiskDetailTypeCode             string
	RegionNo                       string
	SortedBy                       string
	SortingOrder                   string
}

type BlockStorageInstanceList struct {
	common.CommonResponse
	TotalRows            int                    `xml:"totalRows"`
	BlockStorageInstance []BlockStorageInstance `xml:"blockStorageInstanceList>blockStorageInstance,omitempty"`
}

type BlockStorageInstance struct {
	BlockStorageInstanceNo          string            `xml:"blockStorageInstanceNo"`
	ServerInstanceNo                string            `xml:"serverInstanceNo"`
	ServerName                      string            `xml:"serverName"`
	BlockStorageType                common.CommonCode `xml:"blockStorageType"`
	BlockStorageName                string            `xml:"blockStorageName"`
	BlockStorageSize                int               `xml:"blockStorageSize"`
	DeviceName                      string            `xml:"deviceName"`
	BlockStorageProductCode         string            `xml:"blockStorageProductCode"`
	BlockStorageInstanceStatus      common.CommonCode `xml:"blockStorageInstanceStatus"`
	BlockStorageInstanceOperation   common.CommonCode `xml:"blockStorageInstanceOperation"`
	BlockStorageInstanceStatusName  string            `xml:"blockStorageInstanceStatusName"`
	CreateDate                      string            `xml:"createDate"`
	BlockStorageInstanceDescription string            `xml:"blockStorageInstanceDescription"`
	DiskType                        common.CommonCode `xml:"diskType"`
	DiskDetailType                  common.CommonCode `xml:"diskDetailType"`
}

// RequestGetServerProductList : Request to get server product list
type RequestGetServerProductList struct {
	ExclusionProductCode   string
	ProductCode            string
	ServerImageProductCode string
	RegionNo               string
}

type RequestAccessControlGroupList struct {
	AccessControlGroupConfigurationNoList []string
	IsDefault                             bool
	AccessControlGroupName                string
	PageNo                                int
	PageSize                              int
}

type AccessControlGroupList struct {
	common.CommonResponse
	TotalRows          int                  `xml:"totalRows"`
	AccessControlGroup []AccessControlGroup `xml:"accessControlGroupList>accessControlGroup,omitempty"`
}

type AccessControlRuleList struct {
	common.CommonResponse
	TotalRows             int                 `xml:"totalRows"`
	AccessControlRuleList []AccessControlRule `xml:"accessControlRuleList>accessControlRule,omitempty"`
}

type AccessControlRule struct {
	AccessControlRuleConfigurationNo       string            `xml:"accessControlRuleConfigurationNo"`
	AccessControlRuleDescription           string            `xml:"accessControlRuleDescription"`
	SourceAccessControlRuleConfigurationNo string            `xml:"sourceAccessControlRuleConfigurationNo"`
	SourceAccessControlRuleName            string            `xml:"sourceAccessControlRuleName"`
	ProtocolType                           common.CommonCode `xml:"protocolType"`
	SourceIP                               string            `xml:"sourceIp"`
	DestinationPort                        string            `xml:"destinationPort"`
}
