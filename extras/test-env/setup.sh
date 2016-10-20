#! /bin/bash
set -x -e
TMP=$(mktemp -d)

# GD2_VERSION will be set as an ENV variable form the Dockerfile
ARCHIVE=glusterd2-${GD2_VERSION}-linux-amd64.tar.xz
URL=https://github.com/gluster/glusterd2/releases/download/${GD2_VERSION}/${ARCHIVE}

curl -o ${TMP}/${ARCHIVE} -L $URL
tar -C /usr/sbin --xz -xf ${TMP}/${ARCHIVE}
setcap cap_sys_admin+ep /usr/sbin/glusterd2

yum install -y etcd
yum clean all



