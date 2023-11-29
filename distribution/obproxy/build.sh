#!/bin/bash
# @Params TARGETPLATFORM linux/amd64 or linux/arm64
# VERSION: x.y.z-r e.g. 4.2.1-11 which merge old VERSION and RELEASE
docker build -t $1:$2 --build-arg VERSION=$2 --build-arg TARGETPLATFORM=$3 .
