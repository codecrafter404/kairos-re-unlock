#!/bin/bash
# Dracut module for WiFi support during initramfs (replaces Alpine mkinitfs wifi feature)
# This module enables WiFi connectivity in the initramfs for remote LUKS unlocking.

check() {
    require_binaries wpa_supplicant wpa_cli || return 1
    return 0
}

depends() {
    echo "kernel-modules"
    return 0
}

installkernel() {
    # Install WiFi kernel modules (wireless core + device drivers)
    instmods =kernel/net/wireless =kernel/drivers/net/wireless
}

install() {
    # Install wpa_supplicant and wpa_cli for WiFi authentication
    inst_multiple wpa_supplicant wpa_cli

    # Install the WiFi start/stop scripts
    inst_simple /usr/sbin/initramfs-start-wifi.sh
    inst_simple /usr/sbin/initramfs-stop-wifi.sh

    # Install udhcpc for DHCP lease in initramfs
    inst_multiple -o udhcpc

    # Install additional firmware for Raspberry Pi 4 (if available)
    for fw in \
        /lib/firmware/brcm/BCM4345C0.raspberrypi,4-model-b.hcd.zst \
        /lib/firmware/brcm/brcmfmac43455-sdio.raspberrypi,4-model-b.bin.zst \
        /lib/firmware/brcm/brcmfmac43455-sdio.raspberrypi,4-model-b.clm_blob.zst \
        /lib/firmware/brcm/brcmfmac43455-sdio.raspberrypi,4-model-b.txt.zst; do
        [ -f "$fw" ] && inst_simple "$fw"
    done

    # Install the WiFi start hook (runs early in initqueue)
    inst_hook initqueue 80 "$moddir/wifi-start.sh"
}
