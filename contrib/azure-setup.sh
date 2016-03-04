#!/usr/bin/env bash
set -e

AZURE_APP_NAME=
AZURE_RESOURCE_GROUP=
AZURE_STORAGE_ACCOUNT=
AZURE_APPLICATION_NAME=
AZURE_APPLICATION_URL=

showhelp() {
	echo ""
	echo "  azure-setup helps automate setting up an Azure account for packer builds"
	echo ""
	echo "  The script walks through the process of creating a resource group,"
	echo "  storage account, application, service principal, and permissions"
	echo "  and then creates the account and shows you the identifiers you need"
	echo "  to configure packer."
	echo ""
	echo "  azure-setup is meant to be run interactively and will prompt you"
	echo "  for input. Also, it assumes you will run this against an account"
	echo "  that has not previously been configured, or that you are OK with"
	echo "  creating all new resources in Azure. If you want to skip or"
	echo "  customize these steps, please use the azure-cli directly."
	echo ""
	echo "  If the script fails partway through you will have to clean up the"
	echo "  lingering resources yourself."
	echo ""
	echo "REQUIREMENTS"
	echo ""
	echo "  You must install the azure-cli from"
	echo "  https://azure.microsoft.com/en-us/documentation/articles/xplat-cli-install/"
	echo "  and jq from https://stedolan.github.io/jq/"
	echo ""
	echo "  azure-setup will verify these tools are available before starting"
	echo ""
	echo "USAGE"
	echo ""
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

askResourceGroup() {
	echo -n "Choose a name for the resource"
	read 
}

setup() {
	requirements
	echo ""
	echo "Note: Please only use alphanumeric names for Azure resources. For"
	echo "example:"
	echo ""
	echo "  Good: packertest"
	echo "  Bad: packer-test"
	echo ""
	
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
