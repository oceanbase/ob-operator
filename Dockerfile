# Build the manager binary
FROM golang:1.20.4 as builder

WORKDIR /workspace
# copy everything
COPY . .
RUN GO11MODULE=ON CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOPROXY=https://goproxy.cn go build -a -o manager cmd/main.go

# start build docker image
FROM openanolis/anolisos:8.4-x86_64
WORKDIR /
COPY --from=builder /workspace/manager .
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
USER 65532:65532

ENTRYPOINT ["/manager"]
