import { intl } from '@/utils/intl';
import { message } from 'antd';

export const errorHandling = (error: any) => {
  if (error?.response?.status === 401) {
    message.warning(
      intl.formatMessage({
        id: 'src.api.2CA64FC6',
        defaultMessage: '登陆已过期',
      }),
    );
    location.href = '/#/login';
  } else {
    const { response } = error;
    if (
      response?.status === 400 &&
      response?.config?.url?.split('/')?.pop() === 'password' &&
      response?.data?.message === 'Error BadRequest: password is incorrect'
    ) {
      message.error('原密码输入不正确');
    } else {
      message.error(error?.response?.data?.message || error.message);
    }
  }
  return Promise.reject(error.response);
};
