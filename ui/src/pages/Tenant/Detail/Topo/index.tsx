import TopoComponent from '@/components/TopoComponent';
import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { getTenant } from '@/services/tenant';
import { useRequest } from 'ahooks';
import BasicInfo from '../Overview/BasicInfo';

export default function Topo() {
  const [ns, name] = getNSName();
  const { data: tenantResponse } = useRequest(getTenant, {
    defaultParams: [{ ns, name }],
  });
  const tenantTopoData = tenantResponse?.data;
  return (
    <div>
      {tenantTopoData && (
        <TopoComponent
          tenantReplicas={tenantTopoData.replicas}
          clusterNameOfKubectl={tenantTopoData.info.clusterResourceName}
          namespace={tenantTopoData.info.namespace}
          header={
            <BasicInfo
              info={tenantTopoData.info}
              source={tenantTopoData.source}
              style={{ backgroundColor: 'rgb(245, 248, 254)', border: 'none' }}
            />
          }
        />
      )}
    </div>
  );
}
