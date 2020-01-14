FROM centos:8 AS builder

ENV D=/home/bridge

WORKDIR $D
COPY . $D/

RUN yum install epel-release -y && \
    #    yum -y install centos-release-scl && \
    yum update -y --setopt=tsflags=nodocs && \
    yum install qpid-proton-c-devel gcc make --setopt=tsflags=nodocs -y && \
    #    yum install devtoolset-8-gcc* devtoolset-8-make -y && \
    yum clean all

RUN gcc -v && \
    make && \
    mv bridge.c /tmp/
