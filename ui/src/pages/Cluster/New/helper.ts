import { intl } from '@/utils/intl';
function generateRandomPassword() {
  const length = Math.floor(Math.random() * 25) + 8; // 生成8到32之间的随机长度
  const characters =
    'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789~!@#%^&*_-+=|(){}[]:;,.?/`$"<>'; // 可用字符集合

  let password = '';
  let countUppercase = 0; // 大写字母计数器
  let countLowercase = 0; // 小写字母计数器
  let countNumber = 0; // 数字计数器
  let countSpecialChar = 0; // 特殊字符计数器

  // 生成随机密码
  for (let i = 0; i < length; i++) {
    const randomIndex = Math.floor(Math.random() * characters.length);
    const randomChar = characters[randomIndex];
    password += randomChar;

    // 判断字符类型并增加相应计数器
    if (/[A-Z]/.test(randomChar)) {
      countUppercase++;
    } else if (/[a-z]/.test(randomChar)) {
      countLowercase++;
    } else if (/[0-9]/.test(randomChar)) {
      countNumber++;
    } else {
      countSpecialChar++;
    }
  }

  // 检查计数器是否满足要求
  if (
    countUppercase < 2 ||
    countLowercase < 2 ||
    countNumber < 2 ||
    countSpecialChar < 2
  ) {
    return generateRandomPassword(); // 重新生成密码
  }

  return password;
}

const passwordRules = [
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

export { generateRandomPassword, passwordRules};
