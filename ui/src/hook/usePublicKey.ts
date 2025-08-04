import { getAppInfo } from '@/services';
import { useModel } from '@umijs/max';
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
