#!/bin/sh
INTERFACE="wlan0"
CONFIG="/tmp/mnt/OEM/wpa.conf"
CONNECT_TIMEOUT=30

find_interface () {
    for interface in /sys/class/net/*; do
	if [ -d "$interface/wireless" ]; then
		INTERFACE=$(basename "$interface")
	fi
    done
}

enable () {
	/sbin/wpa_supplicant -i "$INTERFACE" -c "$CONFIG" -P /run/initram-wpa_supplicant.pid -B -d

	echo -n "Connecting to wifi network"
	while [ "$CONNECT_TIMEOUT" -ge 0 ]
	do
		echo -n "."

		CONNECT_TIMEOUT=$((CONNECT_TIMEOUT - 1))

		LINK_QUALITY=$(cat /proc/net/wireless | grep "${INTERFACE}:" | awk '{ print $3 }' || true)

		# Check if quality is not "0." and not empty
		if [ "$LINK_QUALITY" != "0." ] && [ -n "$LINK_QUALITY" ]; then
		    break
		fi

		sleep 1
	done

	echo ""

	if [ "$CONNECT_TIMEOUT" -eq -1 ]; then
		echo "Connection timeout reached"
		exit -1
	fi

	udhcpc -i "$INTERFACE" -f -q -n -S

	if [ $? -ne 0 ]; then
		echo "Failed to lease ip address"
		exit -1
	fi

}

if [$# -eq 0]; then
	echo -n "Testing wifi config"

	OEM=$(kairos-agent state get oem.name)

	mkdir -p /tmp/mnt/OEM
	mount ${OEM} /tmp/mnt/OEM
	if [ -f "$CONFIG" ]; then
		find_interface
		enable
	fi
	umount /tmp/mnt/OEM
	rm -rf /tmp/mnt/OEM
else
	CONFIG="$1"
	if [ -f "$CONFIG" ]; then
		find_interface
		enable
	fi
fi
