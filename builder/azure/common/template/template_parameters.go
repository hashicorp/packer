package template

// The intent of these types to facilitate interchange with Azure in the
// appropriate JSON format. A sample format is below.  Each parameter listed
// below corresponds to a parameter defined in the template.
//
// {
//   "storageAccountName": {
//     "value" : "my_storage_account_name"
//   },
//   "adminUserName" : {
//     "value": "admin"
//   }
// }

type TemplateParameter struct {
	Value string `json:"value"`
}

type TemplateParameters struct {
	AdminUsername              *TemplateParameter `json:"adminUsername,omitempty"`
	AdminPassword              *TemplateParameter `json:"adminPassword,omitempty"`
	DnsNameForPublicIP         *TemplateParameter `json:"dnsNameForPublicIP,omitempty"`
	KeyVaultName               *TemplateParameter `json:"keyVaultName,omitempty"`
	KeyVaultSecretValue        *TemplateParameter `json:"keyVaultSecretValue,omitempty"`
	ObjectId                   *TemplateParameter `json:"objectId,omitempty"`
	NicName                    *TemplateParameter `json:"nicName,omitempty"`
	OSDiskName                 *TemplateParameter `json:"osDiskName,omitempty"`
	PublicIPAddressName        *TemplateParameter `json:"publicIPAddressName,omitempty"`
	StorageAccountBlobEndpoint *TemplateParameter `json:"storageAccountBlobEndpoint,omitempty"`
	SubnetName                 *TemplateParameter `json:"subnetName,omitempty"`
	TenantId                   *TemplateParameter `json:"tenantId,omitempty"`
	VirtualNetworkName         *TemplateParameter `json:"virtualNetworkName,omitempty"`
	NsgName                    *TemplateParameter `json:"nsgName,omitempty"`
	VMSize                     *TemplateParameter `json:"vmSize,omitempty"`
	VMName                     *TemplateParameter `json:"vmName,omitempty"`
}
