import type { AcAccount, AcPolicy } from './api/generated';
import { initializeAccess } from './utils/helper';

export type InitialStateType = {
  accountInfo: AcAccount;
  policies: AcPolicy[];
};

export default function (initialState: InitialStateType): {
  [T: string]: boolean;
} {
  const accessObj = Object.create(null);
  if (initialState) initializeAccess(accessObj, initialState);
  return {
    ...accessObj,
  };
}
