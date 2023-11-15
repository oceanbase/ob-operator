 #!/bin/bash

if [[ "$DATABASE_TYPE" == "mysql" ]]; then
    echo 'use mysql'
    sed -e "s/\${SVC_ADDRESS}/${SVC_ADDRESS}/" -e "s/\${SVC_PORT}/${SVC_PORT}/" -e "s/\${DATABASE_TYPE}/${DATABASE_TYPE}/" -e "s/\${META_USER}/${META_USER}/" -e "s/\${META_PASSWORD}/${META_PASSWORD}/" -e "s/\${META_HOST}/${META_HOST}/" -e "s/\${META_PORT}/${META_PORT}/" -e "s/\${META_DATABASE}/${META_DATABASE}/" ./conf/config.yaml.template | grep -v '#CONFIG_FOR_SQLITE' > ./conf/config.yaml
elif [[ "$DATABASE_TYPE" == "sqlite3" ]]; then
    echo 'use sqlite3'
    sed -e "s/\${SVC_ADDRESS}/${SVC_ADDRESS}/" -e "s/\${SVC_PORT}/${SVC_PORT}/" -e "s/\${DATABASE_TYPE}/${DATABASE_TYPE}/" ./conf/config.yaml.template | grep -v '#CONFIG_FOR_MYSQL' > ./conf/config.yaml
else
    echo "database type not supported"
    exit 1
fi
 ./bin/ob-configserver -c ./conf/config.yaml
