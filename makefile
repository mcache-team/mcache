APP=mcache
Version=1.0

.PHONY: help all build run image release e2e-raft e2e-raft-quorum e2e-raft-rolling

help:
	@echo "Usage: make <option>"
	@echo "options and effects:"
	@echo " help   : Show help"
	@echo " all    : Build and run"
	@echo " build  : Build the binary of this project"
	@echo " run    : Run server"
	@echo " image  : Build docker images"
	@echo " e2e-raft : Run the raft cluster smoke test"
	@echo " e2e-raft-quorum : Run the raft quorum-loss recovery test"
	@echo " e2e-raft-rolling : Run the raft rolling-restart recovery test"

all: build run

build:
	CGO_ENABLED=0 go build -o bin/mcache pkg/main.go

run:
	bin/mcache

image:
	docker build . -t ${APP}:${Version}

image-cn:
	docker build . -t ${APP}:${Version} --build-arg GOPROXY=https://goproxy.cn,direct

e2e-raft:
	bash e2e/raft-start.sh

e2e-raft-quorum:
	bash e2e/raft-quorum-start.sh

e2e-raft-rolling:
	bash e2e/raft-rolling-restart-start.sh
