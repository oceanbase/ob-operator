import { intl } from '@/utils/intl';
import { message } from 'antd';

export const errorHandling = (error: any) => {
  if (error?.response?.status === 401) {
    message.warning(
      intl.formatMessage({
        id: 'src.api.2CA64FC6',
        defaultMessage: '登录已过期',
      }),
    );
    location.href = '/#/login';
  } else {
    const { response } = error;
    if (
      response?.status === 400 &&
      response?.config?.url?.split('/')?.pop() === 'password'
    ) {
      if (
        response?.data?.message === 'Error BadRequest: password is incorrect'
      ) {
        message.error(
          intl.formatMessage({
            id: 'src.api.855C826A',
            defaultMessage: '原密码输入不正确',
          }),
        );
      }
      if (
        response?.data?.message ===
        'Error BadRequest: new password is the same as the old password'
      ) {
        message.error(
          intl.formatMessage({
            id: 'src.api.5D7E1724',
            defaultMessage: '新密码不能和初始密码一致',
          }),
        );
      }
    } else if (response?.status === 403) {
      message.warning(
        intl.formatMessage({
          id: 'src.api.1717A275',
          defaultMessage: '无权限访问',
        }),
      );
    } else {
      message.error(error?.response?.data?.message || error.message);
    }
  }
  return Promise.reject(error.response);
};
