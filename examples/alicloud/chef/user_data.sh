HOSTNAME=`ifconfig eth1|grep 'inet addr'|cut -d ":" -f2|cut -d " " -f1`
if [ not $HOSTNAME ] ; then
    HOSTNAME=`ifconfig eth0|grep 'inet addr'|cut -d ":" -f2|cut -d " " -f1`
fi
hostname $HOSTNAME
chef-server-ctl  reconfigure
