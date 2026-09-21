// Conservative Dashboard minimum verified with oceanbase-ai 4.6.2.0.
// 30 GiB initializes only 6 GiB, insufficient for SS system cache + reserves.
export const SS_MIN_CACHE_GIB = 50;
export function minimumDataGiB(sharedStorage: boolean, normalMinimum: number) {
  return sharedStorage ? SS_MIN_CACHE_GIB : normalMinimum;
}
