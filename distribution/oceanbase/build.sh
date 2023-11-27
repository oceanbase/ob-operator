#!/bin/bash
# @Params TARGETPLATFORM linux/amd64 or linux/arm64
docker build -t $1:$2 --build-arg GOPROXY=${GOPROXY} --build-arg VERSION=$2 --build-arg TARGETPLATFORM=$3 .
