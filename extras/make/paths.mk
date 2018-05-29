# Setting up standard path variables similar to autoconf
# The defaults are taken based on
# https://www.gnu.org/prep/standards/html_node/Directory-Variables.html
# and
# https://fedoraproject.org/wiki/Packaging:RPMMacros?rd=Packaging/RPMMacros

PREFIX ?= /usr/local

BASE_PREFIX = $(PREFIX)
ifeq ($(PREFIX), /usr)
BASE_PREFIX = ""
endif

EXEC_PREFIX ?= $(PREFIX)

BINDIR ?= $(EXEC_PREFIX)/bin
SBINDIR ?= $(EXEC_PREFIX)/sbin

DATADIR ?= $(PREFIX)/share
LOCALSTATEDIR ?= $(BASE_PREFIX)/var/lib
LOGDIR ?= $(BASE_PREFIX)/var/log

SYSCONFDIR ?= $(BASE_PREFIX)/etc
RUNDIR ?= $(BASE_PREFIX)/var/run
