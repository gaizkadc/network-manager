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

# Set of predefined credentials
#echo "2af532a281:0:2db710cf8e4cb83cde5ba711214ffcef4f4b2a898e76dc578c1cb946fc3ad33164319554a22a2c3cd8874fdb934ea00ff09e1d9dc351eeaea8c20efcc1bc16dc" > /var/lib/zerotier-one/identity.public \
#echo "2af532a281:0:2db710cf8e4cb83cde5ba711214ffcef4f4b2a898e76dc578c1cb946fc3ad33164319554a22a2c3cd8874fdb934ea00ff09e1d9dc351eeaea8c20efcc1bc16dc:087d3f27276eed35f55ffc663705b4b71df74885293812d599b4073fb9e4a13f63ddaff6eced247f8b93025b8454daa8e9ac559c9bb114325b742f0b093da874" > /var/lib/zerotier-one/identity.secret \
#echo "072acoozr24inpyci48t7ior" > /var/lib/zerotier-one/authtoken.secret

./zerotier-one &

wait_file "/var/lib/zerotier-one/zerotier-one.pid"

pid=$(cat /var/lib/zerotier-one/zerotier-one.pid)
echo "Zerotier-one pid is: $pid"

# Wait until zerotier-one daemon exits
wait $pid || { echo "zerotier-one exited" >&2; exit 1; }