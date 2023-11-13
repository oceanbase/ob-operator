FROM openanolis/anolisos:8.8

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

ARG VERSION
ARG ARCH=x86_64
RUN yum install -y http://mirrors.aliyun.com/oceanbase/community/stable/el/8/${ARCH}/ob-configserver-${VERSION}.el8.${ARCH}.rpm
ADD config.yaml.template /home/admin/ob-configserver/conf/
ADD start.sh /home/admin/ob-configserver
WORKDIR /home/admin/ob-configserver

ENTRYPOINT ["bash", "-c", "./start.sh"]
