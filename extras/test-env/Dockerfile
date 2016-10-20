FROM gluster/glusterd2-dev:latest

MAINTAINER Kaushal M <kshlmster@gmail.com>

ARG GD2_VERSION
ENV GD2_VERSION $GD2_VERSION

ADD setup.sh /setup.sh
RUN chmod +x /setup.sh
RUN /setup.sh
RUN rm /setup.sh
