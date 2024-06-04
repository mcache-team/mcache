#!/bin/bash
set -e

docker-compose version

docker-compose up --build --exit-code-from testing
