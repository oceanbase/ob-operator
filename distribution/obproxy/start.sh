#!/bin/bash


if [ -z $APP_NAME ]; then
    echo "env variable APP_NAME is required"
    exit 1
fi

if [ -z $PROXYRO_PASSWORD_HASH ]; then
    PROXYRO_PASSWORD_HASH=`echo -n "$PROXYRO_PASSWORD" | sha1sum | awk '{print $1}'`
fi

if [ ! -z $CONFIG_URL ]; then
    echo "use config server"
    cd /home/admin/obproxy && /home/admin/obproxy/bin/obproxy -p 2883 -n ${APP_NAME} -o observer_sys_password=${PROXYRO_PASSWORD_HASH},obproxy_config_server_url="${CONFIG_URL}",prometheus_sync_interval=1,prometheus_listen_port=2884,enable_metadb_used=false,skip_proxy_sys_private_check=true,log_dir_size_threshold=10G,enable_proxy_scramble=true,enable_strict_kernel_release=false --nodaemon
elif [ ! -z $RS_LIST ]; then
    echo "use rslist"
    cd /home/admin/obproxy && /home/admin/obproxy/bin/obproxy -p 2883 -n ${APP_NAME} -c ${OB_CLUSTER} -r\"${RS_LIST}\" -o observer_sys_password=${PROXYRO_PASSWORD_HASH},prometheus_sync_interval=1,prometheus_listen_port=2884,enable_metadb_used=false,skip_proxy_sys_private_check=true,log_dir_size_threshold=10G,enable_proxy_scramble=true,enable_strict_kernel_release=false --nodaemon
else 
    echo "no config server or rs list"
    exit 1
fi
