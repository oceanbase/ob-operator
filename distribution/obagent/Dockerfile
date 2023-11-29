FROM openanolis/anolisos:8.8
ARG VERSION
ARG TARGETPLATFORM
WORKDIR /home/admin/obagent
RUN if [[ ${TARGETPLATFORM} == 'linux/amd64' ]] ; then yum install -y http://mirrors.aliyun.com/oceanbase/community/stable/el/8/x86_64/obagent-${VERSION}.el8.x86_64.rpm ; else yum install -y http://mirrors.aliyun.com/oceanbase/community/stable/el/8/aarch64/obagent-${VERSION}.el8.aarch64.rpm ; fi
ADD replace_properties.sh /home/admin/obagent
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

ENTRYPOINT [ "bash", "-c", "./replace_properties.sh && bin/ob_monagent -c conf/monagent.yaml" ]
