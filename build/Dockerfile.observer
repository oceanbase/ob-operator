FROM oceanbase/oceanbase-helper

ARG VERSION
ARG TARGETPLATFORM
# support docker platform linux/amd64 and linux/arm64, mapping platform to x86_64 or aarch64
RUN if [[ ${TARGETPLATFORM} == 'linux/amd64' ]] ; then yum install -y http://mirrors.aliyun.com/oceanbase/community/stable/el/8/x86_64/oceanbase-ce-libs-${VERSION}.el8.x86_64.rpm ; else yum install -y http://mirrors.aliyun.com/oceanbase/community/stable/el/8/aarch64/oceanbase-ce-libs-${VERSION}.el8.aarch64.rpm ; fi ; yum clean all;
RUN if [[ ${TARGETPLATFORM} == 'linux/amd64' ]] ; then yum install -y http://mirrors.aliyun.com/oceanbase/community/stable/el/8/x86_64/oceanbase-ce-${VERSION}.el8.x86_64.rpm ; else yum install -y http://mirrors.aliyun.com/oceanbase/community/stable/el/8/aarch64/oceanbase-ce-${VERSION}.el8.aarch64.rpm ; fi ; yum clean all
ENV LD_LIBRARY_PATH /home/admin/oceanbase/lib
