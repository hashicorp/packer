#! /bin/bash
dir=`cat /etc/my.cnf |grep datadir|cut -b 9-`
echo $dir
mysql_install_db
chown -R mysql $dir
service mysqld start
mysqladmin -u root password 'password'
mysql -uroot -ppassword -e "create database wordpress;"
mysql -uroot -ppassword -e "grant REPLICATION SLAVE ON *.* to 'root'@'%' identified by 'password';"
