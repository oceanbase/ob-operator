FROM openanolis/anolisos:8.8

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
ARG VERSION
ARG RELEASE
ARG TARGETPLATFORM
WORKDIR /home/admin
RUN useradd -m admin
RUN if [[ ${TARGETPLATFORM} == 'linux/amd64' ]] ; then rpm -ivh https://mirrors.aliyun.com/oceanbase/community/stable/el/8/x86_64/obproxy-ce-${VERSION}.el8.x86_64.rpm ; else rpm -ivh https://mirrors.aliyun.com/oceanbase/community/stable/el/8/aarch64/obproxy-ce-${VERSION}.el8.aarch64.rpm ; fi
RUN mv /home/admin/obproxy-* /home/admin/obproxy
RUN chown -R admin:admin /home/admin/obproxy
ADD start.sh /home/admin

ENTRYPOINT ["bash", "-c", "/home/admin/start.sh"]
