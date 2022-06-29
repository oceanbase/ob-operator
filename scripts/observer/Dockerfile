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
ENV OB_VERSION 3.1.3-10100042022051821
RUN rpm -ivh https://mirrors.aliyun.com/oceanbase/community/stable/el/8/x86_64/oceanbase-ce-libs-${OB_VERSION}.el8.x86_64.rpm
RUN rpm -ivh https://mirrors.aliyun.com/oceanbase/community/stable/el/8/x86_64/oceanbase-ce-${OB_VERSION}.el8.x86_64.rpm
RUN mkdir -p /home/admin/oceanbase/run
RUN mkdir -p /home/admin/oceanbase/store/sort_dir
RUN mkdir -p /home/admin/oceanbase/store/sstable
RUN mkdir -p /home/admin/oceanbase/store/clog
RUN mkdir -p /home/admin/oceanbase/store/ilog
RUN mkdir -p /home/admin/oceanbase/store/slog
RUN mkdir -p /home/admin/oceanbase/log
RUN chown -R admin:admin /home/admin/oceanbase/

ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini
ENTRYPOINT ["/tini", "--"]
CMD ["/app/cable"]
