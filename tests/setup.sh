# setup variables for namespace
export NAMESPACE_PREFIX=${NAMESPACE_PREFIX:-oceanbase}
# setup variables for images
export OB_IMAGE=${OB_IMAGE:-oceanbase/oceanbase-cloud-native:4.2.1.6-106000012024042515}
export OB_IMAGE_STANDALONE=${OB_IMAGE_STANDALONE:-oceanbasedev/oceanbase-cloud-native-dev:4.2.1.2-102000042023120514-alpha.1}
export OB_IMAGE_FAIL_SERVICE=${OB_IMAGE_FAIL_SERVICE:-oceanbase/oceanbase-cloud-native:4.2.1.3-103000032023122818}
export OBAGENT_IMAGE=${OBAGENT_IMAGE:-oceanbase/obagent:4.2.0-100000062023080210}
export OB_IMAGE_UPGRADE=${OB_IMAGE_UPGRADE:-oceanbase/oceanbase-cloud-native:4.3.1.0-100000032024051615}
export OB_IMAGE_420=${OB_IMAGE_420:-oceanbase/oceanbase-cloud-native:4.2.0.0-101000032023091319}
export OB_IMAGE_421=${OB_IMAGE_421:-oceanbase/oceanbase-cloud-native:4.2.1.1-101010012023111012}
export OBAGENT_IMAGE_420=${OBAGENT_IMAGE_420:-oceanbase/obagent:4.2.0-100000062023080210}
export OBAGENT_IMAGE_421=${OBAGENT_IMAGE_421:-oceanbase/obagent:4.2.1-100000092023101717}

# setup variables for nfs
export LOG_ARCHIVE_CUSTOM=${LOG_ARCHIVE_CUSTOM:-log_archive_custom2}
export DATA_BACKUP_CUSTOM=${DATA_BACKUP_CUSTOM:-data_backup_custom2}
export DATA_BACKUP_CUSTOM_MODIFY=${DATA_BACKUP_CUSTOM_MODIFY:-data_backup_custom3}
 
