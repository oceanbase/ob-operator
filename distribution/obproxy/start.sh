#!/bin/bash

if [ -z $APP_NAME ]; then
    echo "env variable APP_NAME is required"
    exit 1
fi

if [ -z $PROXYRO_PASSWORD_HASH ]; then
    PROXYRO_PASSWORD_HASH=$(echo -n "$PROXYRO_PASSWORD" | sha1sum | awk '{print $1}')
fi

if [ -z $PROSYSYS_PASSWORD_HASH ]; then
    PROXYSYS_PASSWORD_HASH=$(echo -n "$PROXYSYS_PASSWORD" | sha1sum | awk '{print $1}')
fi

opts="obproxy_sys_password=$PROXYSYS_PASSWORD_HASH"

function concat_opts {
    if [ -z "$1" ]; then
        echo $2
    elif [ -z "$2" ]; then
        echo $1
    else
        echo "$1,$2"
    fi
}

[ -z "$ODP_PROMETHEUS_SYNC_INTERVAL" ] && opts=$(concat_opts $opts "prometheus_sync_interval=1s")
[ -z "$ODP_ENABLE_METADB_USED" ] && opts=$(concat_opts $opts "enable_metadb_used=false")
[ -z "$ODP_SKIP_PROXY_SYS_PRIVATE_CHECK" ] && opts=$(concat_opts $opts "skip_proxy_sys_private_check=true")
[ -z "$ODP_LOG_DIR_SIZE_THRESHOLD" ] && opts=$(concat_opts $opts "log_dir_size_threshold=10G")
[ -z "$ODP_ENABLE_PROXY_SCRAMBLE" ] && opts=$(concat_opts $opts "enable_proxy_scramble=true")
[ -z "$ODP_ENABLE_STRICT_KERNEL_RELEASE" ] && opts=$(concat_opts $opts "enable_strict_kernel_release=false")

while IFS='=' read -r key value; do
    # If the key has prefix "ODP_" then add it to the opts
    if [[ $key == ODP_* ]]; then
        # Remove the prefix "ODP_" from the key and transform to lower case
        key=$(echo $key | sed 's/^ODP_//g' | tr '[:upper:]' '[:lower:]')
        opts=$(concat_opts $opts "$(printf "%s=\"%s\"" "$key" "$value")")
    fi
done < <(env)

echo "$opts"

if [ ! -z $CONFIG_URL ]; then
    echo "use config server"
    cd /home/admin/obproxy && /home/admin/obproxy/bin/obproxy -p 2883 -l 2884 -n ${APP_NAME} -o observer_sys_password=${PROXYRO_PASSWORD_HASH},obproxy_config_server_url="${CONFIG_URL}",$opts --nodaemon
elif [ ! -z $RS_LIST ]; then
    echo "use rslist"
    cd /home/admin/obproxy && /home/admin/obproxy/bin/obproxy -p 2883 -l 2884 -n ${APP_NAME} -c ${OB_CLUSTER} -r "${RS_LIST}" -o observer_sys_password=${PROXYRO_PASSWORD_HASH},$opts --nodaemon
else 
    echo "no config server or rs list"
    exit 1
fi
