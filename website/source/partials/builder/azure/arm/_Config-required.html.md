<!-- Code generated from the comments of the Config struct in builder/azure/arm/config.go; DO NOT EDIT MANUALLY -->

-   `image_publisher` (string) - Name of the publisher to use for your base image (Azure Marketplace Images only). See
    [documentation](https://azure.microsoft.com/en-us/documentation/articles/resource-groups-vm-searching/)
    for details.
    
    CLI example `az vm image list-publishers --location westus`
    
-   `image_offer` (string) - Name of the publisher's offer to use for your base image (Azure Marketplace Images only). See
    [documentation](https://azure.microsoft.com/en-us/documentation/articles/resource-groups-vm-searching/)
    for details.
    
    CLI example
    `az vm image list-offers --location westus --publisher Canonical`
    
-   `image_sku` (string) - SKU of the image offer to use for your base image (Azure Marketplace Images only). See
    [documentation](https://azure.microsoft.com/en-us/documentation/articles/resource-groups-vm-searching/)
    for details.
    
    CLI example
    `az vm image list-skus --location westus --publisher Canonical --offer UbuntuServer`
    
-   `image_url` (string) - URL to a custom VHD to use for your base image. If this value is set, do
    not set image_publisher, image_offer, image_sku, or image_version.
    
-   `custom_managed_image_name` (string) - Name of a custom managed image to use for your base image. If this value is set, do
    not set image_publisher, image_offer, image_sku, or image_version.
    If this value is set, the value
    `custom_managed_image_resource_group_name` must also be set. See
    [documentation](https://docs.microsoft.com/en-us/azure/storage/storage-managed-disks-overview#images)
    to learn more about managed images.
    
-   `custom_managed_image_resource_group_name` (string) - Name of a custom managed image's resource group to use for your base image. If this
    value is set, image_publisher, image_offer, image_sku, or image_version.
    `custom_managed_image_name` must also be set. See
    [documentation](https://docs.microsoft.com/en-us/azure/storage/storage-managed-disks-overview#images)
    to learn more about managed images.
    