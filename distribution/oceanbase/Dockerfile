FROM golang:1.20.4 as builder
ARG GOPROXY=https://goproxy.io,direct
WORKDIR /workspace
COPY ./oceanbase-helper .
RUN GO11MODULE=ON CGO_ENABLED=0 GOOS=linux go build -a -o oceanbase-helper main.go

FROM openanolis/anolisos:8.8
ARG VERSION
ARG TARGETPLATFORM
RUN echo TARGET_PLATFORM is ${TARGETPLATFORM}
WORKDIR /home/admin/oceanbase
RUN mkdir -p /home/admin/oceanbase/bin
COPY --from=builder /workspace/oceanbase-helper /home/admin/oceanbase/bin
RUN yum -y install python27
RUN pip2 install mysql-connector -i http://mirrors.aliyun.com/pypi/simple/ --trusted-host mirrors.aliyun.com
# support docker platform linux/amd64 and linux/arm64, mapping platform to x86_64 or aarch64
RUN if [[ ${TARGETPLATFORM} == 'linux/amd64' ]] ; then yum install -y http://mirrors.aliyun.com/oceanbase/community/stable/el/8/x86_64/oceanbase-ce-libs-${VERSION}.el8.x86_64.rpm ; else yum install -y http://mirrors.aliyun.com/oceanbase/community/stable/el/8/aarch64/oceanbase-ce-libs-${VERSION}.el8.aarch64.rpm ; fi
RUN if [[ ${TARGETPLATFORM} == 'linux/amd64' ]] ; then yum install -y http://mirrors.aliyun.com/oceanbase/community/stable/el/8/x86_64/oceanbase-ce-${VERSION}.el8.x86_64.rpm ; else yum install -y http://mirrors.aliyun.com/oceanbase/community/stable/el/8/aarch64/oceanbase-ce-${VERSION}.el8.aarch64.rpm ; fi
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
ENV LD_LIBRARY_PATH /home/admin/oceanbase/lib
