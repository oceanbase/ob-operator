import React, { useEffect } from 'react';
import { getAppInfoFromStorage } from '@/utils/helper';
import JSEncrypt from 'jsencrypt';

export const usePublicKey = () => {
  const [publicKey, setPublicKey] = React.useState<string>('');

  useEffect(() => {
    getAppInfoFromStorage().then((appInfo) => {
      setPublicKey(appInfo.publicKey); 
    }).catch((err) => {
      console.log(err)
    });
  }, [])

  return publicKey;
}

export const encryptText = (text: string, publicKey: string): string | false => {
  const encrypt = new JSEncrypt(); 
  encrypt.setPublicKey(publicKey);
  return encrypt.encrypt(text);
}