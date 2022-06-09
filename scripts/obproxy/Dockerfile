FROM openanolis/anolisos:8.4-x86_64

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /home/admin

ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini

RUN useradd -m admin
ENV OBPROXY_VERSION 3.2.3
ENV OBPROXY_RELEASE 2
RUN rpm -ivh https://mirrors.aliyun.com/oceanbase/community/stable/el/8/x86_64/obproxy-ce-${OBPROXY_VERSION}-${OBPROXY_RELEASE}.el8.x86_64.rpm
RUN mv /home/admin/obproxy-${OBPROXY_VERSION} /home/admin/obproxy
RUN chown -R admin:admin /home/admin/obproxy

ENTRYPOINT ["/tini", "--"]
CMD ["bash", "-c", "cd /home/admin/obproxy && /home/admin/obproxy/bin/obproxy -p 2883 -n ${APP_NAME} -c ${OB_CLUSTER} -r\"${RS_LIST}\" -o prometheus_sync_interval=1,prometheus_listen_port=2884,enable_metadb_used=false,skip_proxy_sys_private_check=true,log_dir_size_threshold=10G,enable_proxy_scramble=true,enable_strict_kernel_release=false --nodaemon"]
