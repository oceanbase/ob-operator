#!/bin/bash
 docker build -t $1 --build-arg GOPROXY=${GOPROXY} .
