FROM openanolis/anolisos:8.4-x86_64
ARG VERSION
WORKDIR /home/admin/obagent
RUN yum install -y http://mirrors.aliyun.com/oceanbase/community/stable/el/8/x86_64/obagent-${VERSION}.el8.x86_64.rpm
ADD replace_properties.sh /home/admin/obagent
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

ENTRYPOINT [ "bash", "-c", "if [ -e conf/started ]; then echo 'already started'; else echo 'start for the first time' && ./replace_properties.sh && touch conf/started; fi && bin/ob_monagent -c conf/monagent.yaml" ]
