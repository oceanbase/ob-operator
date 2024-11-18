# Deploy OpenStack with OceanBase on Kubernetes

## Overview
This folder contains configuration files to deploy OceanBase and OpenStack on Kubernetes.
* A script for setting up OceanBase cluster and create tenant and deploy obproxy and finally provide a connection to use.
* OpenStack configuration files are located in the openstack directory.

## Deploy steps

### Deploy OceanBase
The script oceanbase.sh provides a simple way to deploy OceanBase ready for use as OpenStack's metadb, it requires the storage class `general` already created in the K8s cluster, to deploy OceanBase you can simply run the following command
```
bash oceanbase.sh
```

### Deploy OpenStack
Once OceanBase is set up, deploying OpenStack is straightforward. Override the necessary variables using the files under the openstack directory. The files are based on OpenStack version 2024.1.
Follow the official OpenStack Helm [document](https://docs.openstack.org/openstack-helm/latest/readme.html).

