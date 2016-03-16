PACKER=$GOPATH/src/github.com/mitchellh/packer
AZURE=/tmp/packer-azure

ls $AZURE >/dev/null || git clone https://github.com/Azure/packer-azure /tmp/packer-azure
PWD=`pwd`
cd $AZURE && git pull
cd $PWD

# copy things
cp -r $AZURE/packer/builder/azure $PACKER/builder/
cp -r $AZURE/packer/communicator/* $PACKER/communicator/
cp -r $AZURE/packer/provisioner/azureVmCustomScriptExtension $PACKER/provisioner/

# remove legacy API client
rm -rf $PACKER/builder/azure/smapi

# fix imports
find $PACKER/builder/azure/ -type f | grep ".go" | xargs sed -e 's/Azure\/packer-azure\/packer\/builder\/azure/mitchellh\/packer\/builder\/azure/g' -i ''
