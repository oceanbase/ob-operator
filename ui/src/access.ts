import type { AcAccount, AcPolicy } from './api/generated';
import { initialAccess } from './utils/helper';

export type InitialStateType = {
  accountInfo: AcAccount;
  policies: AcPolicy[];
};

export default function (initialState: InitialStateType) {
  console.log('initialState', initialState);
  const accessObj = Object.create(null);
  if (initialState) initialAccess(accessObj, initialState);
  console.log('accessObj', accessObj);

  return {
    ...accessObj,
  };
}
