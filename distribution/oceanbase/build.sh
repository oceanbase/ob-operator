#!/bin/bash
docker build -t $1 --build-arg GOPROXY=$(go env GOPROXY) --build-arg VERSION=$2 .
