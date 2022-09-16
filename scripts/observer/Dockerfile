# build builder
FROM golang:1.16 as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY . .
# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOPROXY=https://goproxy.cn go build -a -o cable pkg/cable/main.go

# build server
FROM openanolis/anolisos:8.4-x86_64

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /app

COPY --from=builder /workspace/cable .
RUN chmod +x /app/cable

RUN useradd -m admin
ENV OB_VERSION 4.0.0.0-100000112022082910
RUN yum install -y python2
RUN ln -s /usr/bin/python2 /usr/bin/python
RUN rpm -ivh http://yum.tbsite.net/alios/7/os/x86_64/Packages/libaio-0.3.109-13.1.alios7.x86_64.rpm
RUN rpm -ivh --nodeps http://yum.tbsite.net/taobao/7/x86_64/test/oceanbase/oceanbase-${OB_VERSION}.el7.x86_64.rpm

## 增加了这3行，替换observer的binary文件, ./custom/bin/observer 是自定义的 binary 文件, 替换 admin 目录下的文件
RUN rm -rf /home/admin/oceanbase/admin
ADD ./custom/bin/observer /home/admin/oceanbase/bin/
ADD ./custom/admin /home/admin/oceanbase/

ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini
ENTRYPOINT ["/tini", "--"]
CMD ["/app/cable"]
