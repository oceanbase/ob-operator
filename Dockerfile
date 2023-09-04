# Build the manager binary
FROM golang:1.20.4 as builder

ARG GOPROXY
ARG GOSUMDB

WORKDIR /workspace
# copy everything
COPY . .
RUN GO11MODULE=ON CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -race -o manager cmd/main.go

# start build docker image
FROM openanolis/anolisos:8.4-x86_64
WORKDIR /
COPY --from=builder /workspace/manager .
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
USER 65532:65532

ENTRYPOINT ["/manager"]
