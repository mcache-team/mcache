#!/bin/bash
set -e

atest run -p test-suite.yaml --request-ignore-error --level trace
