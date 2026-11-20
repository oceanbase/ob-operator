#!/usr/bin/env bash
set -euo pipefail

MONITOR_PASSWORD="$(cat /run/secrets/ob_monitor_password)"
export MONITOR_PASSWORD

cd /home/admin/obagent
# The image templates use loopback for a Kubernetes sidecar. In Compose the
# Observer has its own network endpoint and keeps the same service DNS name.
sed -i 's/@tcp(127\.0\.0\.1:/@tcp(observer:/g' conf/module_config/monitor_ob*.yaml
./replace_properties.sh
exec bin/ob_monagent -c conf/monagent.yaml
