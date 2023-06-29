# build builder
FROM golang:1.16 as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY . .
# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOPROXY=https://goproxy.cn go build -a -o cable pkg/cable/main.go

# build server
FROM openanolis/anolisos:8.4-x86_64

RUN yum -y install python27
RUN pip2 install mysql-connector -i http://mirrors.aliyun.com/pypi/simple/ --trusted-host mirrors.aliyun.com 

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /app

COPY --from=builder /workspace/cable .
RUN chmod +x /app/cable

RUN useradd -m admin

ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini

ENV OB_VERSION 4.1.0.1-102000042023061314
RUN rpm -ivh https://mirrors.aliyun.com/oceanbase/community/stable/el/8/x86_64/oceanbase-ce-libs-${OB_VERSION}.el8.x86_64.rpm
RUN rpm -ivh https://mirrors.aliyun.com/oceanbase/community/stable/el/8/x86_64/oceanbase-ce-${OB_VERSION}.el8.x86_64.rpm
RUN chown -R admin:admin /home/admin/oceanbase/
ENV LD_LIBRARY_PATH /home/admin/oceanbase/lib

ENTRYPOINT ["/tini", "--"]
CMD ["/app/cable"]
