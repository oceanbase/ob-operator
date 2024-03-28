import { REPORT_PARAMS_MAP, reportData } from '.';
import {
  addObzone,
  createObclusterReq,
  deleteObcluster,
  deleteObzone,
  scaleObserver,
  upgradeObcluster,
} from '..';

export async function createClusterReportWrap({
  version,
  ...params
}: {
  version: string;
  body: any;
}) {
  const r = await createObclusterReq(params);
  if (r.successful) {
    reportData({
      ...REPORT_PARAMS_MAP['createCluster'],
      version,
      data: r.data,
    });
  }
  return r;
}

export async function deleteClusterReportWrap({
  version,
  ...params
}: API.NamespaceAndName & { version: string }) {
  const r = await deleteObcluster(params);
  if (r.successful) {
    reportData({
      ...REPORT_PARAMS_MAP['deleteCluster'],
      version,
      data: r.data,
    });
  }
  return r;
}

export async function upgradeClusterReportWrap({
  version,
  ...params
}: API.NamespaceAndName & { version: string; image: string }) {
  const r = await upgradeObcluster(params);
  if (r.successful) {
    reportData({
      ...REPORT_PARAMS_MAP['upgradeCluster'],
      version,
      data: r.data,
    });
  }
  return r;
}

export async function addObzoneReportWrap({
  version,
  ...params
}: API.AddZoneParams & { version: string }) {
  const r = await addObzone(params);
  if (r.successful) {
    reportData({ ...REPORT_PARAMS_MAP['addZone'], version, data: r.data });
  }
  return r;
}

export async function deleteObzoneReportWrap({
  version,
  ...params
}: API.NamespaceAndName & {
  zoneName: string;
  version: string;
}) {
  const r = await deleteObzone(params);
  if (r.successful) {
    reportData({ ...REPORT_PARAMS_MAP['deleteZone'], version, data: r.data });
  }
  return r;
}

export async function scaleObserverReportWrap({
  version,
  ...params
}: API.ScaleObserverPrams & { version: string }) {
  const r = await scaleObserver(params);
  if (r.successful) {
    reportData({ ...REPORT_PARAMS_MAP['scaleZone'], version, data: r.data });
  }
  return r;
}
