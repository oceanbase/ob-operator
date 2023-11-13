FROM golang:1.20.4 as builder
ARG GOPROXY=https://goproxy.io,direct
WORKDIR /workspace
COPY ./oceanbase-helper .
RUN GO11MODULE=ON CGO_ENABLED=0 GOOS=linux go build -a -o oceanbase-helper main.go

FROM openanolis/anolisos:8.8
ARG VERSION
ARG ARCH=x86_64
WORKDIR /home/admin/oceanbase
RUN mkdir -p /home/admin/oceanbase/bin
COPY --from=builder /workspace/oceanbase-helper /home/admin/oceanbase/bin
RUN yum -y install python27
RUN pip2 install mysql-connector -i http://mirrors.aliyun.com/pypi/simple/ --trusted-host mirrors.aliyun.com
RUN yum install -y http://mirrors.aliyun.com/oceanbase/community/stable/el/8/${ARCH}/oceanbase-ce-libs-${VERSION}.el8.${ARCH}.rpm
RUN yum install -y http://mirrors.aliyun.com/oceanbase/community/stable/el/8/${ARCH}/oceanbase-ce-${VERSION}.el8.${ARCH}.rpm
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
ENV LD_LIBRARY_PATH /home/admin/oceanbase/lib
