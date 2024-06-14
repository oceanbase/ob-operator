import type { CommonKVPair } from '@/api/generated';

/**
 * 
 * @description Determines whether two parameter lists are different
 */
export const isDifferentParams = (
  newParams: CommonKVPair[],
  oldParams: CommonKVPair[],
) => {
  if (newParams.length !== oldParams.length) return true;
  for (const newParam of newParams) {
    const oldParam = oldParams.find((item) => item.key === newParam.key);
    if (!oldParam || oldParam.value !== newParam.value) return true;
  }
  return false;
};
