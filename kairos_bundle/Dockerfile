FROM golang:1.25.1 AS builder
WORKDIR /workdir
COPY ../ .
ENV GOOS=linux
ENV GOARCH=amd64
RUN go mod download
RUN go build -o kairos-re-unlock .

FROM alpine
COPY --from=builder /workdir/kairos-re-unlock /assets/kcrypt-discovery-challenger
COPY ./run.sh .
