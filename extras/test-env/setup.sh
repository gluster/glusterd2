#! /bin/bash
set -x -e
TMP=$(mktemp -d)

VERSION=v4.0dev-2
ARCHIVE=glusterd2-${VERSION}-linux-amd64.tar.xz
URL=https://github.com/gluster/glusterd2/releases/download/${VERSION}/${ARCHIVE}

curl -o ${TMP}/${ARCHIVE} -L $URL
tar -C /usr/sbin --xz -xf ${TMP}/${ARCHIVE}
setcap cap_sys_admin+ep /usr/sbin/glusterd2

yum install -y etcd
yum clean all



