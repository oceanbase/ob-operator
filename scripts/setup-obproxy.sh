#! /usr/bin/env bash

# This script is used to setup obproxy for an OBCluster handyly for quick testing.
# It will install obproxy and configure it with a simple configuration file.

VERSION=0.1.0
NAMESPACE=default
DEPLOY_NAME=""
PROXY_VERSION="4.2.1.0-11"
DESTROY=false
SVC_TYPE=ClusterIP
DISPLAY_INFO=false
LIST=false
LIST_ALL=false
ENV_VARS=()
CONFIG_MAP=""

function print_help {
  echo "setup-obproxy.sh - Set up obproxy for an OBCluster in a Kubernetes cluster"
  echo "Usage: setup-obproxy.sh [options] <OBCluster>"
  echo "Options:"
  echo "  -h, --help                Display this help message and exit."
  echo "  -v, --version             Display version information and exit."
  echo "  -n <Namespace>            Namespace of the OBCluster. Default is default."
  echo "  -l, --list                List OBProxy deployments in current namespace by default, add -A for all namespace."
  echo "  -A, --all                 List OBProxy deployments in all namespaces, paired with -l."
  echo "  -i, --info                Display the obproxy deployment information."
  echo "  -d, --deploy-name <Name>  Name of the obproxy deployment. Default is obproxy-<OBCluster>."
  echo "  -r, --replicas <Number>   Number of replicas of the obproxy deployment. Default is 2."
  echo "  --destroy                 Destroy the obproxy deployment."
  echo "  -e, --env <key=value>     Environment variable of the obproxy deployment."
  echo "  --cm <ConfigMap>          ConfigMap of the obproxy deployment."
  echo "  --cpu <CPU>               CPU limit of the obproxy deployment. Default is 1."
  echo "  --memory <Memory>         Memory limit of the obproxy deployment. Default is 2Gi."
  echo "  --svc-type <Type>         Service type of the obproxy deployment. Default is ClusterIP. Valid values are ClusterIP, NodePort, LoadBalancer."
  echo "  --proxy-version <Version> Version of the obproxy image. Default is 4.2.1.0-11."
  echo "  --proxy-image <Image>     Image of the obproxy. Default is oceanbase/obproxy-ce:4.2.1.0-11."
}

function print_version {
  echo "setup-obproxy.sh - Set up obproxy for an OBCluster in a Kubernetes cluster"
  echo "Version: $VERSION"
}

if [[ $# -eq 0 ]]; then
  print_help
  exit 0
fi

# Parse options
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
    -d|--deploy-name)
      DEPLOY_NAME=$2
      shift
      ;;
    --proxy-version)
      PROXY_VERSION=$2
      shift
      ;;
    --destroy)
      DESTROY=true
      ;;
    -r|--replicas)
      REPLICAS=$2
      shift
      ;;
    -e|--env)
      if [[ $CONFIG_MAP != "" ]]; then
        echo "Error: Environment variables and ConfigMap cannot be set at the same time."
        exit 1
      fi
      echo "-e" "$2"
      ENV_VARS+=("$2")
      shift
      ;;
    --cm)
      if [[ $2 == "" ]]; then
        echo "Error: ConfigMap name is empty."
        exit 1
      fi
      # if length of ENV_VARS > 0
      if [[ ${#ENV_VARS[@]} -gt 0 ]]; then
        echo "Error: Environment variables and ConfigMap cannot be set at the same time."
        exit 1
      fi
      CONFIG_MAP=$2
      shift
      ;;
    --svc-type)
      SVC_TYPE=$2
      if [[ $SVC_TYPE != "ClusterIP" && $SVC_TYPE != "NodePort" && $SVC_TYPE != "LoadBalancer" ]]; then
        echo "Error: Invalid service type \"$SVC_TYPE\"."
        exit 1
      fi
      shift
      ;;
    -i|--info)
      DISPLAY_INFO=true
      ;;
    --cpu)
      CPU_LIMIT=$2
      shift
      ;;
    --memory)
      MEMORY_LIMIT=$2
      shift
      ;;
    -l|--list)
      LIST=true
      ;;
    -A|--all)
      LIST_ALL=true
      ;;
    --proxy-image)
      PROXY_IMAGE=$2
      shift
      ;;
    *)
      break
      ;;
  esac
  shift
done

if [[ $LIST == true ]]; then
  echo -e "\nOBProxy Deployments: \n"
  if [[ $LIST_ALL == true ]]; then
    kubectl get deployment -A -l obproxy.oceanbase.com/obproxy-from-setup -L obproxy.oceanbase.com/for-obcluster -o wide
  else
    kubectl get deployment -n $NAMESPACE -l obproxy.oceanbase.com/obproxy-from-setup -L obproxy.oceanbase.com/for-obcluster -o wide
  fi
  exit 0
fi

if [[ $LIST_ALL == true ]]; then
  echo "Error: -A|--all must be paired with -l|--list."
  exit 1
fi

# Check if the OBCluster is specified
if [[ $# -eq 0 ]]; then
  echo "Error: OBCluster is not specified."
  exit 1
fi

# OBCluster name
OB_CLUSTER=$1

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

if [[ $CONFIG_MAP == "" ]]; then
  # Create ConfigMap
  CONFIG_MAP="cm-obproxy-$OB_CLUSTER"
  # ENV_VARS is an key=value array
  env_list="--from-literal=DEPLOYED_TIME=$(date +'%Y-%m-%d.%H:%M:%S')"
  for env in "${ENV_VARS[@]}"; do
    env_list="$env_list --from-literal=$env"
  done
  kubectl create configmap $CONFIG_MAP -n $NAMESPACE $env_list
fi

# Check whether the OBCluster exists
kubectl get obcluster $OB_CLUSTER -n $NAMESPACE &> /dev/null
if [[ $? -ne 0 ]]; then
  echo "Error: OBCluster \"$OB_CLUSTER\" does not exist in namespace \"$NAMESPACE\"."
  exit 1
fi

CLUSTER_NAME=$(kubectl get obcluster $OB_CLUSTER -n $NAMESPACE -o jsonpath='{.spec.clusterName}')
PROXYRO_SECRET=$(kubectl get obcluster $OB_CLUSTER -n $NAMESPACE -o jsonpath='{.spec.userSecrets.proxyro}')
ROOT_SECRET=$(kubectl get obcluster $OB_CLUSTER -n $NAMESPACE -o jsonpath='{.spec.userSecrets.root}')
ROOT_PWD=$(kubectl get secret $ROOT_SECRET -n $NAMESPACE -o jsonpath='{.data.password}' | base64 -d)
POD_IP=$(kubectl get pods -n $NAMESPACE -l ref-obcluster=$OB_CLUSTER -o jsonpath='{.items[0].status.podIP}')
RS_LIST=$(mysql -h$POD_IP -P2881 -uroot -p$ROOT_PWD -BN -e "SELECT GROUP_CONCAT(CONCAT(SVR_IP, ':', SQL_PORT) SEPARATOR ';') AS RS_LIST FROM oceanbase.DBA_OB_SERVERS;")

if [[ -z $DEPLOY_NAME ]]; then
  DEPLOY_NAME="obproxy-$OB_CLUSTER"
fi

function display_info {
  echo "[OBProxy Deployment]"
  kubectl get deployment $DEPLOY_NAME -n $NAMESPACE

  echo -e "\n[OBProxy Pods]"
  kubectl get pods -n $NAMESPACE -l app=app-$DEPLOY_NAME -o wide

  echo -e "\n[OBProxy Service]"
  kubectl get service svc-$DEPLOY_NAME -n $NAMESPACE
}

function create_deployment {

cat <<EOF | kubectl create -f -
apiVersion: v1
kind: Service
metadata:
  name: svc-$DEPLOY_NAME
  namespace: $NAMESPACE
spec:
  type: $SVC_TYPE
  selector:
    app: app-$DEPLOY_NAME
  ports:
    - name: "sql"
      port: 2883
      targetPort: 2883
    - name: "prometheus"
      port: 2884
      targetPort: 2884
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $DEPLOY_NAME
  namespace: $NAMESPACE
  labels:
    obproxy.oceanbase.com/obproxy-from-setup: "$DEPLOY_NAME"
    obproxy.oceanbase.com/for-obcluster: "$OB_CLUSTER"
    obproxy.oceanbase.com/for-namespace: "$NAMESPACE"
    obproxy.oceanbase.com/with-config-map: "$CONFIG_MAP"
spec:
  selector:
    matchLabels:
      app: app-$DEPLOY_NAME
  replicas: ${REPLICAS:-2}
  template:
    metadata:
      labels:
        app: app-$DEPLOY_NAME
    spec:
      containers:
        - name: obproxy
          image: ${PROXY_IMAGE:-"oceanbase/obproxy-ce:"$PROXY_VERSION}
          ports:
            - containerPort: 2883
              name: "sql"
            - containerPort: 2884
              name: "prometheus"
          envFrom:
            - configMapRef:
                name: $CONFIG_MAP
          env:
            - name: APP_NAME
              value: $DEPLOY_NAME
            - name: OB_CLUSTER
              value: $CLUSTER_NAME
            - name: RS_LIST
              value: $RS_LIST
            - name: PROXYRO_PASSWORD
              valueFrom: 
                secretKeyRef:
                  name: $PROXYRO_SECRET
                  key: password
          resources:
            limits:
              memory: ${MEMORY_LIMIT:-2Gi}
              cpu: "${CPU_LIMIT:-1}"
            requests: 
              memory: 200Mi
              cpu: 200m
EOF

  echo "Waiting for the obproxy deployment to be ready..."

  # Wait for the obproxy deployment to be ready
  kubectl wait --for=condition=available --timeout=5m deployment/$DEPLOY_NAME -n $NAMESPACE
}

# Check whether the obproxy deployment already exists
kubectl get deployment $DEPLOY_NAME -n $NAMESPACE &> /dev/null
DEPLOY_EXIST=$(if [[ $? -eq 0 ]]; then echo true; else echo false; fi)

if [[ $DISPLAY_INFO == true ]]; then
  if [[ $DEPLOY_EXIST != true ]]; then
    echo "Error: The obproxy deployment \"$DEPLOY_NAME\" in namespace \"$NAMESPACE\" does not exist."
    exit 1
  fi
  display_info
elif [[ $DESTROY == true ]]; then
  if [[ $DEPLOY_EXIST != true ]]; then
    echo "Error: The obproxy deployment \"$DEPLOY_NAME\" in namespace \"$NAMESPACE\" does not exist."
    exit 1
  else
    echo "Destroying the obproxy deployment \"$DEPLOY_NAME\" in namespace \"$NAMESPACE\"..."
  fi

  kubectl delete deployment $DEPLOY_NAME -n $NAMESPACE
  kubectl delete service svc-$DEPLOY_NAME -n $NAMESPACE

  echo "OBProxy has been destroyed successfully."
else
  if [[ $DEPLOY_EXIST == true ]]; then
    # Get svc type of the existing service
    SVC_TYPE_EXIST=$(kubectl get service svc-$DEPLOY_NAME -n $NAMESPACE -o jsonpath='{.spec.type}')
    if [[ $SVC_TYPE_EXIST != "$SVC_TYPE" ]]; then
      if [[ $SVC_TYPE_EXIST == "ClusterIP" ]]; then 
        # Patch the service type
        kubectl patch service svc-$DEPLOY_NAME -n $NAMESPACE --type='json' -p='[{"op": "replace", "path": "/spec/type", "value": "'$SVC_TYPE'"}]'
        echo "Update the service type of the obproxy deployment \"$DEPLOY_NAME\" in namespace \"$NAMESPACE\" to \"$SVC_TYPE\"."
        exit 0
      else
        echo "Error: Can not update the service type of the obproxy deployment \"$DEPLOY_NAME\" in namespace \"$NAMESPACE\" from \"$SVC_TYPE_EXIST\" to \"$SVC_TYPE\"."
        exit 1
      fi
    fi
    echo "Error: The obproxy deployment \"$DEPLOY_NAME\" in namespace \"$NAMESPACE\" already exists."
    exit 1
  fi

  create_deployment

  display_info

  echo ""
  echo "OBProxy has been set up successfully."
fi