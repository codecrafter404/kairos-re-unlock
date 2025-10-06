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
   SAVE IMAGE kairos-re-unlock:latest AS LOCAL kairos-re-unlock:latest
up:
   BUILD +build
   BUILD +image
   HOST docker compose up aurora-boot
   HOST mv ./build/*.iso ./build/boot.iso
   HOST docker compose up -d qemu
down:
   HOST docker compose down -v
   HOST rm -rf ./build
