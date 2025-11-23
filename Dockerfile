ARG BASE_IMAGE=ubuntu:20.04
ARG KAIROS_INIT=v0.6.0-RC1

FROM quay.io/kairos/kairos-init:${KAIROS_INIT} AS kairos-init

FROM ${BASE_IMAGE} AS base-kairos
ARG MODEL=generic
ARG TRUSTED_BOOT=false
ARG KUBERNETES_DISTRO
ARG KUBERNETES_VERSION
ARG VERSION

RUN --mount=type=bind,from=kairos-init,src=/kairos-init,dst=/kairos-init \
    if [ -n "${KUBERNETES_DISTRO}" ]; then \
        K8S_FLAG="-p ${KUBERNETES_DISTRO}"; \
        if [ "${KUBERNETES_DISTRO}" = "k0s" ] && [ -n "${KUBERNETES_VERSION}" ]; then \
            K8S_VERSION_FLAG="--provider-k0s-version \"${KUBERNETES_VERSION}\""; \
        elif [ "${KUBERNETES_DISTRO}" = "k3s" ] && [ -n "${KUBERNETES_VERSION}" ]; then \
            K8S_VERSION_FLAG="--provider-k3s-version \"${KUBERNETES_VERSION}\""; \
        else \
            K8S_VERSION_FLAG=""; \
        fi; \
    else \
        K8S_FLAG=""; \
        K8S_VERSION_FLAG=""; \
    fi; \
    eval /kairos-init -l debug -s install -m \"${MODEL}\" -t \"${TRUSTED_BOOT}\" ${K8S_FLAG} ${K8S_VERSION_FLAG} --version \"${VERSION}\" && \
    eval /kairos-init -l debug -s init -m \"${MODEL}\" -t \"${TRUSTED_BOOT}\" ${K8S_FLAG} ${K8S_VERSION_FLAG} --version \"${VERSION}\"

# Build binary
FROM golang:1.25.1 AS builder
WORKDIR /workdir
COPY . .
ENV CGO_ENABLED=0
RUN go mod download
RUN go build -o kairos-re-unlock ./droplet/main.go

FROM base-kairos as discovery-installed
# Setup initrd Wifi by
# - getting the binary dependencies 
# - adding the wifi modules
# - adding the wifi feature to mkinitramfs.conf
# - manipulating the init script to start up wifi
# TODO: optimize the process to rebuild the initramfs when installing in order to only include the necessary wifi drivers
COPY ./initramfs/wifi.* /etc/mkinitfs/features.d
COPY --chmod=755 ./initramfs/initramfs-* /usr/sbin/
RUN ldd /sbin/wpa_supplicant | sed -E "s/.* \//\//" | sed -E "s/ .*//" | tr -d '[:blank:]' >> /etc/mkinitfs/features.d/wifi.files &&\
    ldd /sbin/wpa_cli | sed -E "s/.* \//\//" | sed -E "s/ .*//" | tr -d '[:blank:]' >> /etc/mkinitfs/features.d/wifi.files &&\
    cat /etc/mkinitfs/features.d/wifi.files | sort -u > /etc/mkinitfs/features.d/wifi.files2 &&\
    rm /etc/mkinitfs/features.d/wifi.files && mv /etc/mkinitfs/features.d/wifi.files2 /etc/mkinitfs/features.d/wifi.files &&\
    sed -E "s/\"\$/ wifi\"/" -i /etc/mkinitfs/mkinitfs.conf &&\
    sed '/rd_break\ post-network/i \/usr\/sbin\/initramfs-start-wifi.sh' -i /usr/share/mkinitfs/initramfs-init &&\
    mkinitfs -o /boot/initrd $(ls -1 /lib/modules | tail -n 1)

# Copy custom system config into /system/oem
COPY --chmod=644 ./system-oem/ /system/oem/

# Install discovery
RUN rm -f /system/discovery/kcrypt-discovery-challenger
COPY --from=builder /workdir/kairos-re-unlock /system/discovery/kcrypt-discovery-re-unlock

# Install wireguard
RUN apk update && \
    apk add wireguard-tools wireguard-tools-openrc iptables && \
    echo 'net.ipv4.ip_forward = 1' >> /etc/sysctl.conf && \
    echo 'net.ipv6.conf.all.forwarding = 1' >> /etc/sysctl.conf && \
    echo 'net.ipv6.conf.default.forwarding = 1' >> /etc/sysctl.conf && \
    ln -s /etc/init.d/wg-quick /etc/init.d/wg-quick.wg0

# Install other utilities
RUN apk add htop tcpdump