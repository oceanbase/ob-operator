import { inspection } from '@/api';
import { PageContainer } from '@ant-design/pro-components';
import {
  Card,
  Col,
  Collapse,
  Descriptions,
  Divider,
  Empty,
  Row,
  Table,
  token,
} from '@oceanbase/design';
import { formatTime } from '@oceanbase/util';
import { history, useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { useEffect, useState } from 'react';
import styles from './Result.less';

const { Panel } = Collapse;

interface Props {
  inspectionRuleResult: API.SuccessResponse_InspectionReport_;
}

interface CollapseContentProps {
  inspectionItem: any;
}

const Report: React.FC<Props> = () => {
  const [activeTabKey, setActiveTabKey] = useState('risk');

  const params = useParams();
  const { namespace, name } = params;

  const {
    data: inspectionReportData,
    run: fetchReport,
    loading,
    refresh,
  } = useRequest(inspection.getInspectionReport, {
    manual: true,
  });

  const getreport = inspectionReportData?.data || {};

  useEffect(() => {
    if (namespace && name) {
      fetchReport(namespace, name);
    }
  }, [namespace, name]);

  const { criticalItems, failedItems, moderateItems } =
    getreport?.resultDetail || {};
  const { criticalCount, failedCount, moderateCount } =
    getreport?.resultStatistics || {};

  const totalCount = criticalCount + failedCount + moderateCount;

  const tabList = [
    {
      key: 'risk',
      tab: '结果概览',
    },
    {
      key: 'item',
      tab: '全部结果',
    },
  ];

  const CollapseDescriptions: React.FC<
    CollapseContentProps & { tabKey?: string }
  > = ({ inspectionItem }) => {
    const columns = [
      {
        title: '巡检项',
        dataIndex: 'name',
        render: (text: any) => {
          return <span>{text || '-'}</span>;
        },
      },

      {
        title: '巡检结果',
        dataIndex: 'results',
        render: (text: any) => {
          if (!text || !Array.isArray(text)) {
            return <span>-</span>;
          }
          return (
            <>
              {text.map((item: any, index: number) => {
                const red = item?.includes('[critical]');
                return (
                  <div
                    key={index}
                    style={{ color: red ? 'rgba(166,29,36,1)' : 'black' }}
                  >
                    {item || '-'}
                  </div>
                );
              })}
            </>
          );
        },
      },
    ];

    // 确保数据源是数组格式
    const dataSource = Array.isArray(inspectionItem)
      ? inspectionItem
      : [inspectionItem];

    // 添加数据验证
    if (!dataSource || dataSource.length === 0) {
      return <Empty description="暂无数据" />;
    }

    try {
      return (
        <Table
          size={'small'}
          bordered={true}
          columns={columns}
          dataSource={dataSource}
          rowKey={(record: any) =>
            record?.name || record?.id || Math.random().toString()
          }
          pagination={false}
        />
      );
    } catch (error) {
      console.error('Table渲染错误:', error);
      return <div>数据渲染出错，请检查数据格式</div>;
    }
  };

  // 风险汇总报告
  const RiskCollapseContent: React.FC<any> = () => {
    const resultList: any[] = [
      {
        key: 'critical',
        label: '高风险',
        children: criticalItems,
      },
      {
        key: 'moderate',
        label: '中风险',
        children: moderateItems,
      },
      {
        key: 'failed',
        label: '失败',
        children: failedItems,
      },
    ];

    return (
      <>
        {resultList?.map(({ key, label, children }) => {
          return (
            <Collapse
              // 默认只摊开高风险
              defaultActiveKey={'critical'}
              key={key}
              style={{
                marginTop: '16px',
              }}
            >
              <Panel header={label} key={key}>
                {children?.length > 0 ? (
                  <CollapseDescriptions inspectionItem={children} />
                ) : (
                  <Empty
                    image={Empty.PRESENTED_IMAGE_SIMPLE}
                    description={<span>{`无${label}目标对象`}</span>}
                  />
                )}
              </Panel>
            </Collapse>
          );
        })}
      </>
    );
  };

  // 汇总报告
  const ItemCollapseContent: React.FC<any> = () => {
    const data = [
      ...(criticalItems || []),
      ...(failedItems || []),
      ...(moderateItems || []),
    ];

    return (
      <>
        {data.length > 0 ? (
          <CollapseDescriptions inspectionItem={data} />
        ) : (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description={<span>无目标巡检项</span>}
          />
        )}
      </>
    );
  };

  const contentList: Record<string, React.ReactNode> = {
    risk: <RiskCollapseContent />,
    item: <ItemCollapseContent />,
  };

  return (
    <PageContainer
      ghost={true}
      loading={loading}
      header={{
        title: '巡检报告',
        onBack: () => {
          history.back();
        },
        reload: {
          spin: loading,
          onClick: () => {
            refresh();
          },
        },
      }}
    >
      <div className={styles.container}>
        <Row gutter={[16, 16]}>
          <Col span={8}>
            <Card
              bordered={false}
              title={'基本信息'}
              style={{
                height: '250px',
              }}
            >
              <Descriptions column={1}>
                <Descriptions.Item label={'巡检对象'}>
                  {`${namespace}/${getreport?.obCluster?.clusterName}`}
                </Descriptions.Item>
                <Descriptions.Item label={'巡检场景'}>
                  {getreport?.scenario === 'basic' ? '基础巡检' : '性能巡检'}
                </Descriptions.Item>
                <Descriptions.Item label={'开始时间'}>
                  {formatTime(getreport?.startTime)}
                </Descriptions.Item>
                <Descriptions.Item label={'结束时间'}>
                  {formatTime(getreport?.finishTime)}
                </Descriptions.Item>
              </Descriptions>
            </Card>
          </Col>
          <Col span={16}>
            <Card
              title={'巡检结果概览'}
              bordered={false}
              style={{
                height: '250px',
              }}
            >
              <Row
                style={{
                  textAlign: 'center',
                  paddingTop: '20px',
                }}
              >
                <Col span={5}>
                  <span>
                    <div>
                      <div>总巡检结果</div>
                      <div
                        style={{
                          fontSize: '38px',
                        }}
                      >
                        {totalCount || 0}
                      </div>
                    </div>
                  </span>
                </Col>
                <Col span={2}>
                  <Divider
                    type="vertical"
                    style={{
                      height: '50px',
                      marginTop: '10px',
                    }}
                  />
                </Col>
                <Col span={5}>
                  <span>
                    <div>高风险结果</div>
                    <div
                      style={{
                        color: 'rgba(166,29,36,1)',
                        fontSize: '38px',
                        cursor: 'pointer',
                      }}
                    >
                      {criticalCount || 0}
                    </div>
                  </span>
                </Col>
                <Col span={6}>
                  <span>
                    <div>中风险结果</div>
                    <div
                      style={{
                        color: token.colorWarning,
                        fontSize: '38px',
                        cursor: 'pointer',
                      }}
                    >
                      {moderateCount || 0}
                    </div>
                  </span>
                </Col>
                <Col span={6}>
                  <span>
                    <div>失败结果</div>
                    <div
                      style={{
                        color: token.colorError,
                        fontSize: '38px',
                        cursor: 'pointer',
                      }}
                    >
                      {failedCount || 0}
                    </div>
                  </span>
                </Col>
              </Row>
            </Card>
          </Col>
          <Col span={24}>
            <Card
              bordered={false}
              tabList={tabList}
              activeTabKey={activeTabKey}
              onTabChange={(key) => {
                setActiveTabKey(key);
              }}
              bodyStyle={{ paddingTop: 8 }}
            >
              {contentList[activeTabKey]}
            </Card>
          </Col>
        </Row>
      </div>
    </PageContainer>
  );
};

export default Report;
