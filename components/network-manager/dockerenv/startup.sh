#!/bin/sh

wait_file() {
  until [ -f "$1" ]
  do
       echo "Check if $1 is available..."
       sleep 5
  done
  echo "File $1 found"
}


export PATH=/bin:/usr/bin:/usr/local/bin:/sbin:/usr/sbin

if [ ! -e /dev/net/tun ]; then
	echo 'FATAL: cannot start ZeroTier One in container: /dev/net/tun not present.'
	exit 1
fi

#echo "Stop zerotier service..."
service zerotier-one stop

echo "Set permission to /dev/net/tun"
# This is a workaround depicted in https://github.com/zerotier/ZeroTierOne/issues/699
chmod 0666 /dev/net/tun

#echo "Start zerotier service..."
service zerotier-one start

wait_file "/var/lib/zerotier-one/zerotier-one.pid"

pid=$(cat /var/lib/zerotier-one/zerotier-one.pid)
echo "Zerotier-one pid is: $pid"

wait_file "/var/lib/zerotier-one/authtoken.secret"

export ZT_ACCESS_TOKEN="$(cat /var/lib/zerotier-one/authtoken.secret)"
echo "Zerotier-one Auth Token is: $ZT_ACCESS_TOKEN"
env | grep ZT

/nalej/network-manager $@

rc=$?

if [[ $rc != 0 ]]; then exit $rc; fi

# Wait until zerotier-one daemon exits
#wait $pid || { echo "zerotier-one exited" >&2; exit 1; }