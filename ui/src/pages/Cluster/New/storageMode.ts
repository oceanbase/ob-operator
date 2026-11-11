/** Normalize preserved form values without mutating the Ant Design form state. */
export function normalizeStorageMode<
  T extends {
    deploymentMode?: string;
    observer: { storage: { redoLog?: unknown } };
    sharedStorageInfo?: unknown;
    logServiceRef?: unknown;
  },
>(values: T): T {
  const normalized = {
    ...values,
    observer: { ...values.observer, storage: { ...values.observer.storage } },
  };
  if (values.deploymentMode === 'shared_storage') {
    delete normalized.observer.storage.redoLog;
  } else {
    delete normalized.sharedStorageInfo;
    delete normalized.logServiceRef;
  }
  return normalized;
}
