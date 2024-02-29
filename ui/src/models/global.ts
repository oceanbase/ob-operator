import { useState } from 'react';

// 全局共享数据
export default () => {
  const [chooseClusterName, setChooseClusterName] = useState<string>('');
  const [userName, setUsername] = useState<string>();
  return {
    chooseClusterName,
    setChooseClusterName,
    userName,
    setUsername,
  };
};
