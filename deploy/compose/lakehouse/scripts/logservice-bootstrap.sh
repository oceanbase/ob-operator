#!/usr/bin/env bash
set -euo pipefail

access_id="$(cat /run/secrets/minio_access_id)"
access_key="$(cat /run/secrets/minio_access_key)"
object_store_url="${LOGSERVICE_BUCKET_URL}&access_id=${access_id}&access_key=${access_key}"
nodes="REGION ${REGION_NAME} AZ ${ZONE_NAME_1} SERVER ${LOGSERVICE_IP_1}:50051, REGION ${REGION_NAME} AZ ${ZONE_NAME_2} SERVER ${LOGSERVICE_IP_2}:50051, REGION ${REGION_NAME} AZ ${ZONE_NAME_3} SERVER ${LOGSERVICE_IP_3}:50051"

for attempt in $(seq 1 60); do
  if access_point=$(/home/admin/oblogservice/bin/ls_ctrl --host "${LOGSERVICE_IP_1}:50052" show ap 2>/dev/null); then
    if grep -Eq "\"cluster_id\":[[:space:]]*${LOGSERVICE_CLUSTER_ID}([,}])" <<<"${access_point}"; then
      echo "LogService is already bootstrapped"
      exit 0
    fi
    echo "LogService endpoint belongs to a different cluster" >&2
    exit 1
  fi
  if output=$(/home/admin/oblogservice/bin/ls_ctrl \
    --host "${LOGSERVICE_IP_1}:50052" \
    bootstrap \
    --object-store-url "${object_store_url}" \
    "${nodes}" 2>&1); then
    echo "LogService bootstrap completed"
    exit 0
  fi
  if grep -qi "already bootstrap" <<<"${output}"; then
    # Only show-ap can confirm that the existing cluster has the expected ID.
    # Retry it instead of treating an unverified "already bootstrap" as success.
    sleep 5
    continue
  fi
  safe_output="${output//${access_id}/***ID***}"
  safe_output="${safe_output//${access_key}/***KEY***}"
  echo "LogService bootstrap attempt ${attempt}/60 failed: ${safe_output}; retrying" >&2
  sleep 5
done

echo "LogService bootstrap failed after 60 attempts" >&2
exit 1
