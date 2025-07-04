import { inspection } from '@/api';
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
import { sortByString } from '@oceanbase/util';
import { useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { useState } from 'react';
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
  const { id } = params;

  const { data: getInspectionReport } = useRequest(
    inspection.getInspectionReport,
    {
      defaultParams: [
        {
          id,
        },
      ],
      //   ready: !!id,
      manual: true,
    },
  );

  const {
    resultStatistics,
    resultDetail,
    obCluster,
    finishTime,
    startTime,
    scenario,
  } = getInspectionReport?.data || {};
  const { criticalItems, failedItems, moderateItems } = resultDetail || {};
  const { criticalCount, failedCount, moderateCount } = resultStatistics || {};

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

  const CollapseDescriptions: React.FC<CollapseContentProps> = ({
    inspectionItem,
  }) => {
    const columns = [
      {
        title: '巡检项',
        dataIndex: 'name',

        onCell: (record) => {
          const targetItem = showCellList.find(
            (item) => item.showCellIndex === record._index,
          );

          if (targetItem) {
            return { rowSpan: targetItem.count };
          }

          return { rowSpan: 0 };
        },
        render: (text) => {
          return <span>{text}</span>;
        },
      },

      {
        title: '巡检结果',
        dataIndex: 'results',
        render: (text) => {
          return (
            <>
              {text?.map((item) => {
                <div>{item}</div>;
              })}
            </>
          );
        },
      },
    ];

    // 表格渲染的 rules 字段需要从 inspectionItem 头上取到
    const dataSource = (inspectionItem.dataSource || []).map((item) =>
      Object.assign(item, { rules: inspectionItem.rules }),
    );

    // 对 target 进行排序，相同的放到一块
    const realDataSource = dataSource
      .sort((a, b) => {
        return sortByString(a, b, 'target') ? 1 : -1;
      })
      .map((item, index) => {
        return {
          // 增加 _index 做 ID 使用，用来判断是否此列进行需要进行合并
          _index: index,
          ...item,
        };
      });

    const arrChage = (arr: any[], num: number) => {
      const newArr = [];
      while (arr.length > 0) {
        newArr.push(arr.splice(0, num));
      }
      return newArr;
    };

    const showCellList: any[] = [];

    // 因为表格存在分页，所以要转换为二维数组，每一页进行对比，将同一个 target 的列进行合并
    arrChage([...realDataSource], 5).forEach((list) => {
      const targetMap = {};
      // 记录下每个不同的 target 第一次出现的位置，并累加得出每个 target 的总数
      list.forEach((data) => {
        const target = data.target as string;

        if (targetMap[target]) {
          // 记录同一个 target 出现次数
          targetMap[target].count += 1;
        } else {
          targetMap[target] = {
            count: 1,
            // 将 index 存下，在 onCell 中和 record 进行对比，看是否需要展示
            showCellIndex: data._index,
          };
        }
      });

      showCellList.push(...Object.values(targetMap));
    });

    return (
      <Table
        size={'small'}
        bordered={true}
        columns={columns}
        dataSource={inspectionItem}
        rowKey={(record) => record?.name}
        // pagination={{
        //   // ...PAGINATION_OPTION_5,
        //   // 没有数据或只有一页数据时隐藏分页栏
        //   hideOnSinglePage: true,
        //   showSizeChanger: false,
        // }}
      />
    );
  };

  // 风险汇总报告
  const RiskCollapseContent: React.FC<any> = () => {
    const resultList: CollapseProps['items'] = [
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
        {resultList.map(({ key, label, children }) => {
          return (
            <Collapse
              defaultActiveKey={key}
              key={key}
              style={{
                marginTop: '16px',
              }}
            >
              <Panel header={label} key={key}>
                {children?.length > 0 ? (
                  children?.map((item, index) => {
                    return (
                      <>
                        <CollapseDescriptions inspectionItem={item} />
                        {index !== children?.length - 1 && <Divider />}
                      </>
                    );
                  })
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

  // 巡检项汇总报告
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
                {`${obCluster?.namespace}/${obCluster?.name}`}
              </Descriptions.Item>
              <Descriptions.Item label={'巡检场景'}>
                {scenario === 'basic' ? '基础巡检' : '性能巡检'}
              </Descriptions.Item>
              <Descriptions.Item label={'开始时间'}>
                {startTime}
              </Descriptions.Item>
              <Descriptions.Item label={'结束时间'}>
                {finishTime}
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
                  <div>失败</div>
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
  );
};
export default Report;
