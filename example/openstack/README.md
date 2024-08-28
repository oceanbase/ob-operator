# Deploy OpenStack with OceanBase on Kubernetes

## Overview
This folder contains configuration files to deploy OceanBase and OpenStack on Kubernetes.
* OceanBase configuration files are located in the oceanbase directory.
* OpenStack configuration files are located in the openstack directory.

## Deploy steps

### Deploy OceanBase
1. Deploy cert-manager
Deploy the cert-manager using the following command. Ensure all pods are running before proceeding to the next step:
```
kubectl apply -f https://raw.githubusercontent.com/oceanbase/ob-operator/2.2.2_release/deploy/cert-manager.yaml

```
2. deploy ob-operator
Deploy the ob-operator using the command below. Wait until all pods are running:
```
kubectl apply -f https://raw.githubusercontent.com/oceanbase/ob-operator/2.2.2_release/deploy/cert-manager.yaml
```
3. deploy obcluster
Deploy the obcluster using the following command:
```
kubectl apply -f oceanbase/obcluster.yaml
```
Wait until the obcluster status changes to `running`. You can check the status using:
```
kubectl get obcluster openstack -n openstack -o wide 
```

4. deploy obtenant
Deploy the obtenant using the command below:
```
kubectl apply -f oceanbase/obtenant.yaml
```
Wait until the obtenant status changes to `running`. You can verify this using:
```
kubectl get obtenant openstack -n openstack -o wide 
```

5. deploy obproxy
A script is provided to set up obproxy. Download the script with the following command:
```
wget https://raw.githubusercontent.com/oceanbase/ob-operator/master/scripts/setup-obproxy.sh
```
Run the script to set up obproxy:
```
bash setup-obproxy.sh -n openstack --proxy-version 4.2.3.0-3 --env ODP_MYSQL_VERSION=8.0.30 --env ODP_PROXY_TENANT_NAME=openstack -d openstack  openstack
```

6. Configure Tenant Variables
Connect to the openstack tenant using the command below (${ip} need to be replaced with the obproxy ip, root@openstack is the root user of openstack tenant, and the default password is password):
```
mysql -h${ip} -P2883 -uroot@openstack -p
```
Set the necessary variables
```
set global version='8.0.30';
set global autocommit=0;
```


### Deploy OpenStack
Once OceanBase is set up, deploying OpenStack is straightforward. Override the necessary variables using the files under the openstack directory. The files are based on OpenStack version 2024.1.
Follow the official OpenStack Helm [document](https://docs.openstack.org/openstack-helm/latest/readme.html).

