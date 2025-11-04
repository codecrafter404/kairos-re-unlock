#!/bin/sh

PID_FILE="/run/initram-wpa_supplicant.pid"

if [ -f "$PID_FILE" ]; then
	kill `cat $PID_FILE`
fi

