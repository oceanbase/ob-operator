import { encryptText } from '@/hook/usePublicKey';
import dayjs from 'dayjs';
import { clone } from 'lodash';
export function formatNewTenantForm(
  originFormData: any,
  clusterName: string,
  publicKey: string,
): API.TenantBody {
  let result: API.TenantBody = {};
  Object.keys(originFormData).forEach((key) => {
    if (key === 'connectWhiteList') {
      result[key] = originFormData[key].join(',');
    } else if (key === 'obcluster') {
      result[key] = clusterName;
    } else if (key === 'pools') {
      result[key] = Object.keys(originFormData[key])
        .map((zone) => ({
          zone,
          priority: originFormData[key]?.[zone]?.priority,
        }))
        .filter((item) => item.priority);
    } else if (key === 'source') {
      if (originFormData[key]['tenant'] || originFormData[key]['restore'])
        result[key] = {};
      if (originFormData[key]['tenant']) {
        result[key]['tenant'] = originFormData[key]['tenant'];
      }
      if (originFormData[key]['restore']) {
        let { until } = originFormData[key]['restore'];
        result[key]['restore'] = {
          ...originFormData[key]['restore'],
          ossAccessId: encryptText(
            originFormData[key]['restore'].ossAccessId,
            publicKey,
          ),
          ossAccessKey: encryptText(
            originFormData[key]['restore'].ossAccessKey,
            publicKey,
          ),
          until:
            until && until.date && until.time
              ? {
                  timestamp:
                    dayjs(until.date).format('YYYY-MM-DD') +
                    ' ' +
                    dayjs(until.time).format('HH:mm:ss'),
                }
              : { unlimited: true },
        };
        if (originFormData[key]['restore'].bakEncryptionPassword) {
          result[key]['restore']['bakEncryptionPassword'] = encryptText(
            originFormData[key]['restore'].bakEncryptionPassword,
            publicKey,
          );
        } else {
          delete result[key]['restore']['bakEncryptionPassword'];
        }
      }
    } else if (key === 'rootPassword') {
      result[key] = encryptText(originFormData[key], publicKey);
    } else {
      result[key] = originFormData[key];
    }
  });
  console.log('result', result);

  return result;
}
/**
 * encrypt ossAccessId,ossAccessKey,bakEncryptionPassword
 *
 * format scheduleDates
 */
export function formatNewBackupForm(originFormData: any, publicKey: string) {
  let formData = clone(originFormData);
  if (formData.bakEncryptionPassword) {
    formData.bakEncryptionPassword = encryptText(
      originFormData.bakEncryptionPassword,
      publicKey,
    );
  }
  formData.ossAccessId = encryptText(originFormData.ossAccessId, publicKey);
  formData.ossAccessKey = encryptText(originFormData.ossAccessKey, publicKey);
  formData.scheduleTime = dayjs(formData.scheduleTime).format('HH:MM');
  delete formData.scheduleDates.days;
  formData.scheduleDates = Object.keys(formData.scheduleDates).map((key) => ({
    day: Number(key),
    backupType: formData.scheduleDates[key],
  }));
  return formData;
}
