<!-- Code generated from the comments of the Config struct in builder/azure/arm/config.go; DO NOT EDIT MANUALLY -->

-   `image_publisher` (string) - PublisherName for your base image. See
    [documentation](https://azure.microsoft.com/en-us/documentation/articles/resource-groups-vm-searching/)
    for details.
    
    CLI example `az vm image list-publishers --location westus`
    
-   `image_offer` (string) - Offer for your base image. See
    [documentation](https://azure.microsoft.com/en-us/documentation/articles/resource-groups-vm-searching/)
    for details.
    
    CLI example
    `az vm image list-offers --location westus --publisher Canonical`
    
-   `image_sku` (string) - SKU for your base image. See
    [documentation](https://azure.microsoft.com/en-us/documentation/articles/resource-groups-vm-searching/)
    for details.
    
    CLI example
    `az vm image list-skus --location westus --publisher Canonical --offer UbuntuServer`
    