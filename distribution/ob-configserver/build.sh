#!/bin/bash
docker build -t $1:$2 --build-arg VERSION=$2 --build-arg ARCH=$3 .
