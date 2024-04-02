import { REPORT_PARAMS_MAP, reportData } from '.';
import {
  addObzone,
  createObclusterReq,
  deleteObcluster,
  deleteObzone,
  scaleObserver,
  upgradeObcluster,
} from '..';

export async function createClusterReportWrap({ ...params }: { body: any }) {
  const r = await createObclusterReq(params);
  if (r.successful) {
    reportData({
      ...REPORT_PARAMS_MAP['createCluster'],
      data: r.data,
    });
  }
  return r;
}

export async function deleteClusterReportWrap({
  ...params
}: API.NamespaceAndName) {
  const r = await deleteObcluster(params);
  if (r.successful) {
    reportData({
      ...REPORT_PARAMS_MAP['deleteCluster'],
      data: r.data,
    });
  }
  return r;
}

export async function upgradeClusterReportWrap({
  ...params
}: API.NamespaceAndName & { image: string }) {
  const r = await upgradeObcluster(params);
  if (r.successful) {
    reportData({
      ...REPORT_PARAMS_MAP['upgradeCluster'],
      data: r.data,
    });
  }
  return r;
}

export async function addObzoneReportWrap({ ...params }: API.AddZoneParams) {
  const r = await addObzone(params);
  if (r.successful) {
    reportData({ ...REPORT_PARAMS_MAP['addZone'], data: r.data });
  }
  return r;
}

export async function deleteObzoneReportWrap({
  ...params
}: API.NamespaceAndName & {
  zoneName: string;
}) {
  const r = await deleteObzone(params);
  if (r.successful) {
    reportData({ ...REPORT_PARAMS_MAP['deleteZone'], data: r.data });
  }
  return r;
}

export async function scaleObserverReportWrap({
  ...params
}: API.ScaleObserverPrams) {
  const r = await scaleObserver(params);
  if (r.successful) {
    reportData({ ...REPORT_PARAMS_MAP['scaleZone'], data: r.data });
  }
  return r;
}
