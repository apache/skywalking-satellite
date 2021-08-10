cat /proc/net/tcp* | awk '{print $2}'|grep -i :094B
return $?