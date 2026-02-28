# DO NOT EDIT MANUALLY - THIS FILE IS AUTO-GENERATED
ARG BASE_IMAGE=ubuntu:20.04
ARG KAIROS_INIT=v0.7.0

FROM quay.io/kairos/kairos-init:${KAIROS_INIT} AS kairos-init

FROM ${BASE_IMAGE} AS base-kairos
ARG MODEL=generic
ARG TRUSTED_BOOT=false
ARG KUBERNETES_DISTRO
ARG KUBERNETES_VERSION
ARG VERSION
ARG FIPS=no-fips

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
    if [ "$FIPS" == "fips" ]; then FIPS_FLAG="--fips"; else FIPS_FLAG=""; fi; \
    eval /kairos-init -l debug -s install -m \"${MODEL}\" -t \"${TRUSTED_BOOT}\" ${K8S_FLAG} ${K8S_VERSION_FLAG} --version \"${VERSION}\" \"${FIPS_FLAG}\" && \
    eval /kairos-init -l debug -s init -m \"${MODEL}\" -t \"${TRUSTED_BOOT}\" ${K8S_FLAG} ${K8S_VERSION_FLAG} --version \"${VERSION}\" \"${FIPS_FLAG}\"
# Build binary
FROM golang:1.25.5 AS builder
WORKDIR /workdir
COPY . .
ENV CGO_ENABLED=0
RUN go mod download
RUN go build -o kairos-re-unlock ./droplet/main.go

# Alpine stage for musl-compatible packages (both Hadron and Alpine use musl libc)
FROM alpine:3.21 AS alpine-deps
RUN apk add --no-cache wpa_supplicant wireguard-tools htop tcpdump
# Prepare exports: binaries + all shared library dependencies
RUN mkdir -p /export && \
    for bin in /usr/sbin/wpa_supplicant /usr/sbin/wpa_cli /usr/bin/wg /usr/bin/htop /usr/sbin/tcpdump; do \
        dir="/export$(dirname $bin)" && mkdir -p "$dir" && cp "$bin" "$dir/"; \
        ldd "$bin" 2>/dev/null | awk '/=>/{print $3}' | while read lib; do \
            if [ -n "$lib" ] && [ -f "$lib" ]; then \
                dir="/export$(dirname $lib)" && mkdir -p "$dir" && cp -n "$lib" "$dir/" 2>/dev/null || true; \
            fi; \
        done; \
    done && \
    mkdir -p /export/usr/bin && cp /usr/bin/wg-quick /export/usr/bin/

FROM base-kairos as discovery-installed
# Install WiFi and networking tools from Alpine (musl-compatible with Hadron)
# Use rsync --ignore-existing to avoid overwriting Hadron's own system libraries
RUN --mount=from=alpine-deps,src=/export,dst=/alpine-export \
    rsync -a --ignore-existing /alpine-export/ /

# Verify that wpa_supplicant works correctly after installation
RUN wpa_supplicant -v

# Setup initrd WiFi using dracut
# - Create a custom dracut module for WiFi support in initramfs
# - Include WiFi kernel modules and firmware
# - Add hook to start WiFi during initramfs boot
COPY --chmod=755 ./initramfs/initramfs-* /usr/sbin/
COPY --chmod=755 ./initramfs/dracut/90wifi/ /usr/lib/dracut/modules.d/90wifi/
COPY ./initramfs/dracut/wifi.conf /etc/dracut.conf.d/wifi.conf

# Rebuild initramfs with dracut including WiFi support
RUN dracut -f --regenerate-all

# Copy custom system config into /system/oem
COPY --chmod=644 ./system-oem/ /system/oem/

# Install discovery
RUN rm -f /system/discovery/kcrypt-discovery-challenger
COPY --from=builder /workdir/kairos-re-unlock /system/discovery/kcrypt-discovery-re-unlock

# Configure WireGuard with systemd (replaces OpenRC setup)
RUN echo 'net.ipv4.ip_forward = 1' >> /etc/sysctl.d/99-wireguard.conf && \
    echo 'net.ipv6.conf.all.forwarding = 1' >> /etc/sysctl.d/99-wireguard.conf && \
    echo 'net.ipv6.conf.default.forwarding = 1' >> /etc/sysctl.d/99-wireguard.conf && \
    systemctl enable wg-quick@wg0
