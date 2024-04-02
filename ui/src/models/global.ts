import { useRef, useState } from 'react';

// Global shared data
export default () => {
  const [chooseClusterName, setChooseClusterName] = useState<string>('');
  const [userName, setUsername] = useState<string>();
  const reportDataInterval = useRef<NodeJS.Timer>();
  return {
    chooseClusterName,
    setChooseClusterName,
    userName,
    setUsername,
    reportDataInterval,
  };
};
