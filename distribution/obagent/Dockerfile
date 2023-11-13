FROM openanolis/anolisos:8.8
ARG VERSION
ARG ARCH=x86_64
WORKDIR /home/admin/obagent
RUN yum install -y http://mirrors.aliyun.com/oceanbase/community/stable/el/8/${ARCH}/obagent-${VERSION}.el8.${ARCH}.rpm
ADD replace_properties.sh /home/admin/obagent
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

ENTRYPOINT [ "bash", "-c", "./replace_properties.sh && bin/ob_monagent -c conf/monagent.yaml" ]
