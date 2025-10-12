FROM golang:1.25.1 AS builder
WORKDIR /workdir
COPY . .
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0
RUN go mod download
RUN go build -o kairos-re-unlock ./droplet/main.go

FROM quay.io/kairos/alpine:3.21-standard-amd64-generic-v3.5.0-k3s-v1.33.2-k3s1
RUN rm -f /system/discovery/kcrypt-discovery-challenger
COPY --from=builder /workdir/kairos-re-unlock /system/discovery/kcrypt-discovery-re-unlock
