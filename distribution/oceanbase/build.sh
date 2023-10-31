#!/bin/bash
docker build -t $1:$2 --build-arg GOPROXY=${GOPROXY} --build-arg VERSION=$2 .
