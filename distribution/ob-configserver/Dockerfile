FROM openanolis/anolisos:8.8

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

ARG VERSION
ARG TARGETPLATFORM
RUN if [[ ${TARGETPLATFORM} == 'linux/amd64' ]] ; then yum install -y http://mirrors.aliyun.com/oceanbase/community/stable/el/8/x86_64/ob-configserver-${VERSION}.el8.x86_64.rpm ; else yum install -y http://mirrors.aliyun.com/oceanbase/community/stable/el/8/aarch64/ob-configserver-${VERSION}.el8.aarch64.rpm ; fi
ADD config.yaml.template /home/admin/ob-configserver/conf/
ADD start.sh /home/admin/ob-configserver
WORKDIR /home/admin/ob-configserver

ENTRYPOINT ["bash", "-c", "./start.sh"]
