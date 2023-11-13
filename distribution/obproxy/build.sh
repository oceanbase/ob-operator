#!/bin/bash
docker build -t $1:$2-$3 --build-arg VERSION=$2 --build-arg RELEASE=$3 --build-arg ARCH=$4 .
