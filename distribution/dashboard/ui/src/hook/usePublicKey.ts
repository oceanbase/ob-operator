import React, { useEffect } from 'react';
import { infoReq } from '@/services';
import JSEncrypt from 'jsencrypt';

export const usePublicKey = () => {
  const [publicKey, setPublicKey] = React.useState<string>('');

  useEffect(() => {
    infoReq().then((res) => {
      setPublicKey(res.data.publicKey); 
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