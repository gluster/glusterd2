<!--
This issue template is meant mainly for bug reports.

You can still use an issue to ask for help. In this case
you should remove the template contents or use it as a guideline
for providing information for debugging.
-->


### Observed behavior


### Expected/desired behavior


### Details on how to reproduce (minimal and precise)


### Information about the environment:

- Glusterd2 version used (e.g. v4.1.0 or master): 
- Operating system used: 
- Glusterd2 compiled from sources, as a package (rpm/deb), or container: 
- Using External ETCD: (yes/no, if yes ETCD version):
- If container, which container image: 
- Using kubernetes, openshift, or direct install: 
- If kubernetes/openshift, is gluster running inside kubernetes/openshift or outside: 


### Other useful information

- glusterd2 config files from all nodes (default /etc/glusterd2/glusterd2.toml)
- glusterd2 log files from all nodes (default /var/log/glusterd2/glusterd2.log)
- ETCD configuration
- Contents of `uuid.toml` from all nodes (default /var/lib/glusterd2/uuid.toml)
- Output of `statedump` from any one of the node

### Useful commands

- To get glusterd2 version
    ```
    glusterd2 --version
    ```
- To get ETCD version
    ```
    etcd --version
    ```
- To get output of statedump
    ```
    curl http://glusterd2-IP:glusterd2-Port/statedump
    ```