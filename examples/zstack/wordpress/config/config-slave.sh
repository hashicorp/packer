#! /bin/bash
if [ ! -n "$master" ] ;then
    echo "cannot find master-host ip, please input it"
    exit 1
fi
isslave=`grep "master-host=$master" /etc/my.cnf`
if [ -n $isslave ] ;then
    echo "start copy mysql slave config"
    /bin/cp -f /etc/my-slave.cnf /etc/my.cnf
    sed -i "s/{{master}}/$master/" /etc/my.cnf
    service mysqld restart
fi
status=`mysql -uroot -ppassword -e "show slave status\G;"|grep Slave|grep -i No`
if [ ! -n "$status" ] ;then
    exit 0
fi
echo "start config mysql slave"
mysql -uroot -ppassword -e "slave stop;"
mysql -uroot -ppassword -e "set GLOBAL SQL_SLAVE_SKIP_COUNTER=1;"
mysql -uroot -ppassword -e "slave start;"
