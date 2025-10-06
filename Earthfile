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
iso:
   WITH DOCKER --load=kcrypt-re-unclock:latest=+image --compose docker-compose.yaml --service aurora-boot
      RUN mv ./build/*.iso ./build/boot.iso
   END
   SAVE ARTIFACT build/boot.iso AS LOCAL ./build/boot.iso
