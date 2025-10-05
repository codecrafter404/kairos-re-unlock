VERSION 0.8
FROM golang:1.25.1
WORKDIR /workdir
build:
   COPY . .
   ENV GOOS=linux
   ENV GOARCH=amd64
   RUN go build -o kairos-re-unlock .
   SAVE ARTIFACT kairos-re-unlock
image:
   FROM quay.io/kairos/alpine:3.21-standard-amd64-generic-v3.5.0-k3s-v1.33.2-k3s1
   COPY +build/kairos-re-unlock /system/discovery/kcrypt-re-unlock
   RUN rm -f /system/discovery/kcrypt-discovery-challenger
   SAVE IMAGE kairos-re-unlock:latest
build-iso:
   FROM quay.io/kairos/auroraboot
   RUN /usr/bin/auroraboot --set container_image=+image/kairos-re-unlock \
             --set "disable_http_server=true" \
             --set "disable_netboot=true" \
             --cloud-config ./config.yaml \
             --set "state_dir=/tmp/auroraboot"
   RUN mv /tmp/auroraboot/*.iso build/image.iso
   SAVE ARTIFACT build/image.iso AS LOCAL build/image.iso

