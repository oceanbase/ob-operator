/home/admin/obagent/bin/ob_agentctl config -u \
agent.http.basic.auth.metricAuthEnabled=false,\
monagent.ob.monitor.user=${MONITOR_USER},\
monagent.ob.monitor.password=${MONITOR_PASSWORD},\
monagent.host.ip=`hostname -i`,\
monagent.cluster.id=${CLUSTER_ID},\
monagent.ob.cluster.name=${CLUSTER_NAME},\
monagent.ob.cluster.id=${CLUSTER_ID},\
monagent.ob.zone.name=${ZONE_NAME},\
monagent.pipeline.ob.status=active,\
monagent.pipeline.node.status=inactive,\
monagent.second.metric.cache.update.interval=5s,\
ocp.agent.monitor.http.port=8088
