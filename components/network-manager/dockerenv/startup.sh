#!/bin/sh

wait_file() {
  until [ -f "$1" ]
  do
       echo "Check if $1 is available..."
       sleep 5
  done
  echo "File $1 found"
}

zt_connected() {
    string="$1"
    case "$string" in
      *"OK PRIVATE zt0"*) return 0 ;;
      *)                  return 1 ;;
    esac
}

export PATH=/bin:/usr/bin:/usr/local/bin:/sbin:/usr/sbin

if [ ! -e /dev/net/tun ]; then
	echo 'FATAL: cannot start ZeroTier One in container: /dev/net/tun not present.'
	exit 1
fi

./zerotier-one &

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