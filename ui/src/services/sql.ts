import { request } from '@umijs/max';

const sqlPrefix = '/api/v1/sql';

export async function listSqlMetrics(params: {
  language?: string;
}): Promise<API.SqlMetricsResponse> {
  return request(`${sqlPrefix}/metrics`, {
    method: 'GET',
    params,
  });
}

export async function listSqlStats(
  data: API.SqlFilter,
): Promise<API.SqlStatsResponse> {
  return request(`${sqlPrefix}/stats`, {
    method: 'POST',
    data,
  });
}

export async function listRequestStatistics(
  data: API.SqlRequestStatisticParam,
): Promise<API.RequestStatisticsResponse> {
  return request(`${sqlPrefix}/requestStatistics`, {
    method: 'POST',
    data,
  });
}

export async function querySqlDetailInfo(
  data: API.SqlDetailParam,
): Promise<API.SqlDetailResponse> {
  return request(`${sqlPrefix}/querySqlDetailInfo`, {
    method: 'POST',
    data,
  });
}

export async function querySqlHistoryInfo(
  data: API.SqlHistoryParam,
): Promise<API.SqlHistoryInfo> {
  return request(`${sqlPrefix}/querySqlHistoryInfo`, {
    method: 'POST',
    data,
  });
}

export async function queryPlanDetailInfo(
  data: API.PlanDetailParam,
): Promise<API.PlanDetailResponse> {
  return request(`${sqlPrefix}/queryPlanDetailInfo`, {
    method: 'POST',
    data,
  });
}
