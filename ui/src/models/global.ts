import { useState } from 'react';

// Global shared data
export default () => {
  const [chooseClusterName, setChooseClusterName] = useState<string>('');
  const [userName, setUsername] = useState<string>();
  const [appInfo, setAppInfo] = useState({});
  return {
    chooseClusterName,
    setChooseClusterName,
    userName,
    setUsername,
    appInfo,
    setAppInfo,
  };
};
