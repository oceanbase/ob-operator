import type { CommonKVPair } from '@/api/generated';
import { ObproxyCreateOBProxyParam } from '@/api/generated';

const buildLabelsMap = (labels: CommonKVPair[]) => {
  const labelsMap = new Map();
  for (const label of labels) {
    labelsMap.set(label.key, label.value);
  }
  return labelsMap;
};

/**
 *
 * @description Determines whether two parameter lists are different
 */
export const isDifferentParams = (
  newParams: CommonKVPair[],
  oldParams: CommonKVPair[],
) => {
  const newParamsMap = buildLabelsMap(newParams),
    oldParamsMap = buildLabelsMap(oldParams);
  if (newParamsMap.size !== oldParamsMap.size) return true;
  for (const key of newParamsMap.keys()) {
    if (newParamsMap.get(key) !== oldParamsMap.get(key)) return true;
  }
  return false;
};

export const filterParams = (proxyConfig: ObproxyCreateOBProxyParam) => {
  proxyConfig.parameters = proxyConfig.parameters?.filter(
    (param) => param.key && param.value,
  );
  if (!proxyConfig.parameters || !proxyConfig.parameters.length)
    delete proxyConfig.parameters;
};
