#!/bin/bash
set -e

## This scripts is used to build nightly containers out of nightly builds of GlusterD2 and GlusterFS.
## Buildah and Ansible are used to build the container, and need to be installed on the host used.
## Buildah >= 1.0.0
## Ansible >= 2.4

IMG_NAME="gluster/glusterd2-nightly"
IMG_VERSION=$(date +%Y%m%d)

echo "# Building $IMG_NAME:$IMG_VERSION"


echo "## Setting up build container using buildah"
# Base image on CentOS:7
CID=$(buildah from centos:7)

# Use Ansible to provision the build container
echo "## Provisioning build container using Ansible"

INVENTORY=$(mktemp)
cat > "$INVENTORY" <<EOF
builder ansible_host=$CID ansible_connection=buildah
EOF

ansible-playbook -i "$INVENTORY" ./provision.yml

rm -f "$INVENTORY"

GLFS_VERSION=$(buildah run "$CID" rpm -q --queryformat '%{VERSION}-%{RELEASE}' glusterfs-server)
GD2_VERSION=$(buildah run "$CID" rpm -q --queryformat '%{VERSION}-%{RELEASE}' glusterd2)

# Set metadata/labels/etc.
echo "## Configuring image options and metadata"
buildah config --author "Kaushal M <kshlmster@gmail.com>" \
               --arch "x86_64" \
               --created-by "Buildah + Ansible" \
               --label "name=$IMG_NAME" \
               --label "version=$IMG_VERSION" \
               --label "glusterfs-version=$GLFS_VERSION" \
               --label "glusterd2-version=$GD2_VERSION" \
               --label "vendor=Gluster Community" \
               --label "summary=Image with GlusterD2 and GlusterFS nightly builds from $IMG_VERSION" \
               --label "io.k8s.display-name=(GlusterFS+GD2)nightly on CentOS7" \
               --label "io.openshift.tags=gluster,glusterfs,glusterfs-centos,glusterd2,gcs" \
               --label "description=CentOS7 based image with GlusterFS and GlusterD2 nightly rpms installed and configured to be deployed on Kubernetes or Openshift" \
               --label "io.k8s.description=CentOS7 based image with GlusterFS and GlusterD2 nightly rpms installed and configured to be deployed on Kubernetes or Openshift" \
               "$CID"


# Setup final config options
buildah config --volume /sys/fs/cgroup --port 24007 --port 24008 --cmd /usr/sbin/init "$CID"

# Commit image
echo "## Commiting image"
buildah commit --squash --rm "$CID" "$IMG_NAME:$IMG_VERSION"

echo "# Image $IMG_NAME:$IMG_VERSION created with glusterfs-$GLFS_VERSION and glusterd2-$GD2_VERSION"
