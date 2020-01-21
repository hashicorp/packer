#!/usr/bin/env bash
set -e

meta_name=
azure_client_id=       # Derived from application after creation
azure_client_name=     # Application name
azure_client_secret=   # Application password
azure_group_name=
azure_storage_name=
azure_subscription_id= # Derived from the account after login
azure_tenant_id=       # Derived from the account after login
location=
azure_object_id=
azureversion=
create_sleep=10

showhelp() {
    echo "azure-setup"
    echo ""
    echo "  azure-setup helps you generate packer credentials for azure"
    echo ""
    echo "  The script creates a resource group, storage account, application"
    echo "  (client), service principal, and permissions and displays a snippet"
    echo "  for use in your packer templates."
    echo ""
    echo "  For simplicity we make a lot of assumptions and choose reasonable"
    echo "  defaults. If you want more control over what happens, please use"
    echo "  the azure-cli directly."
    echo ""
    echo "  Note that you must already have an Azure account, username,"
    echo "  password, and subscription. You can create those here:"
    echo ""
    echo "  - https://azure.microsoft.com/en-us/account/"
    echo ""
    echo "REQUIREMENTS"
    echo ""
    echo "  - azure-cli"
    echo "  - jq"
    echo ""
    echo "  Use the requirements command (below) for more info."
    echo ""
    echo "USAGE"
    echo ""
    echo "  ./azure-setup.sh requirements"
    echo "  ./azure-setup.sh setup"
    echo ""
}

requirements() {
    found=0

    azureversion=$(az --version)
    if [ $? -eq 0 ]; then
        found=$((found + 1))
        echo "Found azure-cli version: $azureversion"
    else
        echo "azure-cli is missing. Please install azure-cli from"
        echo "https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest"
        echo "Alternatively, you can use the Cloud Shell https://docs.microsoft.com/en-us/azure/cloud-shell/overview right from the Azure Portal or even VS Code."
    fi

    jqversion=$(jq --version)
    if [ $? -eq 0 ]; then
        found=$((found + 1))
        echo "Found jq version: $jqversion"
    else
        echo "jq is missing. Please install jq from"
        echo "https://stedolan.github.io/jq/"
    fi

    if [ $found -lt 2 ]; then
        exit 1
    fi
}

askSubscription() {
    az account list -otable
    echo ""
    echo "Please enter the Id of the account you wish to use. If you do not see"
    echo "a valid account in the list press Ctrl+C to abort and create one."
    echo "If you leave this blank we will use the Current account."
    echo -n "> "
    read azure_subscription_id

    if [ "$azure_subscription_id" != "" ]; then
        az account set --subscription $azure_subscription_id
    else
        azure_subscription_id=$(az account list --output json | jq -r '.[] | select(.isDefault==true) | .id')
    fi
    azure_tenant_id=$(az account list --output json | jq -r '.[] | select(.id=="'$azure_subscription_id'") |  .tenantId')
    echo "Using subscription_id: $azure_subscription_id"
    echo "Using tenant_id: $azure_tenant_id"
}

askName() {
    echo ""
    echo "Choose a name for your resource group, storage account and client"
    echo "client. This is arbitrary, but it must not already be in use by"
    echo "any of those resources. ALPHANUMERIC ONLY. Ex: mypackerbuild"
    echo -n "> "
    read meta_name
}

askSecret() {
    echo ""
    echo "Enter a secret for your application. We recommend generating one with"
    echo "openssl rand -base64 24. If you leave this blank we will attempt to"
    echo "generate one for you using openssl. THIS WILL BE SHOWN IN PLAINTEXT."
    echo "Ex: mypackersecret8734"
    echo -n "> "
    read azure_client_secret
    if [ "$azure_client_secret" = "" ]; then
        azure_client_secret=$(openssl rand -base64 24)
        if [ $? -ne 0 ]; then
            echo "Error generating secret"
            exit 1
        fi
        echo "Generated client_secret: $azure_client_secret"
    fi
}

askLocation() {
    az account list-locations -otable
    echo ""
    echo "Choose which region your resource group and storage account will be created.  example: westus"
    echo -n "> "
    read location
}

createResourceGroup() {
    echo "==> Creating resource group"
    az group create -n $meta_name -l $location
    if [ $? -eq 0 ]; then
        azure_group_name=$meta_name
    else
        echo "Error creating resource group: $meta_name"
        return 1
    fi
}

createStorageAccount() {
    echo "==> Creating storage account"
    az storage account create --name $meta_name --resource-group $meta_name --location $location --kind Storage --sku Standard_LRS
    if [ $? -eq 0 ]; then
        azure_storage_name=$meta_name
    else
        echo "Error creating storage account: $meta_name"
        return 1
    fi
}

createApplication() {
    echo "==> Creating application"
    echo "==> Does application exist?"
    azure_client_id=$(az ad app list --output json | jq -r '.[] | select(.displayName | contains("'$meta_name'")) ')

    if [ "$azure_client_id" != "" ]; then
        echo "==> application already exist, grab appId"
        azure_client_id=$(az ad app list --output json | jq -r '.[] | select(.displayName | contains("'$meta_name'")) .appId')
    else
        echo "==> application does not exist"
        azure_client_id=$(az ad app create --display-name $meta_name --identifier-uris http://$meta_name --homepage http://$meta_name --password $azure_client_secret --output json | jq -r .appId)
    fi

    if [ $? -ne 0 ]; then
        echo "Error creating application: $meta_name @ http://$meta_name"
        return 1
    fi
}

createServicePrincipal() {
    echo "==> Creating service principal"
    azure_object_id=$(az ad sp create --id $azure_client_id --output json | jq -r .objectId)
    echo $azure_object_id "was selected."

    if [ $? -ne 0 ]; then
        echo "Error creating service principal: $azure_client_id"
        return 1
    fi
}

createPermissions() {
    echo "==> Creating permissions"
    az role assignment create --assignee $azure_object_id --role "Owner" --scope /subscriptions/$azure_subscription_id
    # If the user wants to use a more conservative scope, she can.  She must
    # configure the Azure builder to use build_resource_group_name.  The
    # easiest solution is subscription wide permission.
    # az role assignment create --spn http://$meta_name -g $azure_group_name -o "API Management Service Contributor"
    if [ $? -ne 0 ]; then
        echo "Error creating permissions for: http://$meta_name"
        return 1
    fi
}

showConfigs() {
    echo ""
    echo "Use the following configuration for your packer template:"
    echo ""
    echo "{"
    echo "      \"client_id\": \"$azure_client_id\","
    echo "      \"client_secret\": \"$azure_client_secret\","
    echo "      \"object_id\": \"$azure_object_id\","
    echo "      \"subscription_id\": \"$azure_subscription_id\","
    echo "      \"tenant_id\": \"$azure_tenant_id\","
    echo "      \"resource_group_name\": \"$azure_group_name\","
    echo "      \"storage_account\": \"$azure_storage_name\","
    echo "}"
    echo ""
}

doSleep() {
    local sleep_time=${PACKER_SLEEP_TIME-$create_sleep}
    echo ""
    echo "Sleeping for ${sleep_time} seconds to wait for resources to be "
    echo "created. If you get an error about a resource not existing, you can "
    echo "try increasing the amount of time we wait after creating resources "
    echo "by setting PACKER_SLEEP_TIME to something higher than the default."
    echo ""
    sleep $sleep_time
}

retryable() {
    n=0
    until [ $n -ge $1 ]
    do
        $2 && return 0
        echo "$2 failed. Retrying..."
        n=$[$n+1]
        doSleep
    done
    echo "$2 failed after $1 tries. Exiting."
    exit 1
}


setup() {
    requirements

    az login

    askSubscription
    askName
    askSecret
    askLocation

    # Some of the resources take a while to converge in the API. To make the
    # script more reliable we'll add a sleep after we create each resource.

    retryable 3 createResourceGroup
    retryable 3 createStorageAccount
    retryable 3 createApplication
    retryable 3 createServicePrincipal
    retryable 3 createPermissions

    showConfigs
}

case "$1" in
    requirements)
        requirements
        ;;
    setup)
        setup
        ;;
    *)
        showhelp
        ;;
esac
