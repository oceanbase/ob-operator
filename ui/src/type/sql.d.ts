declare namespace API {
  export interface SqlStatisticMetric {
    name: string;
    value: number;
  }

  export interface SqlDiagnoseInfo {
    reason: string;
    suggestion?: string;
  }

  export interface SqlMetaInfo {
    svrIp: string;
    svrPort: number;
    tenantId: number;
    tenantName: string;
    userId: number;
    userName: string;
    dbId: number;
    dbName: string;
    sqlId: string;
    planId: number;
    querySql: string;
    clientIp: string;
    event: string;
    formatSqlId: string;
    effectiveTenantId: number;
    traceId: string;
    sid: number;
    userClientIp: string;
    txId: string;
    subPlanCount: number;
    lastFailInfo: number;
    causeType: number;
  }

  export interface SqlInfo extends SqlMetaInfo {
    executionStatistics: SqlStatisticMetric[];
    latencyStatistics: SqlStatisticMetric[];
    diagnoseInfo?: SqlDiagnoseInfo[];
  }

  export interface MetricValue {
    timestamp: number;
    value: number;
  }

  export interface Metric {
    name: string;
    labels: { key: string; value: string }[] | null;
  }

  export interface MetricData {
    metric: Metric;
    values: MetricValue[];
  }

  export interface IndexInfo {
    tableName: string;
    indexType: string;
    uniqueness: string;
    indexName: string;
    columns: string[];
    status: string;
  }

  export interface SqlDetailedInfo {
    executionTrend: MetricData[];
    latencyTrend: MetricData[];
    diagnoseInfo?: SqlDiagnoseInfo[];
    plans: PlanStatistic[];
    indexies?: IndexInfo[];
  }

  export type MetricCategory = 'meta' | 'latency' | 'execution';

  export interface SqlMetricMeta {
    key: string;
    name: string;
    description: string;
    unit: string;
    displayByDefault: boolean;
    immutable?: boolean;
  }

  export interface SqlMetricMetaCategory {
    category: MetricCategory;
    metrics: SqlMetricMeta[];
  }

  export interface RequestStatisticInfo {
    tenant: string;
    user: string;
    database: string;
    planCategoryStatistics: SqlStatisticMetric[];
    totalExecutions: number;
    failedExecutions: number;
    totalLatency: number;
    averageLatency: number;
    executionTrend: MetricValue[];
    latencyTrend: MetricValue[];
  }

  export type PlanCategory = 'local' | 'remote' | 'distributed';

  export interface PlanIdentity {
    tenantID: number;
    svrIP: string;
    svrPort: number;
    planID: number;
  }

  export interface PlanMeta extends PlanIdentity {
    planHash: string;
    generatedTime: number;
  }

  export interface PlanStatistic extends PlanMeta {
    ioCost: number;
    cpuCost: number;
    cost: number;
    realCost: number;
  }

  export interface PlanOperator {
    operator: string;
    name?: string;
    estimatedRows: number;
    cost: number;
    outputOrFilter?: string;
    childOperators?: PlanOperator[];
  }

  export interface PlanDetail extends PlanMeta {
    planDetail: PlanOperator;
  }

  export interface BaseSqlRequestParam {
    namespace: string;
    obtenant: string;
    user?: string;
    database?: string;
    includeInnerSql?: boolean;
    startTime?: number;
    endTime?: number;
  }

  export interface Pagination {
    sortColumn?: string;
    sortOrder?: string;
    pageNum?: number;
    pageSize?: number;
  }

  export interface SqlFilter extends BaseSqlRequestParam, Pagination {
    outputColumns?: string[];
    keyword?: string;
    suspiciousOnly?: boolean;
  }

  export interface SqlRequestStatisticParam
    extends BaseSqlRequestParam,
      Pagination {
    statisticScopes: string[];
  }

  export interface SqlDetailParam extends BaseSqlRequestParam {
    interval: number;
    sqlId: string;
    outputColumns?: string[];
  }

  export interface PlanDetailParam extends PlanIdentity {
    namespace: string;
    obtenant: string;
  }

  export interface SqlMetricsResponse {
    successful: boolean;
    data: SqlMetricMetaCategory[];
  }

  export interface SqlStatsResponse {
    successful: boolean;
    data: {
      items: SqlInfo[];
      totalCount: number;
    };
  }

  export interface RequestStatisticsResponse {
    successful: boolean;
    data: RequestStatisticInfo[];
  }

  export interface SqlDetailResponse {
    successful: boolean;
    data: SqlDetailedInfo;
  }

  export interface PlanDetailResponse {
    successful: boolean;
    data: PlanDetail;
  }
}
