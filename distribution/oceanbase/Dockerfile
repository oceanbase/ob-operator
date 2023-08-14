FROM golang:1.20.4 as builder
ARG GOPROXY
WORKDIR /workspace
COPY ./oceanbase-helper .
RUN GO11MODULE=ON CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o oceanbase-helper main.go

FROM openanolis/anolisos:8.4-x86_64
WORKDIR /home/admin/oceanbase
RUN mkdir -p /home/admin/oceanbase/bin
COPY --from=builder /workspace/oceanbase-helper /home/admin/oceanbase/bin
RUN yum -y install python27
RUN pip2 install mysql-connector -i http://mirrors.aliyun.com/pypi/simple/ --trusted-host mirrors.aliyun.com

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

RUN yum install -y http://mirrors.aliyun.com/oceanbase/community/stable/el/8/x86_64/oceanbase-ce-libs-4.1.0.1-102000042023061314.el8.x86_64.rpm
RUN yum install -y http://mirrors.aliyun.com/oceanbase/community/stable/el/8/x86_64/oceanbase-ce-4.1.0.1-102000042023061314.el8.x86_64.rpm

ENV LD_LIBRARY_PATH /home/admin/oceanbase/lib
