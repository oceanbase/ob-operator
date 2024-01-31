export const getInitialObjOfKeys = ( targetObj: any,keys: string[]) => {
  return keys.reduce((pre, cur) => {
    pre[cur] = targetObj[cur];
    return pre;
  }, {});
};
