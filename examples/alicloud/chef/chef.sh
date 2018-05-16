#!/bin/sh
#if the related deb pkg not found, please replace with it other avaiable repository url
HOSTNAME=`ifconfig eth1|grep 'inet addr'|cut -d ":" -f2|cut -d " " -f1`
if [ not $HOSTNAME ] ; then
    HOSTNAME=`ifconfig eth0|grep 'inet addr'|cut -d ":" -f2|cut -d " " -f1`
fi
CHEF_SERVER_URL='http://dubbo.oss-cn-shenzhen.aliyuncs.com/chef-server-core_12.8.0-1_amd64.deb'
CHEF_CONSOLE_URL='http://dubbo.oss-cn-shenzhen.aliyuncs.com/chef-manage_2.4.3-1_amd64.deb'
CHEF_SERVER_ADMIN='admin'
CHEF_SERVER_ADMIN_PASSWORD='vmADMIN123'
ORGANIZATION='aliyun'
ORGANIZATION_FULL_NAME='Aliyun, Inc'
#specify hostname
hostname $HOSTNAME

mkdir ~/.pemfile
#install chef server
wget $CHEF_SERVER_URL
sudo dpkg -i chef-server-core_*.deb
sudo chef-server-ctl reconfigure

#create admin user
sudo chef-server-ctl user-create $CHEF_SERVER_ADMIN $CHEF_SERVER_ADMIN $CHEF_SERVER_ADMIN 641002259@qq.com $CHEF_SERVER_ADMIN_PASSWORD -f ~/.pemfile/admin.pem

#create aliyun organization
sudo chef-server-ctl org-create $ORGANIZATION $ORGANIZATION_FULL_NAME --association_user $CHEF_SERVER_ADMIN -f ~/.pemfile/aliyun-validator.pem

#install chef management console
wget $CHEF_CONSOLE_URL
sudo dpkg -i chef-manage_*.deb
sudo chef-server-ctl reconfigure

type expect >/dev/null 2>&1 || { echo >&2 "Install Expect..."; apt-get -y install expect; }
echo "spawn sudo chef-manage-ctl reconfigure" >> chef-manage-confirm.exp
echo "expect \"*Press any key to continue\""  >> chef-manage-confirm.exp
echo "send \"a\\\n\"" >> chef-manage-confirm.exp
echo "expect \".*chef-manage 2.4.3 license: \\\"Chef-MLSA\\\".*\"" >> chef-manage-confirm.exp
echo "send \"q\"" >> chef-manage-confirm.exp
echo "expect \".*Type 'yes' to accept the software license agreement, or anything else to cancel.\"" >> chef-manage-confirm.exp
echo "send \"yes\\\n\"" >> chef-manage-confirm.exp
echo "interact" >> chef-manage-confirm.exp
expect chef-manage-confirm.exp
rm -f chef-manage-confirm.exp

#clean
rm -rf chef-manage_2.4.3-1_amd64.deb
rm -rf chef-server-core_12.8.0-1_amd64.deb
