import { intl } from '@/utils/intl';

export const passwordRules = [
  {
    required: true,
    message: intl.formatMessage({
      id: 'OBDashboard.Cluster.New.helper.EnterAPassword',
      defaultMessage: '请输入密码',
    }),
  },
  () => ({
    validator(_: any, value: string) {
      if (value.length >= 8 && value.length <= 32) {
        return Promise.resolve();
      }
      return Promise.reject(
        new Error(
          intl.formatMessage({
            id: 'OBDashboard.Cluster.New.helper.ToCharactersInLength',
            defaultMessage: '长度为 8~32 个字符',
          }),
        ),
      );
    },
  }),
  () => ({
    validator(_: any, value: string) {
      const regex = /^[a-zA-Z0-9~!@#%^&*\-_+=|(){}[\]:;,.?/`$"<>]+$/;
      if (regex.test(value)) {
        return Promise.resolve();
      }
      return Promise.reject(
        new Error(
          intl.formatMessage({
            id: 'OBDashboard.Cluster.New.helper.CanOnlyContainLettersNumbers',
            defaultMessage:
              '只能包含字母、数字和特殊字符（~!@#%^&*_-+=|(){}[]:;,.?/`$"<>）',
          }),
        ),
      );
    },
  }),
  () => ({
    validator(_: any, value: string) {
      if (
        /^(?=(.*[a-z]){2,})(?=(.*[A-Z]){2,})(?=(.*\d){2,})(?=(.*[~!@#%^&*_\-+=|(){}[\]:;,.?/`$'"<>\\]){2,})[A-Za-z\d~!@#%^&*_\-+=|(){}[\]:;,.?/`$'"<>\\]{2,}$/.test(
          value,
        )
      ) {
        return Promise.resolve();
      }
      return Promise.reject(
        new Error(
          intl.formatMessage({
            id: 'OBDashboard.Cluster.New.helper.AtLeastUppercaseAndLowercase',
            defaultMessage: '大小写字母、数字和特殊字符都至少包含 2 个',
          }),
        ),
      );
    },
  }),
];
