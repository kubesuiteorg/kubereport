FROM golang:1.22.0 AS builder

ARG upx_version=4.2.4

RUN apt-get update && \
    apt-get install -y --no-install-recommends xz-utils curl && \
    curl -Ls https://github.com/upx/upx/releases/download/v${upx_version}/upx-${upx_version}-amd64_linux.tar.xz | \
    tar -xJf - -C /tmp && \
    cp /tmp/upx-${upx_version}-amd64_linux/upx /usr/local/bin/ && \
    chmod +x /usr/local/bin/upx && \
    apt-get remove -y xz-utils curl && \
    rm -rf /var/lib/apt/lists/* /tmp/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o kubereport -a -ldflags="-s -w" -installsuffix cgo

RUN upx --ultra-brute -qq kubereport && upx -t kubereport

FROM alpine:3.18.0

COPY --from=builder /app/kubereport /usr/local/bin/kubereport

WORKDIR /kubereport
VOLUME /kubereport


RUN apk add --no-cache bash

ENTRYPOINT ["kubereport"]