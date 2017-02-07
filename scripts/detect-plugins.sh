#!/usr/bin/env bash

# This script fetches the list of plugins avalable in plugins directory

IMPORT_PREFIX=github.com/gluster/glusterd2/plugins
echo "package plugins"
echo

echo "import ("
cd "plugins"
for p in `ls -d */`
do
    # ${p%/} is to remove trailing slash
    echo -e "\t\"${IMPORT_PREFIX}/${p%/}\""
done
echo ")"

echo
echo "var PluginsList = []Gd2Plugin{"

for p in `ls -d */`
do
    p=${p%/}
    # Like &helloplugin.HelloPlugin{}
    echo -e "\t&${p}plugin.${p^}Plugin{},"
done

echo "}"
