PACKER=$GOPATH/src/github.com/mitchellh/packer
AZURE=$HOME/hashi/packer-azure

# copy things
cp -r $AZURE/packer/builder/azure $PACKER/builder/azure
# remove legacy API client
rm -rf $PACKER/builder/azure/smapi
cp -r $AZURE/packer/communicator/* $PACKER/communicator/
cp -r $AZURE/packer/provisioner/azureVmCustomScriptExtension $PACKER/provisioner/azureVmCustomScriptExtension

# fix imports
find $PACKER/builder/azure/ -type f | grep ".go" | xargs sed -i -e 's/Azure\/azure-sdk-for-go/mitchellh\/packer\/builder/g'
