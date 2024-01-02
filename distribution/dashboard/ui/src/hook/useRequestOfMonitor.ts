import type {
  BaseOptions,
  CombineService,
} from '@ahooksjs/use-request/es/types';
import { useRequest } from 'ahooks';
import { useCallback, useRef } from 'react';

const useLockFn = (fn: (...args: any) => Promise<any>) => {
  let lockRef = useRef(false);

  return [
    useCallback(
      async (...args: any) => {
        if (lockRef.current) return;
        lockRef.current = true;
        try {
          const ret = await fn(...args);
          lockRef.current = false;
          return ret;
        } catch (err) {
          lockRef.current = false;
          throw err;
        }
      },
      [fn],
    ),
    (val: boolean) => {
      lockRef.current = val;
    },
  ];
};

const useRequestOfMonitor = <R, P extends any[]>(
  service: CombineService<R, P>,
  options?: BaseOptions<R, P> & { isRealTime?: boolean },
) => {
  const { isRealTime } = options || {}
  const requestResult = useRequest(service, { ...options });
  // 使用竞态锁来防止监控请求堆积
  const [run, setIsLock] = useLockFn(requestResult.run);
  if (!isRealTime) setIsLock(false);
  return {
    ...requestResult,
    run: (params) => {
      if (!isRealTime) {
        setIsLock(false);
        requestResult.run(params);
      } else {
        run(params);
      }
    },
  };
};

export { useRequestOfMonitor };
