FROM node:18-alpine as builder-fe
WORKDIR /workspace
COPY ./ui .
ENV NODE_OPTIONS=--max_old_space_size=5120
RUN yarn
RUN yarn build

FROM golang:1.21.5 as builder-be
ARG GOPROXY=https://goproxy.io,direct
ARG GOSUMDB=sum.golang.org
ARG COMMIT_HASH=unknown
WORKDIR /workspace
COPY . .
RUN CGO_ENABLED=0 go build -o bin/todo ./main.go

# start build docker image
FROM alpine:3.17
WORKDIR /bin
COPY --from=builder-be /workspace/bin/todo /bin/todo
COPY --from=builder-fe /workspace/dist /bin/ui/dist
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
USER root

ENTRYPOINT ["/bin/todo"]
