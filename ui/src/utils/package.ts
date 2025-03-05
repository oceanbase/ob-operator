import BigNumber from 'bignumber.js';
import { toNumber } from 'lodash';

/**
 * 支持的比较操作符
 */
export type Comparer =
  // > isGreaterThan
  | 'gt'
  // < isLessThan
  | 'lt'
  // ==  isEqualTo
  | 'eq'
  // >= isGreaterThanOrEqualTo
  | 'gte'
  // <=  isLessThanOrEqualTo
  | 'lte';

/**
 * 版本比较，返回值含义如下:
 * 1: 大于
 * 0: 等于
 * -1: 小于
 * 该函数不导出，外部直接使用 versionCompare 和 buildVersionCompare 函数即可
 * */
function versionCompareValue(version1?: string, version2?: string) {
  const v1 = version1?.split('.') || [];
  const v2 = version2?.split('.') || [];
  for (let i = 0; i < v1.length || i < v2.length; ++i) {
    let x = 0;
    let y = 0;
    if (i < v1.length) {
      x = toNumber(v1[i]);
    }
    if (i < v2.length) {
      y = toNumber(v2[i]);
    }
    if (x > y) {
      return 1;
    }
    if (x < y) {
      return -1;
    }
  }
  return 0;
}

/**
 * 版本比较，支持复合比较，比如 gte、lte
 */
export function versionCompare(
  version1?: string,
  version2?: string,
  comparer?: Comparer,
) {
  // 如果版本号为空，直接返回 false
  if (!version1 || !version2) {
    return false;
  }
  const gt = versionCompareValue(version1, version2) === 1;
  const lt = versionCompareValue(version1, version2) === -1;
  const eq = versionCompareValue(version1, version2) === 0;
  if (comparer === 'gt') {
    return gt;
  } else if (comparer === 'lt') {
    return lt;
  } else if (comparer === 'eq') {
    return eq;
  } else if (comparer === 'gte') {
    return gt || eq;
  } else if (comparer === 'lte') {
    return lt || eq;
  }
  // 如果传入的 comparer 不合法，直接返回 false
  return false;
}

/**
 * 大数比较，支持复合比较，比如 gte、lte
 */

function comparerFn(
  buildVersion1?: string,
  buildVersion2?: string,
  comparer?: Comparer,
) {
  const buildNumber1 = new BigNumber(buildVersion1?.split('-')?.[1] || 0);
  const buildNumber2 = new BigNumber(buildVersion2?.split('-')?.[1] || 0);

  if (comparer === 'gt') {
    return buildNumber1.gt(buildNumber2);
  } else if (comparer === 'lt') {
    return buildNumber1.lt(buildNumber2);
  } else if (comparer === 'eq') {
    return buildNumber1.eq(buildNumber2);
  } else if (comparer === 'gte') {
    return buildNumber1.gte(buildNumber2);
  } else if (comparer === 'lte') {
    return buildNumber1.lte(buildNumber2);
  }
  // 如果传入的 comparer 不合法，直接返回 false
  return false;
}

/**
 * 对构建版本号进行比较
 * 构建版本号: ${version}-${buildNumber}
 */
export function buildVersionCompare(
  buildVersion1?: string,
  buildVersion2?: string,
  comparer?: Comparer,
): boolean {
  // 根据构建版本号解析得到语义化版本
  const version1 = buildVersion1?.split('-')?.[0];
  // 根据构建版本号解析得到 buildNumber
  // 同上
  const version2 = buildVersion2?.split('-')?.[0];

  // 语义化版本如果相等，再比较构建版本号
  return versionCompare(version1, version2, 'eq')
    ? comparerFn(buildVersion1, buildVersion2, comparer)
    : versionCompare(version1, version2, comparer);
}

/**
 * 判断版本是否 >= 4.2.0
 */
export const isGte4_2 = (version?: string) => {
  return buildVersionCompare(version, '4.2.0', 'gte');
};

/** 判断版本是否 >= 4.3.3.0 */
export const isGte4_3_3 = (version?: string) => {
  return buildVersionCompare(version, '4.3.3.0', 'gte');
};
