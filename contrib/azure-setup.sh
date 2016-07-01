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

showhelp() {
	echo "azure-setup"
	echo ""
	echo "  azure-setup helps you generate packer credentials for Azure"
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
	echo "  - https://account.windowsazure.com/"
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

	azureversion=$(azure -v)
	if [ $? -eq 0 ]; then
		found=$((found + 1))
		echo "Found azure-cli version: $azureversion"
	else
		echo "azure-cli is missing. Please install azure-cli from"
		echo "https://azure.microsoft.com/en-us/documentation/articles/xplat-cli-install/"
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
	azure account list
	echo ""
	echo "Please enter the Id of the account you wish to use. If you do not see"
	echo "a valid account in the list press Ctrl+C to abort and create one."
	echo "If you leave this blank we will use the Current account."
	echo -n "> "
	read azure_subscription_id
	if [ "$azure_subscription_id" != "" ]; then
		azure account set $azure_subscription_id
	else
		azure_subscription_id=$(azure account show --json | jq -r .[].id)
	fi
	azure_tenant_id=$(azure account show --json | jq -r .[].tenantId)
	echo "Using subscription_id: $azure_subscription_id"
	echo "Using tenant_id: $azure_tenant_id"
}

askName() {
	echo ""
	echo "Choose a name for your resource group, storage account, and client"
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

createResourceGroup() {
	echo "==> Creating resource group"
	azure group create -n $meta_name -l westus
	if [ $? -eq 0 ]; then
		azure_group_name=$meta_name
	else
		echo "Error creating resource group: $meta_name"
		exit 1
	fi
}

createStorageAccount() {
	echo "==> Creating storage account"
	azure storage account create -g $meta_name -l westus --sku-name LRS --kind Storage $meta_name
	if [ $? -eq 0 ]; then
		azure_storage_name=$meta_name
	else
		echo "Error creating storage account: $meta_name"
		exit 1
	fi
}

createApplication() {
	echo "==> Creating application"
	azure_client_id=$(azure ad app create -n $meta_name -i http://$meta_name --home-page http://$meta_name -p $azure_client_secret --json | jq -r .appId)
	if [ $? -ne 0 ]; then
		echo "Error creating application: $meta_name @ http://$meta_name"
		exit 1
	fi
}

createServicePrinciple() {
	echo "==> Creating service principal"
	azure ad sp create $azure_client_id
	if [ $? -ne 0 ]; then
		echo "Error creating service principal: $azure_client_id"
		exit 1
	fi
}

createPermissions() {
	echo "==> Creating permissions"
	azure role assignment create -o "Owner" --spn http://$meta_name -c /subscriptions/$azure_subscription_id
	# We want to use this more conservative scope but it does not work with the
	# current implementation which uses temporary resource groups
	# azure role assignment create --spn http://$meta_name -g $azure_group_name -o "API Management Service Contributor"
	if [ $? -ne 0 ]; then
		echo "Error creating permissions for: http://$meta_name"
		exit 1
	fi
}

showConfigs() {
	echo ""
	echo "Use the following configuration for your packer template:"
	echo ""
	echo "      \"client_id\": \"$azure_client_id\","
	echo "      \"client_secret\": \"$azure_client_secret\","
	echo "      \"resource_group_name\": \"$azure_group_name\","
	echo "      \"storage_account\": \"$azure_storage_name\","
	echo "      \"subscription_id\": \"$azure_subscription_id\","
	echo ""
}

setup() {
	requirements

	azure config mode arm
	azure login

	askSubscription
	askName
	askSecret

	# Some of the resources take a while to converge in the API. To make the
	# script more reliable we'll add a sleep after we create each resource.

	createResourceGroup
	sleep 5
	createStorageAccount
	sleep 5
	createApplication
	sleep 5
	createServicePrinciple
	sleep 5
	createPermissions

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
