FROM golang:1.25.1 AS builder
WORKDIR /workdir
COPY . .
ENV GOOS=linux
ENV GOARCH=amd64
RUN go mod download
RUN go build -o kairos-re-unlock .

FROM quay.io/kairos/alpine:3.21-standard-amd64-generic-v3.5.0-k3s-v1.33.2-k3s1
COPY --from=builder /workdir/kairos-re-unlock /system/discovery/kairos-re-unlock
RUN rm -f /system/discovery/kcrypt-discovery-challenger
