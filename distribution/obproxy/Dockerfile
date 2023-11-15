FROM openanolis/anolisos:8.8

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
ARG VERSION
ARG RELEASE
ARG ARCH=x86_64
WORKDIR /home/admin
RUN useradd -m admin
RUN rpm -ivh https://mirrors.aliyun.com/oceanbase/community/stable/el/8/${ARCH}/obproxy-ce-${VERSION}-${RELEASE}.el8.${ARCH}.rpm
RUN mv /home/admin/obproxy-${VERSION} /home/admin/obproxy
RUN chown -R admin:admin /home/admin/obproxy
ADD start.sh /home/admin

ENTRYPOINT ["bash", "-c", "/home/admin/start.sh"]
