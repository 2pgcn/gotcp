#!/bin/bash
go build main.go
ext=$?
if [[ $ext -ne 0 ]]; then
	exit $ext
fi
sudo setcap cap_net_admin=eip ~/pg/gotcp/main
~/pg/gotcp/main  &
pid=$!
sudo ip addr add 192.168.2.0/24 dev tun0
sudo ip link set up dev tun0
trap "kill $pid" INT TERM
wait $pid