import TopoComponent from '@/components/TopoComponent';
import { useParams } from '@umijs/max';
import { getTenant } from '@/services/tenant';
import {
  getEssentialParameters as getEssentialParametersReq,
  getSimpleClusterList,
  } from '@/services';
import { useRequest } from 'ahooks';
import BasicInfo from '../Overview/BasicInfo';
import { REFRESH_TENANT_TIME,RESULT_STATUS } from '@/constants';
import { getClusterFromTenant } from '../../helper';
import { useState, useEffect, useRef } from 'react';

export default function Topo() {
  const { ns, name } = useParams();
  const [clusterList, setClusterList] = useState<API.SimpleClusterList>([]);
  const [editZone, setEditZone] = useState<string>('');
  const timerRef = useRef<NodeJS.Timeout>();
  const [defaultUnitCount, setDefaultUnitCount] = useState<number>(1);
  const { data: tenantResponse,refresh: reGetTenantDetail } = useRequest(getTenant, {
    defaultParams: [{ ns, name }],
    onSuccess:({ data, successful }) => {
      if (successful) {
        if (data.info.unitNumber) {
          setDefaultUnitCount(data.info.unitNumber);
        }
        if (!RESULT_STATUS.includes(data.info.status)) {
          timerRef.current = setTimeout(() => {
            reGetTenantDetail();
          }, REFRESH_TENANT_TIME);
        } else if (timerRef.current) {
          clearTimeout(timerRef.current);
        }
      }
    }
  });

  useRequest(getSimpleClusterList, {
    onSuccess: ({ successful, data }) => {
      if (successful) {
        data.forEach((cluster) => {
          cluster.topology.forEach((zone) => {
            zone.checked = false;
          });
        });
        setClusterList(data);
      }
    },
  });
  const { data: essentialParameterRes, run: getEssentialParameters } =
  useRequest(getEssentialParametersReq, {
    manual: true,
  });
  const tenantTopoData = tenantResponse?.data;
  const essentialParameter = essentialParameterRes?.data;
  useEffect(() => {
    if (tenantTopoData && clusterList) {
      const cluster = getClusterFromTenant(
        clusterList,
        tenantTopoData.info.clusterResourceName,
      );
      if (cluster) {
        const { name, namespace } = cluster;
        getEssentialParameters({
          ns: namespace,
          name,
        });
      }
    }
  }, [clusterList, tenantTopoData]);

  return (
    <div>
      {tenantTopoData && (
        <TopoComponent
          defaultUnitCount={defaultUnitCount}
          tenantReplicas={tenantTopoData.replicas}
          clusterNameOfKubectl={tenantTopoData.info.clusterResourceName}
          namespace={tenantTopoData.info.namespace}
          resourcePoolDefaultValue={{
            clusterList: clusterList,
            essentialParameter,
            clusterResourceName: tenantTopoData?.info.clusterResourceName,
            setClusterList,
            setEditZone,
            replicaList: tenantTopoData?.replicas,
            editZone,
          }}
          refreshTenant={reGetTenantDetail}
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
