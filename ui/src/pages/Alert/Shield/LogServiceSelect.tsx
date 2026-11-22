import { LSItem, lsRequest } from '@/services/logservice';
import { Alert } from '@/type/alert';
import { useAccess } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Select } from 'antd';
export default function LogServiceSelect({
  value,
  onChange,
}: {
  value?: Alert.InstancesType;
  onChange?: (v: Alert.InstancesType) => void;
}) {
  const access = useAccess();
  const { data, loading } = useRequest(() => lsRequest<LSItem[]>(), {
    ready: !!access.oblogserviceread,
  });
  return (
    <Select
      mode="multiple"
      loading={loading}
      value={value?.logservice}
      options={data?.map((ls) => ({
        label: `${ls.namespace}/${ls.name}`,
        value: `${ls.namespace}/${ls.name}`,
      }))}
      onChange={(names) =>
        onChange?.({ type: 'logservice', obcluster: [], logservice: names })
      }
    />
  );
}
