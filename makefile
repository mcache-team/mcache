APP=mcache
Version=1.0

.PHONY: help all build run image release

help:
	@echo "Usage: make <option>"
	@echo "options and effects:"
	@echo " help   : Show help"
	@echo " all    : Build and run"
	@echo " build  : Build the binary of this project"
	@echo " run    : Run server"
	@echo " image  : Build docker images"

all: build run

build:
	CGO_ENABLED=0 go build -o bin/mcache pkg/main.go

run:
	bin/mcache

image:
	docker build . -t ${APP}:${Version}

image-cn:
	docker build . -t ${APP}:${Version} --build-arg GOPROXY=https://goproxy.cn,direct
