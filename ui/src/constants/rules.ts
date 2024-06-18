import type { AlarmMatcher, CommonKVPair } from '@/api/generated';
import { intl } from '@/utils/intl';
import type { RuleObject } from 'antd/es/form';
import { LABELNAME_REG, TZ_NAME_REG } from '.';

/**
 * Check whether the resource name conforms to the domain name format.
 * The resource name may be spliced into the domain name.
 **/
function checkName(name: string): boolean {
  const regex =
    /[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/g;
  const res = name.match(regex);
  if (res && res[0] === name) return true;
  return false;
}

const resourceNameRule = () => ({
  validator(_: any, value: string) {
    if (checkName(value)) {
      return Promise.resolve();
    }
    return Promise.reject(
      new Error(
        intl.formatMessage({
          id: 'Dashboard.src.constants.rules.TheResourceNameMayBe',
          defaultMessage: '资源名可能拼接到域名中，需要符合域名格式',
        }),
      ),
    );
  },
});

const RULER_ZONE = [
  {
    required: true,
    message: intl.formatMessage({
      id: 'Dashboard.src.constants.rules.EnterAZoneName',
      defaultMessage: '请输入zone名称',
    }),
  },
  {
    pattern: TZ_NAME_REG,
    message: intl.formatMessage({
      id: 'Dashboard.src.constants.rules.TheFirstCharacterMustBe',
      defaultMessage: '首字符必须是字母或者下划线，不能包含 -',
    }),
  },
  resourceNameRule,
];

const LABEL_NAME_RULE = {
  validator: (_: RuleObject, value: AlarmMatcher[] | CommonKVPair[]) => {
    if (
      value.some((item) => !LABELNAME_REG.test(item?.name || item?.key || ''))
    ) {
      return Promise.reject(
        intl.formatMessage({
          id: 'src.constants.93456456',
          defaultMessage:
            '标签名需满足以字母或下划线开头，包含字母，数字，下划线',
        }),
      );
    }
    return Promise.resolve();
  },
};

export { LABEL_NAME_RULE, RULER_ZONE, checkName, resourceNameRule };
