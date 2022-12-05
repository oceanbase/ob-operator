# build builder
FROM golang:1.16 as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY . .
# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOPROXY=https://goproxy.cn go build -a -o cable pkg/cable/main.go

FROM centos:8 as builder1
WORKDIR /workspace
RUN curl https://mirrors.aliyun.com/oceanbase/community/stable/el/8/x86_64/oceanbase-ce-3.1.2-10000392021123010.el8.x86_64.rpm --output oceanbase-ce-3.1.2-10000392021123010.el8.x86_64.rpm
RUN rpm2cpio oceanbase-ce-3.1.2-10000392021123010.el8.x86_64.rpm| cpio -div
RUN mkdir 3.1.2
RUN cp -r ./home/admin/oceanbase/etc/* ./3.1.2

RUN curl https://mirrors.aliyun.com/oceanbase/community/stable/el/8/x86_64/oceanbase-ce-3.1.3-10000292022032916.el8.x86_64.rpm --output oceanbase-ce-3.1.3-10000292022032916.el8.x86_64.rpm
RUN mkdir 3.1.3
RUN cp -r ./home/admin/oceanbase/etc/* ./3.1.3

# build server
FROM openanolis/anolisos:8.4-x86_64

RUN yum -y install python27
RUN pip2 install mysql-connector

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /app

COPY --from=builder /workspace/cable .
RUN chmod +x /app/cable

RUN useradd -m admin

RUN mkdir -p /home/admin/oceanbase/scripts/{3.1.2,3.1.3,3.1.4}

COPY --from=builder1 /workspace/3.1.2 /home/admin/oceanbase/scripts/3.1.2
COPY --from=builder1 /workspace/3.1.3 /home/admin/oceanbase/scripts/3.1.3

ENV OB_VERSION 3.1.4-10000092022071511
RUN rpm -ivh https://mirrors.aliyun.com/oceanbase/community/stable/el/8/x86_64/oceanbase-ce-libs-${OB_VERSION}.el8.x86_64.rpm
RUN rpm -ivh https://mirrors.aliyun.com/oceanbase/community/stable/el/8/x86_64/oceanbase-ce-${OB_VERSION}.el8.x86_64.rpm
RUN chown -R admin:admin /home/admin/oceanbase/

RUN cp /home/admin/oceanbase/etc/* /home/admin/oceanbase/scripts/3.1.4

ENV TINI_VERSION v0.19.0
ENV LD_LIBRARY_PATH /home/admin/oceanbase/lib
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini
ENTRYPOINT ["/tini", "--"]
CMD ["/app/cable"]
