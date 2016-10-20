FROM gluster/glusterd2-dev:latest

MAINTAINER Kaushal M <kshlmster@gmail.com>

ADD setup.sh /setup.sh
RUN chmod +x /setup.sh
RUN /setup.sh
RUN rm /setup.sh
