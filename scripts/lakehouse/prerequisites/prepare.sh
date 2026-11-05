#!/bin/sh
set -eu
# K3s kubectl otherwise defaults to its node-local kubeconfig, even inside a Pod.
umask 077
export KUBECONFIG=/tmp/prerequisites-kubeconfig
cat > "$KUBECONFIG" <<'EOF'
apiVersion: v1
kind: Config
clusters:
  - name: in-cluster
    cluster:
      server: https://kubernetes.default.svc
      certificate-authority: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
users:
  - name: service-account
    user:
      tokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
contexts:
  - name: in-cluster
    context:
      cluster: in-cluster
      user: service-account
current-context: in-cluster
EOF
k() { /bin/k3s kubectl --request-timeout=30s "$@"; }
cleanup() {
  k delete clusterrolebinding "$BOOTSTRAP_BINDING" --ignore-not-found >/dev/null 2>&1 || true
}
trap cleanup EXIT
trap 'exit 1' INT TERM
wait_operator() {
  ns="$1"
  k -n "$ns" rollout status deployment/oceanbase-controller-manager --timeout=300s
  for attempt in $(seq 1 120); do
    endpoints="$(k -n "$ns" get endpoints oceanbase-webhook-service -o jsonpath='{.subsets[*].addresses[*].ip}')"
    ca="$(k get validatingwebhookconfiguration oceanbase-validating-webhook-configuration -o jsonpath='{.webhooks[0].clientConfig.caBundle}')"
    mutating_ca="$(k get mutatingwebhookconfiguration oceanbase-mutating-webhook-configuration -o jsonpath='{.webhooks[0].clientConfig.caBundle}')"
    if [ -n "$endpoints" ] && [ -n "$ca" ] && [ -n "$mutating_ca" ]; then
      echo "Operator webhook ready in namespace $ns"
      return
    fi
    sleep 2
  done
  echo 'Operator webhook endpoints or CA bundle did not become ready' >&2
  exit 1
}
# Reject a different installed controller before installing any infrastructure.
controllers="$(k get deployments -A --field-selector metadata.name=oceanbase-controller-manager -o jsonpath='{range .items[*]}{.metadata.namespace}{" "}{.spec.template.spec.containers[?(@.name=="manager")].image}{"\n"}{end}')"
if [ -n "$controllers" ]; then
  count="$(printf '%s\n' "$controllers" | wc -l | tr -d ' ')"
  [ "$count" = 1 ] || { echo 'Multiple OceanBase Operators exist; refusing to install another.' >&2; exit 1; }
  set -- $controllers
  namespace="$1"; installed_image="$2"
  matched=false
  while IFS= read -r allowed; do
    if [ "$installed_image" = "$allowed" ]; then matched=true; fi
  done < /payload/operator-compatible-images.txt
  [ "$matched" = true ] || { echo "Existing Operator image $installed_image is not in this release's compatible image list; upgrade it explicitly before retrying." >&2; exit 1; }
  echo "Reusing compatible Operator in $namespace"
  wait_operator "$namespace"
  exit 0
fi
if k -n cert-manager get deployment cert-manager >/dev/null 2>&1; then
  echo 'Reusing existing cert-manager'
else
  controllers="$(k get deployments -A -o jsonpath='{range .items[*]}{.spec.template.spec.containers[*].image}{"\n"}{end}')"
  case "$controllers" in *cert-manager-controller*) echo 'cert-manager exists outside the standard namespace; refusing a duplicate installation.' >&2; exit 1;; esac
  echo 'Installing bundled cert-manager'
  k apply --server-side --field-manager=lakehouse-prerequisites -f /payload/cert-manager.yaml
fi
for name in cert-manager cert-manager-cainjector cert-manager-webhook; do
  k -n cert-manager rollout status deployment/"$name" --timeout=300s
done
# cert-manager Deployment readiness may precede API-server webhook CA injection.
for attempt in $(seq 1 90); do
  ca="$(k get validatingwebhookconfiguration cert-manager-webhook -o jsonpath='{.webhooks[0].clientConfig.caBundle}')"
  [ -n "$ca" ] && break
  sleep 2
done
[ -n "$ca" ] || { echo 'cert-manager webhook CA is not ready' >&2; exit 1; }
k create namespace oceanbase-system --dry-run=client -o yaml | k apply --server-side --field-manager=lakehouse-prerequisites -f -
# Do not take over resources owned by a separately installed Helm release.
owned="$(k get -f /payload/operator.yaml --ignore-not-found -o jsonpath='{range .items[*]}{.metadata.annotations.meta\.helm\.sh/release-name}{end}')"
[ -z "$owned" ] || { echo 'Existing Operator resources are owned by another Helm release; refusing takeover.' >&2; exit 1; }
echo 'Installing matched OceanBase Operator'
# Retry API webhook readiness races; SSA never forces another field manager's fields.
ready=false
for attempt in $(seq 1 30); do
  if k apply --server-side --field-manager=lakehouse-prerequisites -f /payload/operator.yaml; then ready=true; break; fi
  sleep 3
done
[ "$ready" = true ] || exit 1
k -n oceanbase-system wait --for=condition=Ready certificate/oceanbase-serving-cert --timeout=300s
wait_operator oceanbase-system
echo 'Cluster prerequisites ready; Helm will now install Lakehouse and Dashboard.'
