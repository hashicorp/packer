package constants

// complete flags
const (
	AuthorizedKey string = "authorizedKey"
	Certificate   string = "certificate"
	Error         string = "error"
	SSHHost       string = "sshHost"
	Thumbprint    string = "thumbprint"
	Ui            string = "ui"
)

// Default replica count for image versions in shared image gallery
const (
	SharedImageGalleryImageVersionDefaultMinReplicaCount int32 = 1
	SharedImageGalleryImageVersionDefaultMaxReplicaCount int32 = 10
)

const (
	ArmCaptureTemplate                 string = "arm.CaptureTemplate"
	ArmComputeName                     string = "arm.ComputeName"
	ArmImageParameters                 string = "arm.ImageParameters"
	ArmCertificateUrl                  string = "arm.CertificateUrl"
	ArmKeyVaultDeploymentName          string = "arm.KeyVaultDeploymentName"
	ArmDeploymentName                  string = "arm.DeploymentName"
	ArmNicName                         string = "arm.NicName"
	ArmKeyVaultName                    string = "arm.KeyVaultName"
	ArmLocation                        string = "arm.Location"
	ArmOSDiskVhd                       string = "arm.OSDiskVhd"
	ArmAdditionalDiskVhds              string = "arm.AdditionalDiskVhds"
	ArmPublicIPAddressName             string = "arm.PublicIPAddressName"
	ArmResourceGroupName               string = "arm.ResourceGroupName"
	ArmIsResourceGroupCreated          string = "arm.IsResourceGroupCreated"
	ArmDoubleResourceGroupNameSet      string = "arm.DoubleResourceGroupNameSet"
	ArmStorageAccountName              string = "arm.StorageAccountName"
	ArmTags                            string = "arm.Tags"
	ArmVirtualMachineCaptureParameters string = "arm.VirtualMachineCaptureParameters"
	ArmIsExistingResourceGroup         string = "arm.IsExistingResourceGroup"
	ArmIsExistingKeyVault              string = "arm.IsExistingKeyVault"

	ArmIsManagedImage                                         string = "arm.IsManagedImage"
	ArmManagedImageResourceGroupName                          string = "arm.ManagedImageResourceGroupName"
	ArmManagedImageName                                       string = "arm.ManagedImageName"
	ArmManagedImageSigPublishResourceGroup                    string = "arm.ManagedImageSigPublishResourceGroup"
	ArmManagedImageSharedGalleryName                          string = "arm.ManagedImageSharedGalleryName"
	ArmManagedImageSharedGalleryImageName                     string = "arm.ManagedImageSharedGalleryImageName"
	ArmManagedImageSharedGalleryImageVersion                  string = "arm.ManagedImageSharedGalleryImageVersion"
	ArmManagedImageSharedGalleryReplicationRegions            string = "arm.ManagedImageSharedGalleryReplicationRegions"
	ArmManagedImageSharedGalleryId                            string = "arm.ArmManagedImageSharedGalleryId"
	ArmManagedImageSharedGalleryImageVersionEndOfLifeDate     string = "arm.ArmManagedImageSharedGalleryImageVersionEndOfLifeDate"
	ArmManagedImageSharedGalleryImageVersionReplicaCount      string = "arm.ArmManagedImageSharedGalleryImageVersionReplicaCount"
	ArmManagedImageSharedGalleryImageVersionExcludeFromLatest string = "arm.ArmManagedImageSharedGalleryImageVersionExcludeFromLatest"
	ArmManagedImageSharedGalleryImageVersionStorageType       string = "arm.ArmManagedImageSharedGalleryImageVersionStorageType"
	ArmManagedImageSubscription                               string = "arm.ArmManagedImageSubscription"
	ArmAsyncResourceGroupDelete                               string = "arm.AsyncResourceGroupDelete"
	ArmManagedImageOSDiskSnapshotName                         string = "arm.ManagedImageOSDiskSnapshotName"
	ArmManagedImageDataDiskSnapshotPrefix                     string = "arm.ManagedImageDataDiskSnapshotPrefix"
)
