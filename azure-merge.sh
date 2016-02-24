PACKER=$GOPATH/src/github.com/mitchellh/packer
AZURE=$HOME/hashi/packer-azure

# copy things
cp -r $AZURE/packer/builder/azure $PACKER/builder/azure
cp -r $AZURE/packer/communicator/* $PACKER/communicator/
cp -r $AZURE/packer/provisioner/azureVmCustomScriptExtension $PACKER/provisioner/azureVmCustomScriptExtension

# remove legacy API client
rm -rf $PACKER/builder/azure/smapi

# fix imports
find $PACKER/builder/azure/ -type f | grep ".go" | xargs sed -e 's/Azure\/packer-azure\/packer\/builder\/azure/mitchellh\/packer\/builder\/azure/g' -i ''
