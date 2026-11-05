#!/usr/bin/env bash
set -euo pipefail

MONITOR_PASSWORD="$(cat /run/secrets/ob_monitor_password)"
export MONITOR_PASSWORD

cd /home/admin/obagent
./replace_properties.sh
exec bin/ob_monagent -c conf/monagent.yaml
