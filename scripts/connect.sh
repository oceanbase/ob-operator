#! /usr/bin/env bash

# This script is used to connect to an OBCluster in a Kubernetes cluster.

VERSION=0.1.0
NAMESPACE=default
TENANT=sys
USER=root
PASSWORD=""

function print_help {
  echo "connect.sh - Connect to an OBCluster in a Kubernetes cluster"
  echo "Usage: connect.sh [options] <OBCluster>"
  echo "Options:"
  echo "  -h, --help            Display this help message and exit."
  echo "  -v, --version         Display version information and exit."
  echo "  -n                    Namespace of the OBCluster. Default is default."
  echo "  -l, --list            List OBClusters in the namespace."
  echo "  -A                    List all OBClusters in all namespaces."
  echo "  -t, --tenant          OBTenant of the OBCluster. Default is sys."
  echo "  --list-tenants        List all tenants in all namespaces."
  echo "  -u, --user            User of the tenant. Default is root."
  echo "  -p, --password        Password of the user. If the user is not root, the password is required."
}

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
      TENANT=$2
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

POD_IP=$(kubectl get pods -n $NAMESPACE -l ref-obcluster=$OB_CLUSTER -o jsonpath='{.items[0].status.podIP}')
if [[ -z $POD_IP ]]; then
  echo "Error: No pod is found for OBCluster \"$OB_CLUSTER\" in namespace \"$NAMESPACE\"."
  exit 1
fi

ROOT_SECRET=$(kubectl get obcluster $OB_CLUSTER -n $NAMESPACE -o jsonpath='{.spec.userSecrets.root}')
export ROOT_PWD=$(kubectl get secret $ROOT_SECRET -n $NAMESPACE -o jsonpath='{.data.password}' | base64 -d)

echo "Connecting to OBCluster \"$OB_CLUSTER\" in namespace \"$NAMESPACE\"..."
echo "OBTenant: $TENANT"
echo "User: $USER"

if [[ $TENANT == "sys" ]]; then
  echo ""
  if [[ $USER == "root" ]]; then
    mysql -h$POD_IP -uroot -p$ROOT_PWD -P2881 -A oceanbase
  else
    mysql -h$POD_IP -u$USER -p$PASSWORD -P2881 -A oceanbase
  fi
else
  kubectl get obtenants.oceanbase.oceanbase.com $TENANT -n $NAMESPACE &> /dev/null
  if [[ $? -ne 0 ]]; then
    echo "Error: OBTenant \"$TENANT\" does not exist in namespace \"$NAMESPACE\"."
    exit 1
  fi

  TENANT_ID=$(kubectl get obtenants.oceanbase.oceanbase.com $TENANT -n $NAMESPACE -o jsonpath='{.status.tenantRecordInfo.tenantID}')
  UNIT_POD_IP=$(mysql -h$POD_IP -P2881 -uroot -p$ROOT_PWD -BN -e "SELECT svr_ip FROM oceanbase.DBA_OB_UNITS where tenant_id = $TENANT_ID;")
  TENANT_NAME=$(kubectl get obtenants.oceanbase.oceanbase.com $TENANT -n $NAMESPACE -o jsonpath='{.spec.tenantName}')
  echo "Tenant Name: $TENANT_NAME"
  echo ""
  if [[ $USER == "root" ]]; then
    TENANT_ROOT_SECRET=$(kubectl get obtenants.oceanbase.oceanbase.com $TENANT -n $NAMESPACE -o jsonpath='{.spec.credentials.root}')
    export TENANT_ROOT_PWD=$(kubectl get secret $TENANT_ROOT_SECRET -n $NAMESPACE -o jsonpath='{.data.password}' | base64 -d)
    mysql -h$UNIT_POD_IP -uroot@$TENANT_NAME -p$TENANT_ROOT_PWD -P2881 -A oceanbase
  else 
    mysql -h$UNIT_POD_IP -u$USER@$TENANT_NAME -p$PASSWORD -P2881 -A oceanbase
  fi
fi
