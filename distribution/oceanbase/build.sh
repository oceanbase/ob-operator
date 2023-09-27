#!/bin/bash
docker build -t $1 --build-arg GOPROXY=${GOPROXY} --build-arg VERSION=$2 .
