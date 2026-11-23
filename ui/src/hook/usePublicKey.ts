import { getAppInfo } from '@/services';
import { intl } from '@/utils/intl';
import { useModel } from '@umijs/max';
import { message } from 'antd';
import JSEncrypt from 'jsencrypt';
import { useEffect } from 'react';

export const usePublicKey = () => {
  const { publicKey, setPublicKey } = useModel('global');

  useEffect(() => {
    getAppInfo()
      .then(({ data }) => {
        setPublicKey(data.publicKey);
      })
      .catch((err) => {
        console.log(err);
      });
  }, []);

  return publicKey;
};

export const encryptText = (
  text: string,
  publicKey: string,
): string | false => {
  const encrypt = new JSEncrypt();
  encrypt.setPublicKey(publicKey);
  return encrypt.encrypt(text);
};

// A new login or password-reset page can be submitted before its first request completes.
export async function encryptPassword(
  value: string,
): Promise<string | undefined> {
  const { data } = await getAppInfo();
  const encrypted = data?.publicKey && encryptText(value, data.publicKey);
  if (!encrypted) {
    message.error(
      intl.formatMessage({
        id: 'dashboard.password.EncryptionFailed',
        defaultMessage: '密码加密失败，请重试',
      }),
    );
    return;
  }
  return encrypted;
}
