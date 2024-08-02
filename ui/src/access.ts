import type { AcAccount, AcPolicy } from './api/generated';
import { initialAccess } from './utils/helper';

export type InitialStateType = {
  accountInfo: AcAccount;
  policies: AcPolicy[];
};

export default function (initialState: InitialStateType) {
  const accessObj = Object.create(null);
  if (initialState) initialAccess(accessObj, initialState);
  return {
    ...accessObj,
  };
}
