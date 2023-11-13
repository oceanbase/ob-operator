# Build the manager binary
FROM golang:1.20.4 as builder

ARG GOPROXY
ARG GOSUMDB
ARG RACE

WORKDIR /workspace
# copy everything
COPY . .
RUN GO11MODULE=ON CGO_ENABLED=1 GOOS=linux go build ${RACE} -o manager cmd/main.go

# start build docker image
FROM openanolis/anolisos:8.8
WORKDIR /
COPY --from=builder /workspace/manager .
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
USER 65534:65534

ENTRYPOINT ["/manager"]
