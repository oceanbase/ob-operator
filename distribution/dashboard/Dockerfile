FROM node:16-alpine as builder-fe
WORKDIR /workspace
COPY ./ui .
ENV NODE_OPTIONS=--max_old_space_size=3072
RUN yarn --verbose
RUN yarn build

FROM golang:1.20.4 as builder-be
ARG GOPROXY=https://goproxy.io,direct
ARG GOSUMDB=sum.golang.org
WORKDIR /workspace
COPY . .
RUN make dep-install build

# start build docker image
FROM openanolis/anolisos:8.4-x86_64
WORKDIR /root
COPY --from=builder-be /workspace/bin/oceanbase-dashboard .
COPY --from=builder-fe /workspace/dist ./ui/dist
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
USER root

ENTRYPOINT ["/root/oceanbase-dashboard"]
