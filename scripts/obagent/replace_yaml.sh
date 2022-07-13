#!/bin/bash
### 关闭这个功能
http_basic_auth_user='admin'
http_basic_auth_password='root'
pprof_basic_auth_user='admin'
pprof_basic_auth_password='root'


monitor_user=$monitor_user
monitor_password=$monitor_password

sql_port='2881'
rpc_port='2882'

ob_install_path='ori_path'
host_ip=$host_ip

cluster_name=$cluster_name
cluster_id=$cluster_id

zone_name=$zone_name
ob_monitor_status='active'
ob_log_monitor_status='inactive'
host_monitor_status='inactive'
alertmanager_address='temp'
disable_http_basic_auth='true'
disable_pprof_basic_auth='true'

### 配置 monagent_basic_auth.yaml
sed -i "s/{http_basic_auth_user}/${http_basic_auth_user}/g" ./conf/config_properties/monagent_basic_auth.yaml
sed -i "s/{http_basic_auth_password}/${http_basic_auth_password}/g" ./conf/config_properties/monagent_basic_auth.yaml
sed -i "s/{pprof_basic_auth_user}/${pprof_basic_auth_user}/g" ./conf/config_properties/monagent_basic_auth.yaml
sed -i "s/{pprof_basic_auth_password}/${pprof_basic_auth_password}/g" ./conf/config_properties/monagent_basic_auth.yaml

### 配置 monagent_pipeline.yaml
sed -i "s/{monitor_user}/${monitor_user}/g" ./conf/config_properties/monagent_pipeline.yaml
sed -i "s/{monitor_password}/${monitor_password}/g" ./conf/config_properties/monagent_pipeline.yaml
sed -i "s/{sql_port}/${sql_port}/g" ./conf/config_properties/monagent_pipeline.yaml
sed -i "s/{rpc_port}/${rpc_port}/g" ./conf/config_properties/monagent_pipeline.yaml
sed -i "s/{ob_install_path}/${ob_install_path}/g" ./conf/config_properties/monagent_pipeline.yaml
sed -i "s/{host_ip}/${host_ip}/g" ./conf/config_properties/monagent_pipeline.yaml

sed -i "s/{cluster_name}/${cluster_name}/g" ./conf/config_properties/monagent_pipeline.yaml
sed -i "s/{cluster_id}/${cluster_id}/g" ./conf/config_properties/monagent_pipeline.yaml
sed -i "s/{zone_name}/${zone_name}/g" ./conf/config_properties/monagent_pipeline.yaml

sed -i "s/{ob_monitor_status}/${ob_monitor_status}/g" ./conf/config_properties/monagent_pipeline.yaml
sed -i "s/{ob_log_monitor_status}/${ob_log_monitor_status}/g" ./conf/config_properties/monagent_pipeline.yaml
sed -i "s/{host_monitor_status}/${host_monitor_status}/g" ./conf/config_properties/monagent_pipeline.yaml
sed -i "s/{alertmanager_address}/${alertmanager_address}/g" ./conf/config_properties/monagent_pipeline.yaml

###  替换 ip
sed -i 's/127.0.0.1/${monagent.host.ip}/g' ./conf/module_config/monitor_ob.yaml
sed -i "s/{disable_http_basic_auth}/${disable_http_basic_auth}/g" ./conf/module_config/monagent_basic_auth.yaml
sed -i "s/{disable_pprof_basic_auth}/${disable_pprof_basic_auth}/g" ./conf/module_config/monagent_basic_auth.yaml