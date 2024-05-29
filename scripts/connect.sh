#! /usr/bin/env bash

# This script is used to connect to an OBCluster in a Kubernetes cluster.

VERSION=0.1.0
NAMESPACE=default
OB_TENANT=""
USER=root
PASSWORD=""
PROXY_DEPLOY_NAME=""
CONNECT_PROXY=false

function print_help {
  echo "connect.sh - Connect to an OBCluster in a Kubernetes cluster"
  echo "Usage: connect.sh [options] <OBCluster>"
  echo "Options:"
  echo "  -h, --help            Display this help message and exit."
  echo "  -v, --version         Display version information and exit."
  echo "  -n                    Namespace of the OBCluster. Default is default."
  echo "  -A                    List all OBClusters in all namespaces."
  echo "  -l, --list            List OBClusters in the namespace."
  echo "  --list-tenants        List all tenants in all namespaces."
  echo "  -t <OBTenant>         OBTenant of the OBCluster. If not specified, the script will connect to the sys tenant."
  echo "  -u, --user <User>     User of the tenant. Default is root."
  echo "  -p, --password <Pwd>  Password of the user. If the user is not root, the password is required."
  echo "  --show-password       Show the password in the output."
  echo "  --proxy               Connect to the obproxy deployment with default name."
  echo "  --proxy-name <Name>   Connect to the obproxy deployment with the specified name. (--deploy-name in setup-obproxy.sh)"
}

# elif [[ $CONNECT == true ]]; then
#   if [[ $DEPLOY_EXIST != true ]]; then
#     echo "Error: The obproxy deployment \"$DEPLOY_NAME\" in namespace \"$NAMESPACE\" does not exist."
#     exit 1
#   fi

#   echo "Connecting to the obproxy deployment \"$DEPLOY_NAME\" in namespace \"$NAMESPACE\"..."
#   SVC_CLUSTER_IP=$(kubectl get service svc-$DEPLOY_NAME -n $NAMESPACE -o jsonpath='{.spec.clusterIP}')
#   mysql -h$SVC_CLUSTER_IP -P2883 -uroot -p$ROOT_PWD -A oceanbase

function print_version {
  echo "connect.sh - Connect to an OBCluster in a Kubernetes cluster"
  echo "Version: $VERSION"
}

if [[ $# -eq 0 ]]; then
  print_help
  exit 0
fi

function check_requirements {
  # Check whether mysql is installed
  if ! command -v mysql &> /dev/null; then
    echo "Error: mysqlclient is not installed."
    exit 1
  fi

  # Check whether kubectl is installed
  if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl is not installed."
    exit 1
  fi
}

check_requirements

while [[ $# -gt 0 ]]; do
  case $1 in
    -h|--help)
      print_help
      exit 0
      ;;
    -v|--version)
      print_version
      exit 0
      ;;
    -n)
      NAMESPACE=$2
      shift
      ;;
    -l|--list)
      kubectl get obclusters.oceanbase.oceanbase.com -n $NAMESPACE -o wide
      exit 0
      ;;
    -A)
      kubectl get obclusters.oceanbase.oceanbase.com --all-namespaces -o wide
      exit 0
      ;;
    -t|--tenant)
      OB_TENANT=$2
      shift
      ;;
    --list-tenants)
      kubectl get obtenants.oceanbase.oceanbase.com -A -o wide
      exit 0
      ;;
    -u|--user)
      USER=$2
      shift
      ;;
    -p|--password)
      PASSWORD=$2
      shift
      ;;
    --proxy)
      CONNECT_PROXY=true
      ;;
    --proxy-name)
      CONNECT_PROXY=true
      PROXY_DEPLOY_NAME=$2
      shift
      ;;
    --show-password)
      SHOW_PASSWORD=true
      ;;
    *)
      break
      ;;
  esac
  shift
done

if [[ $# -eq 0 ]]; then
  echo "Error: OBCluster is not specified."
  exit 1
fi

OB_CLUSTER=$1

if [[ $USER != "root" && -z $PASSWORD ]]; then
  echo "Error: Password is required for the user $USER."
  exit 1
fi

# Check whether the OBCluster exists
kubectl get obcluster $OB_CLUSTER -n $NAMESPACE &> /dev/null
if [[ $? -ne 0 ]]; then
  echo "Error: OBCluster \"$OB_CLUSTER\" does not exist in namespace \"$NAMESPACE\"."
  exit 1
fi

CONNECTING_HOST=""
CONNECTING_PORT=""

if [[ $CONNECT_PROXY == true ]]; then
  if [[ -z $PROXY_DEPLOY_NAME ]]; then
    PROXY_DEPLOY_NAME=obproxy-$OB_CLUSTER
  fi

  kubectl get deployment $PROXY_DEPLOY_NAME -n $NAMESPACE &> /dev/null
  if [[ $? -ne 0 ]]; then
    echo "Error: The obproxy deployment \"$PROXY_DEPLOY_NAME\" in namespace \"$NAMESPACE\" does not exist."
    exit 1
  fi

  CONNECTING_HOST=$(kubectl get service svc-$PROXY_DEPLOY_NAME -n $NAMESPACE -o jsonpath='{.spec.clusterIP}')
  CONNECTING_PORT=2883
else
  POD_IP=$(kubectl get pods -n $NAMESPACE -l ref-obcluster=$OB_CLUSTER -o jsonpath='{.items[0].status.podIP}')
  if [[ -z $POD_IP ]]; then
    echo "Error: No pod is found for OBCluster \"$OB_CLUSTER\" in namespace \"$NAMESPACE\"."
    exit 1
  fi
  CONNECTING_HOST=$POD_IP
  CONNECTING_PORT=2881
fi

ROOT_SECRET=$(kubectl get obcluster $OB_CLUSTER -n $NAMESPACE -o jsonpath='{.spec.userSecrets.root}')
export ROOT_PWD=$(kubectl get secret $ROOT_SECRET -n $NAMESPACE -o jsonpath='{.data.password}' | base64 -d)

if [[ $SHOW_PASSWORD == true ]]; then
  echo "Root Password: $ROOT_PWD"
  exit 0
fi

echo "Connecting to OBCluster \"$OB_CLUSTER\" in namespace \"$NAMESPACE\"..."
echo "Host IP: $CONNECTING_HOST"
if [[ -n $OB_TENANT ]]; then
  echo "OBTenant: $OB_TENANT"
fi
echo "User: $USER"
if [[ $CONNECT_PROXY == true ]]; then
  echo "Proxy Deployment: $PROXY_DEPLOY_NAME"
fi

if [[ -z $OB_TENANT ]]; then
  echo ""
  if [[ $USER == "root" ]]; then
    mysql -h$CONNECTING_HOST -uroot -p$ROOT_PWD -P$CONNECTING_PORT -A oceanbase
  else
    mysql -h$CONNECTING_HOST -u$USER -p$PASSWORD -P$CONNECTING_PORT -A oceanbase
  fi
else
  kubectl get obtenants.oceanbase.oceanbase.com $OB_TENANT -n $NAMESPACE &> /dev/null
  if [[ $? -ne 0 ]]; then
    echo "Error: OBTenant \"$OB_TENANT\" does not exist in namespace \"$NAMESPACE\"."
    exit 1
  fi

  TENANT_CONNECT_HOST=$CONNECTING_HOST
  if [[ $CONNECT_PROXY == false ]]; then
    TENANT_ID=$(kubectl get obtenants.oceanbase.oceanbase.com $OB_TENANT -n $NAMESPACE -o jsonpath='{.status.tenantRecordInfo.tenantID}')
    UNIT_POD_IP=$(mysql -h$CONNECTING_HOST -P$CONNECTING_PORT -uroot -p$ROOT_PWD -BN -e "SELECT svr_ip FROM oceanbase.DBA_OB_UNITS where tenant_id = $TENANT_ID;")
    TENANT_CONNECT_HOST=$UNIT_POD_IP
  fi
  TENANT_NAME=$(kubectl get obtenants.oceanbase.oceanbase.com $OB_TENANT -n $NAMESPACE -o jsonpath='{.spec.tenantName}')

  echo "Tenant Name: $TENANT_NAME"
  echo ""
  if [[ $USER == "root" ]]; then
    TENANT_ROOT_SECRET=$(kubectl get obtenants.oceanbase.oceanbase.com $OB_TENANT -n $NAMESPACE -o jsonpath='{.spec.credentials.root}')
    export TENANT_ROOT_PWD=$(kubectl get secret $TENANT_ROOT_SECRET -n $NAMESPACE -o jsonpath='{.data.password}' | base64 -d)
    mysql -h$TENANT_CONNECT_HOST -uroot@$TENANT_NAME -p$TENANT_ROOT_PWD -P$CONNECTING_PORT -A oceanbase
  else 
    mysql -h$TENANT_CONNECT_HOST -u$USER@$TENANT_NAME -p$PASSWORD -P$CONNECTING_PORT -A oceanbase
  fi
fi
