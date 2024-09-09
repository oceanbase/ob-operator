
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

export { generateRandomPassword};
