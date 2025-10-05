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
   FROM earthly/dind:alpine-3.19-docker-25.0.5-r0
   COPY ./kairos_config.yaml config.yaml
   WITH DOCKER --load image:latest=+image
      RUN docker run -v ./config.yaml:/config.yaml \
             -v ./build:/tmp/auroraboot \
             -v /var/run/docker.sock:/var/run/docker.sock \
             --rm -ti quay.io/kairos/auroraboot \
             --set container_image=docker://image \
             --set "disable_http_server=true" \
             --set "disable_netboot=true" \
             --cloud-config /config.yaml \
             --set "state_dir=/tmp/auroraboot"
   END
   RUN mv build/*.iso build/image.iso
   SAVE ARTIFACT build/image.iso AS LOCAL build/image.iso

