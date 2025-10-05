FROM golang:1.24.7 as builder
WORKDIR /usr/source/app
COPY .* .
ENV GOOS=linux GOARCH=amd64
RUN "go build -o kairos-re-unlock ."

FROM quay.io/kairos/alpine:3.21-standard-amd64-generic-v3.5.0-k3s-v1.33.2-k3s1

COPY --from builder /usr/source/kairos-re-unlock /system/discovery/kcrypt-re-unlock
RUN rm -f /system/discovery/kcrypt-discovery-challenger
