FROM ubuntu:24.04

RUN mkdir -p /opt/ss/

WORKDIR /opt/ss/
ENV WORKDIR=/opt/ss/

ADD mecha-proxy /usr/bin/mecha-proxy
CMD /usr/bin/mecha-proxy


# GOOS=linux GOARCH=amd64 go build -o mecha-proxy .
# podman build -t serverless-server-base .