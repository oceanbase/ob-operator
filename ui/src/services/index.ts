import { formatTopoData } from '@/components/TopoComponent/helper';
import { formatClusterData } from '@/pages/Cluster/Detail/Overview/helper';
import { formatStatisticData } from '@/utils/helper';
import { intl } from '@/utils/intl'; //@ts-nocheck
import { request } from '@umijs/max';
import _ from 'lodash';
import moment from 'moment';

const obClusterPrefix = '/api/v1/obclusters';
const clusterPrefix = '/api/v1/cluster';

export async function loginReq(body: API.User) {
  return request('/api/v1/login', {
    method: 'POST',
    data: body,
  });
}

export async function logoutReq() {
  return request('/api/v1/logout', {
    method: 'POST',
  });
}

export async function infoReq() {
  return request('/api/v1/info', {
    method: 'GET',
  });
}

/**
 * If no parameters are passed, all events will be returned.
 */
export async function getEventsReq(params: API.EventParams) {
  const r = await request(`${clusterPrefix}/events`, {
    method: 'GET',
    params,
  });
  if (r.successful) {
    let count = 0;
    r.data.sort((pre, next) => next.lastSeen - pre.lastSeen);
    for (let event of r.data) {
      event.id = ++count;
      event.firstOccur = moment
        .unix(event.firstOccur)
        .format('YYYY-MM-DD HH:mm:ss');
      event.lastSeen = moment
        .unix(event.lastSeen)
        .format('YYYY-MM-DD HH:mm:ss');
    }
  }
  return r.data;
}

export async function getNodeInfoReq() {
  const r = await request(`${clusterPrefix}/nodes`, { method: 'GET' });
  let res = [];
  if (r.successful) {
    for (let node of r.data) {
      let obj = {};
      Object.assign(obj, node.info, node.resource);
      obj.cpu = ((obj.cpuUsed / obj.cpuTotal) * 100).toFixed(1);
      obj.memory = ((obj.memoryUsed / obj.memoryTotal) * 100).toFixed(1);
      obj.uptime = moment.unix(obj.uptime).format('YYYY-MM-DD HH:mm:ss');
      res.push(obj);
    }
  }
  return res;
}

export async function getNodeLabelsReq() {
  const r = await request(`${clusterPrefix}/nodes`, { method: 'GET' });
  let res = { key: [], value: [], originLabels: [] };
  if (r.successful) {
    for (let node of r.data) {
      res.originLabels = res.originLabels.concat(node.info.labels);
    }
    for (let label of res.originLabels) {
      const { key, value } = label;
      res.key.push({ label: key, value: key });
      res.value.push({ label: value, value: value });
    }
    res.key = _.uniqBy(res.key, 'label');
    res.value = _.uniqBy(res.value, 'label');
  }

  return res;
}

export async function getClusterStatisticReq(): Promise<API.StatisticDataResponse> {
  const r = await request(`${obClusterPrefix}/statistic`, {
    method: 'GET',
  });
  return {
    ...r,
    data: formatStatisticData('cluster', r.data),
  };
}

export async function createObclusterReq(body: any) {
  const r = await request(obClusterPrefix, { method: 'POST', data: body });
  if (r.successful && !r.message) {
    r.message = intl.formatMessage({
      id: 'OBDashboard.src.services.OperationSucceededTheClusterIs',
      defaultMessage: '操作成功！集群正在创建中',
    });
  }
  return {
    successful: r.successful,
    data: r.data,
    message: r.message,
  };
}

export async function getObclusterListReq() {
  const r = await request<API.ClusterListResponse>(obClusterPrefix, {
    method: 'GET',
  });
  if (r.successful) {
    let res: API.ClusterList = [];
    for (let cluster of r.data) {
      let obj = {};
      cluster.createTime = moment
        .unix(cluster.createTime)
        .format('YYYY-MM-DD HH:mm:ss');
      for (let key in cluster) {
        if (key !== 'metrics') {
          obj[key] = cluster[key];
        } else if (cluster['metrics']) {
          obj['cpuPercent'] = cluster['metrics']['cpuPercent'];
          obj['memoryPercent'] = cluster['metrics']['memoryPercent'];
          obj['diskPercent'] = cluster['metrics']['diskPercent'];
        }
      }
      res.push(obj);
    }
    return {
      ...r,
      data:res
    }
  }

  return r;
}

export async function getSimpleClusterList(): Promise<API.SimpleClusterListResponse> {
  const r = await request<API.SimpleClusterListResponse>(obClusterPrefix, {
    method: 'GET',
  });
  if (r.successful) {
    return{
      ...r,
      data:r.data.map((clusterDetail) => ({
        clusterId: clusterDetail.clusterId,
        name: clusterDetail.name,
        namespace: clusterDetail.namespace,
        topology: clusterDetail.topology,
        clusterName: clusterDetail.clusterName,
        status: clusterDetail.status
      }))
    }
  };
  return r;
}

export async function getClusterDetailReq({
  ns,
  name,
  useFor,
  tenantReplicas,
}: {
  ns: string;
  name: string;
  useFor?: string;
  tenantReplicas?: API.ReplicaDetailType[];
}) {
  const r = await request(`${obClusterPrefix}/namespace/${ns}/name/${name}`, {
    method: 'GET',
  });
  if (r.successful) {
    if (useFor === 'topo') return formatTopoData(r.data, tenantReplicas);
    return formatClusterData(r.data);
  }
  return r.data;
}
export async function upgradeObcluster({
  ns,
  name,
  image,
}: {
  ns: string;
  name: string;
  image: string;
}) {
  const r = await request(`${obClusterPrefix}/namespace/${ns}/name/${name}`, {
    method: 'POST',
    data: { image },
  });
  return {
    successful: r.successful,
    data: r.data,
    message: r.message,
  };
}
export async function scaleObserver({
  namespace,
  name,
  zoneName,
  replicas,
}: {
  namespace: string;
  name: string;
  zoneName: string;
  replicas: number;
}) {
  const r = await request(
    `${obClusterPrefix}/namespace/${namespace}/name/${name}/obzones/${zoneName}/scale`,
    {
      method: 'POST',
      data: { replicas },
    },
  );
  if (r.successful && !r.message)
    r.message = intl.formatMessage({
      id: 'OBDashboard.src.services.OperationSucceeded',
      defaultMessage: '操作成功！',
    });
  return {
    successful: r.successful,
    data: r.data,
    message: r.message,
  };
}
export async function addObzone({
  namespace,
  name,
  ...body
}: {
  namespace: string;
  name: string;
  zone: string;
  replicas: number;
  nodeSelector: { key: string; value: string }[];
}) {
  const r = await request(
    `${obClusterPrefix}/namespace/${namespace}/name/${name}/obzones`,
    { method: 'POST', data: body },
  );
  if (r.successful && !r.message)
    r.message = intl.formatMessage({
      id: 'OBDashboard.src.services.TheOperationIsSuccessfulAnd',
      defaultMessage: '操作成功，正在添加中',
    });
  return {
    successful: r.successful,
    data: r.data,
    message: r.message,
  };
}
export async function deleteObcluster({
  ns,
  name,
}: {
  ns: string;
  name: string;
}) {
  const r = await request(`${obClusterPrefix}/namespace/${ns}/name/${name}`, {
    method: 'DELETE',
  });
  return {
    successful: r.successful,
    data: r.data,
    message: r.message,
  };
}

export async function deleteObzone({
  ns,
  name,
  zoneName,
}: {
  ns: string;
  name: string;
  zoneName: string;
}) {
  const r = await request(
    `${obClusterPrefix}/namespace/${ns}/name/${name}/obzones/${zoneName}`,
    {
      method: 'DELETE',
    },
  );
  return {
    successful: r.successful,
    data: r.data,
    message: r.message,
  };
}

export async function deleteObserver({
  ns,
  name,
}: {
  ns: string;
  name: string;
}) {
  const r = await request(`/api/v1/observers/namespace/${ns}/name/${name}`, {
    method: 'DELETE',
  });
  return {
    successful: r.successful,
    data: r.data,
  };
}

export async function getNameSpaces() {
  const r = await request(`${clusterPrefix}/namespaces`, {
    method: 'GET',
  });
  if (r.successful) {
    let res = [];
    for (let item of r.data) {
      res.push({
        value: item.namespace,
        label: item.namespace,
        disabled: item.status !== 'Active',
      });
    }
    return res;
  }
  return r.data;
}

export async function createNameSpace(namespace: string) {
  const r = await request(`${clusterPrefix}/namespaces`, {
    method: 'POST',
    data: { namespace },
  });
  return {
    successful: r.successful,
    data: r.data,
  };
}

export async function getStorageClasses(): Promise<API.StorageClassesResponse> {
  const r = await request(`${clusterPrefix}/storageClasses`, {
    method: 'GET',
  });
  if (r.successful) {
    let res = [];
    for (let item of r.data) {
      let toolTipData = [];
      Object.keys(item).forEach((key) => {
        // if (key !== 'name') toolTipData.push({[key]:item[key]})
        toolTipData.push({ [key]: item[key] });
      });
      res.push({
        value: item.name,
        label: item.name,
        toolTipData: toolTipData,
      });
    }
    return {
      ...r,
      data: res,
    };
  }
  return r;
}

export async function getAllMetrics(type: API.MetricScope) {
  const r = await request('/api/v1/metrics', {
    method: 'GET',
    params: { scope: type },
  });
  return r.data;
}

const setMetricNameFromLabels = (labels:API.MetricsLabels)=>{
  let tenantName = labels.find((label) => label.key === 'tenant_name')?.value;
  let clustetName = labels
    .filter((label) => label.key === 'ob_cluster_name')
    .map((label) => label.value)
    .join(',');
  
  return `${tenantName}(${clustetName})`
}

const filterMetricsData = (
  type: 'tenant' | 'cluster',
  metricsData: any,
  filterData: API.ClusterItem[] | API.TenantDetail[],
) => {
  return metricsData.filter((item) => {
    let targetName =
      item?.metric?.labels?.find(
        (label) =>
          label.key === (type === 'tenant' ? 'tenant_name' : 'ob_cluster_name'),
      ).value || '';
    if (type === 'cluster') {
      return !!filterData.find((cluster) => cluster.clusterName === targetName);
    } else {
      return !!filterData.find((tenant) => tenant.tenantName === targetName);
    }
  });
};

export async function queryMetricsReq({
  useFor,
  type,
  filterData,
  ...data
}: API.QueryMetricsType) {
  const r = await request('/api/v1/metrics/query', {
    method: 'POST',
    data,
  });
  if (r.successful) {
    if(filterData){
      r.data = filterMetricsData(useFor, r.data, filterData);
    }
    if (!r.data || !r.data.length) return [];
    r.data.forEach((metric) => {
      metric.values.forEach((item) => {
        // item.date = moment.unix(item.timestamp).format('YYYY-MM-DD HH:mm:ss');
        item.date = item.timestamp * 1000;
        if (type === 'OVERVIEW') {
          if (useFor === 'tenant') {
            let metricLabels = metric.metric.labels;
            if (metricLabels.length > 1) {
              item.name = setMetricNameFromLabels(metricLabels);
            } else {
              item.name = metricLabels[0]?.value || '';
            }
          } else {
            item.name =
              metric.metric.labels.find(
                (label) => label.key === 'ob_cluster_name',
              ).value || '';
          }
        } else {
          item.name = metric.metric.name;
        }
      });
    });
    let res = _.flatten(r.data.map((metric) => metric.values));
    return res;
  }
  return r.data || [];
}

export async function getEssentialParameters({
  ns,
  name,
}: API.NamespaceAndName): Promise<API.EssentialParametersTypeResponse> {
  return request(`${obClusterPrefix}/${ns}/${name}/resource-usages`);
}