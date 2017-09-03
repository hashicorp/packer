#!/bin/sh

JENKINS_URL='http://mirrors.jenkins.io/war-stable/2.32.2/jenkins.war'

TOMCAT_VERSION='7.0.77'
TOMCAT_NAME="apache-tomcat-$TOMCAT_VERSION"
TOMCAT_PACKAGE="$TOMCAT_NAME.tar.gz"
TOMCAT_URL="http://mirror.bit.edu.cn/apache/tomcat/tomcat-7/v$TOMCAT_VERSION/bin/$TOMCAT_PACKAGE"
TOMCAT_PATH="/opt/$TOMCAT_NAME"

#install jdk
if grep -Eqi "Ubuntu|Debian|Raspbian" /etc/issue || grep -Eq "Ubuntu|Debian|Raspbian" /etc/*-release; then
        sudo apt-get update -y
        sudo apt-get install -y openjdk-7-jdk
elif grep -Eqi "CentOS|Fedora|Red Hat Enterprise Linux Server" /etc/issue || grep -Eq "CentOS|Fedora|Red Hat Enterprise Linux Server" /etc/*-release; then
        sudo yum update -y
        sudo yum install -y openjdk-7-jdk
else
        echo "Unknown OS type."
fi

#install jenkins server
mkdir ~/work
cd ~/work

#install tomcat
wget $TOMCAT_URL
tar -zxvf $TOMCAT_PACKAGE
mv $TOMCAT_NAME /opt

#install
wget $JENKINS_URL
mv jenkins.war $TOMCAT_PATH/webapps/

#set emvironment
echo "TOMCAT_PATH=\"$TOMCAT_PATH\"">>/etc/profile
echo "JENKINS_HOME=\"$TOMCAT_PATH/webapps/jenkins\"">>/etc/profile
echo PATH="\"\$PATH:\$TOMCAT_PATH:\$JENKINS_HOME\"">>/etc/profile
. /etc/profile

#start tomcat & jenkins
$TOMCAT_PATH/bin/startup.sh

#set start on boot
sed -i "/#!\/bin\/sh/a$TOMCAT_PATH/bin/startup.sh" /etc/rc.local

#clean
rm -rf ~/work
