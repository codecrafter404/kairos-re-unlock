FROM golang:1.25.1 AS builder
WORKDIR /workdir
COPY . .
ENV GOOS=linux
ENV GOARCH=amd64
RUN go mod download
RUN go build -o kairos-re-unlock .

FROM quay.io/kairos/alpine:3.21-standard-amd64-generic-v3.5.0-k3s-v1.33.2-k3s1
<<<<<<< HEAD
COPY --from builder kairos-re-unlock /system/discovery/kcrypt-discovery-kairos-re-unlock
=======
COPY --from=builder /workdir/kairos-re-unlock /system/discovery/kairos-re-unlock
>>>>>>> 91db5ba2c94dfee4a423447a8ef2d6286f27c9b7
RUN rm -f /system/discovery/kcrypt-discovery-challenger
