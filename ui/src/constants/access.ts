import { intl } from '@/utils/intl';

export const ACCESS_ROLES_LIST = [
{
  label: intl.formatMessage({ id: "src.constants.D057DB4A", defaultMessage: "集群与租户" }),
  value: 'obcluster'
},
{
  label: intl.formatMessage({ id: "src.constants.9E29D52E", defaultMessage: "系统信息" }),
  value: 'system'
},
{
  label: intl.formatMessage({ id: "src.constants.4EBEFDA8", defaultMessage: "监控告警" }),
  value: 'alarm'
},
{
  label: 'OBProxy',
  value: 'obproxy'
},
{
  label: intl.formatMessage({ id: "src.constants.939130EF", defaultMessage: "权限控制" }),
  value: 'ac'
}];