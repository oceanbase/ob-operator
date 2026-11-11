import { Select } from 'antd';
import { useRequest } from 'ahooks';
import { useAccess } from '@umijs/max';
import { LSItem, lsRequest } from '@/services/logservice';
import { Alert } from '@/type/alert';
export default function LogServiceSelect({ value, onChange }: { value?: Alert.InstancesType; onChange?: (v: Alert.InstancesType) => void }) {
  const access = useAccess();
  const { data, loading } = useRequest(() => lsRequest<LSItem[]>(), { ready: !!access.oblogserviceread });
  return <Select mode="multiple" loading={loading} value={value?.logservice} options={data?.map(ls => ({ label: `${ls.namespace}/${ls.name}`, value: `${ls.namespace}/${ls.name}` }))} onChange={names => onChange?.({ type: 'logservice', obcluster: [], logservice: names })} />;
}
