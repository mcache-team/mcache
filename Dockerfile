FROM golang:1.21 AS build-stage

WORKDIR /opt/build

COPY . .

RUN go env -w GO111MODULE=on && \
    go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -o /mcache pkg/main.go

FROM ubuntu:24.04 AS build-release-stage

RUN mkdir -p /opt/mcache/bin

WORKDIR /opt

COPY --from=build-stage /mcache /opt/mcache/bin/mcache
ADD boot.sh /opt/mcache/boot.sh

EXPOSE 8080

ENTRYPOINT ["/opt/mcache/boot.sh"]


