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
# Note: Alpine paths differ from FHS — wpa_supplicant/wpa_cli at /sbin/, tcpdump at /usr/bin/
RUN mkdir -p /export && \
    for bin in /sbin/wpa_supplicant /sbin/wpa_cli /usr/bin/wg /usr/bin/htop /usr/bin/tcpdump; do \
        dir="/export$(dirname $bin)" && mkdir -p "$dir" && cp "$bin" "$dir/"; \
        ldd "$bin" 2>/dev/null | awk '/=>/{print $3}' | while read lib; do \
            if [ -n "$lib" ] && [ -f "$lib" ]; then \
                dir="/export$(dirname $lib)" && mkdir -p "$dir" && cp -n "$lib" "$dir/" 2>/dev/null || true; \
            fi; \
        done; \
    done && \
    mkdir -p /export/usr/bin && cp /usr/bin/wg-quick /export/usr/bin/ && \
    mkdir -p /export/sbin && cp /bin/busybox /export/sbin/udhcpc && \
    mkdir -p /export/usr/share/udhcpc && cp /usr/share/udhcpc/default.script /export/usr/share/udhcpc/ && \
    mkdir -p /export/usr/lib/wpa-compat && \
    cp /usr/lib/libcrypto.so.3 /export/usr/lib/wpa-compat/ && \
    cp /usr/lib/libssl.so.3 /export/usr/lib/wpa-compat/

FROM base-kairos AS discovery-installed
# Install WiFi and networking tools from Alpine (musl-compatible with Hadron)
# Use rsync --ignore-existing to avoid overwriting Hadron's own system libraries
# Alpine's OpenSSL is placed in /usr/lib/wpa-compat/ to avoid conflicting with Hadron's OpenSSL
# (wpa_supplicant needs EVP_rc4/EVP_md4 from Alpine's OpenSSL 3.3, not available in Hadron's 3.6)
RUN --mount=from=alpine-deps,src=/export,dst=/alpine-export \
    rsync -a --ignore-existing /alpine-export/ /

# Create musl soname symlink for dracut-install compatibility:
# Alpine binaries have NEEDED: libc.musl-<arch>.so.1, but Hadron installs musl as /usr/lib/libc.so
RUN ARCH=$(uname -m) && \
    ln -sf /lib/ld-musl-${ARCH}.so.1 /usr/lib/libc.musl-${ARCH}.so.1 && \
    echo "Created musl compat symlink for ${ARCH}"

# Verify that wpa_supplicant (with Alpine's OpenSSL) and udhcpc work correctly
RUN LD_LIBRARY_PATH=/usr/lib/wpa-compat wpa_supplicant -v && udhcpc --help 2>&1 | head -1

# Setup initrd WiFi using dracut
# - Create a custom dracut module for WiFi support in initramfs
# - Include WiFi kernel modules and firmware
# - Add hook to start WiFi during initramfs boot
COPY --chmod=755 ./initramfs/initramfs-* /usr/sbin/
COPY --chmod=755 ./initramfs/dracut/90wifi/ /usr/lib/dracut/modules.d/90wifi/
RUN mkdir -p /etc/dracut.conf.d
COPY ./initramfs/dracut/wifi.conf /etc/dracut.conf.d/wifi.conf

# Rebuild initramfs with dracut including WiFi support
RUN dracut -f --regenerate-all

# Copy custom system config into /system/oem
COPY --chmod=644 ./system-oem/ /system/oem/

# Install discovery
RUN rm -f /system/discovery/kcrypt-discovery-challenger
COPY --from=builder /workdir/kairos-re-unlock /system/discovery/kcrypt-discovery-re-unlock

# Install wg-quick@.service systemd unit template (not included in Alpine's wireguard-tools)
COPY --chmod=644 ./system-files/wg-quick@.service /usr/lib/systemd/system/wg-quick@.service

# Configure WireGuard with systemd (replaces OpenRC setup)
# Sysctl settings are applied automatically on boot by systemd-sysctl.service
RUN echo 'net.ipv4.ip_forward = 1' >> /etc/sysctl.d/99-wireguard.conf && \
    echo 'net.ipv6.conf.all.forwarding = 1' >> /etc/sysctl.d/99-wireguard.conf && \
    echo 'net.ipv6.conf.default.forwarding = 1' >> /etc/sysctl.d/99-wireguard.conf && \
    systemctl enable wg-quick@wg0
