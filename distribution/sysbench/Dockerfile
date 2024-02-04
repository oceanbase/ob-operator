FROM ubuntu:24.04

# Install sysbench
RUN apt-get update && \ 
  apt-get install -y curl && \
  curl -s https://packagecloud.io/install/repositories/akopytov/sysbench/script.deb.sh | bash && \
  apt-get install -y sysbench

