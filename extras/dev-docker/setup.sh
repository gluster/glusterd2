#! /bin/bash

# Install required stuff
yum install -y gdb tmux curl which gcc git mercurial bzr subversion

# Setup Go
if ! yum install -y 'golang >= 1.6'
then
  GOURL=https://storage.googleapis.com/golang
  GOVERSION=1.6.3
  GOARCHIVE=go${GOVERSION}.linux-amd64.tar.gz
  mkdir /usr/local
  curl -o /tmp/${GOARCHIVE} -L ${GOURL}/${GOARCHIVE}
  tar -C /usr/local -xzf /tmp/${GOARCHIVE}
  rm -f /tmp/${GOARCHIVE}
fi

# Setup GOPATH
GOPATH=/go
GOPATH_PROFILE=/etc/profile.d/gopath.sh
mkdir -p /go
chown vagrant: /go
## Create $GOPATH_PROFILE to ensure GOPATH is setup for all users
cat >${GOPATH_PROFILE} <<EOF
#!/bin/sh
export GOPATH=$GOPATH
export PATH=\$GOPATH/bin:/usr/local/go/bin:\$PATH
EOF
chmod +x ${GOPATH_PROFILE}

#Install GlusterFS
yum install -y centos-release-gluster
yum install -y glusterfs-server

# Setup BASH prompt to show container IPs
PROMPT_FILE=/etc/profile.d/0-prompt.sh
cat >${PROMPT_FILE} <<'EOF'
#!/bin/bash
export PS1="[\u@\h($(hostname -i)) \W]\$ "
EOF

# Cleanup
yum clean all
