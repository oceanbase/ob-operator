import en_US from '@/i18n/strings/en-US.json';
import zh_CN from '@/i18n/strings/zh-CN.json';
import { createIntl } from 'react-intl';

const messages = {
  'en-US': en_US,
  'zh-CN': zh_CN,
};

export const getLocale = () => {
  const lang =
    typeof localStorage !== 'undefined'
      ? window.localStorage.getItem('umi_locale')
      : '';
  return lang || 'zh-CN';
};

export const locale = getLocale();

export const intl = createIntl({
  locale,
  messages: messages[locale],
});
