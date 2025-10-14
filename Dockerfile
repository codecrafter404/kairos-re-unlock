ARG ARCH=amd64
ARG BASE_IMAGE=ubuntu:20.04
ARG KAIROS_INIT=v0.6.0-RC1

# Build binary
FROM golang:1.25.1 AS builder
WORKDIR /workdir
COPY . .
ENV CGO_ENABLED=0
RUN go mod download
RUN go build -o kairos-re-unlock ./droplet/main.go


FROM quay.io/kairos/kairos-init:${KAIROS_INIT} AS kairos-init

# Build the image
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

# Install discovery
RUN rm -f /system/discovery/kcrypt-discovery-challenger
COPY --from=builder /workdir/kairos-re-unlock /system/discovery/kcrypt-discovery-re-unlock

# Install wireguard
RUN apk update && \
    apk add wireguard-tools wireguard-tools-openrc iptables && \
    echo 'net.ipv4.ip_forward = 1' >> /etc/sysctl.conf && \
    echo 'net.ipv6.conf.all.forwarding = 1' >> /etc/sysctl.conf && \
    echo 'net.ipv6.conf.default.forwarding = 1' >> /etc/sysctl.conf && \
    ln -s /etc/init.d/wg-quick /etc/init.d/wg-quick.wg0 && \
    rc-update add wg-quick.wg0

# Install other utilities
RUN apk add htop tcpdump
