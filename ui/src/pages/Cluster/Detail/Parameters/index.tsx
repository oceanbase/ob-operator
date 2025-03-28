import { obcluster } from '@/api';
import CustomTooltip from '@/components/CustomTooltip';
import IconTip from '@/components/IconTip';
import { getClusterDetailReq } from '@/services';
import { getColumnSearchProps } from '@/utils/component';
import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Button, Card, Col, message, Row, Space, Table, Tag } from 'antd';
import { isEmpty } from 'lodash';
import { useState } from 'react';
import ParametersModal from './ParametersModal';

export default function Parameters() {
  const { ns, name } = useParams();

  const [isDrawerOpen, setIsDrawerOpen] = useState<boolean>(false);
  const [parametersRecord, setParametersRecord] = useState({});

  const { runAsync: patchOBCluster, loading: patchOBClusterloading } =
    useRequest(obcluster.patchOBCluster, {
      manual: true,
      onSuccess: (res) => {
        if (res.successful) {
          message.success(
            intl.formatMessage({
              id: 'src.pages.Cluster.Detail.Overview.FF85D01F',
              defaultMessage: '解除托管已成功',
            }),
          );

          refresh();
        }
      },
    });

  const { data: clusterDetail, refresh: clusterDetailRefresh } = useRequest(
    getClusterDetailReq,
    {},
  );

  const {
    data: listOBClusterParameters,
    loading,
    refresh,
  } = useRequest(obcluster.listOBClusterParameters, {
    defaultParams: [ns, name],
    refreshDeps: [clusterDetail?.status],
  });

  const parameters = clusterDetail?.info?.parameters;
  const getNewData = (data) => {
    const obt = data?.map((element: any) => {
      // 在 obcluster 的 parameters  里面的就是托管给 operator
      const findName = parameters?.find(
        (item: any) => element.name === item.name,
      );

      if (!isEmpty(findName)) {
        return {
          ...element,
          controlParameter: true,
          accordance: findName?.value === findName?.specValue,
        };
      } else if (isEmpty(findName)) {
        return { ...element, controlParameter: false, accordance: 'null' };
      }
    });

    return obt;
  };

  const parametersData = getNewData(listOBClusterParameters?.data);
  const controlParameters = [
    {
      label: intl.formatMessage({
        id: 'src.pages.Cluster.Detail.Overview.403B7E1C',
        defaultMessage: '已托管',
      }),
      value: true,
    },
    {
      label: intl.formatMessage({
        id: 'src.pages.Cluster.Detail.Overview.46B66B3E',
        defaultMessage: '未托管',
      }),
      value: false,
    },
  ];

  const accordanceList = [
    {
      label: (
        <Tag color={'green'}>
          {intl.formatMessage({
            id: 'src.pages.Cluster.Detail.Overview.D5CCD27D',
            defaultMessage: '已匹配',
          })}
        </Tag>
      ),

      value: true,
    },
    {
      label: (
        <Tag color={'gold'}>
          {intl.formatMessage({
            id: 'src.pages.Cluster.Detail.Overview.DF83C06D',
            defaultMessage: '不匹配',
          })}
        </Tag>
      ),

      value: false,
    },
    {
      label: '/',

      value: 'null',
    },
  ];

  const columns = [
    {
      title: intl.formatMessage({
        id: 'src.pages.Cluster.Detail.Overview.E5342F26',
        defaultMessage: '参数名',
      }),
      dataIndex: 'name',
      ...getColumnSearchProps({
        frontEndSearch: true,
        dataIndex: 'name',
      }),
    },
    {
      title: '当前值',
      dataIndex: 'value',
      width: 160,
      render: (text: string, record) => {
        const values = record?.values;

        const singleValue = values?.map((item) => item.value);
        const MultipleValue = values?.map(
          (item) => `${item.value} {${item.metasStr}}`,
        );
        const content = values?.length !== 1 ? MultipleValue : singleValue;

        return (
          <>
            {content?.join('') ? (
              <CustomTooltip text={content} width={150} />
            ) : (
              <span>-</span>
            )}
          </>
        );
      },
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Cluster.Detail.Overview.93A9D19D',
        defaultMessage: '参数说明',
      }),
      dataIndex: 'info',
      width: 200,
      render: (text) => {
        return <CustomTooltip text={text} width={190} />;
      },
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Cluster.Detail.Overview.4FCF90AF',
        defaultMessage: '托管 operator',
      }),
      width: 140,
      dataIndex: 'controlParameter',
      filters: controlParameters.map(({ label, value }) => ({
        text: label,
        value,
      })),
      onFilter: (value: any, record) => {
        return record?.controlParameter === value;
      },
      render: (text: boolean) => {
        return (
          <span>
            {text
              ? intl.formatMessage({
                  id: 'src.pages.Cluster.Detail.Overview.319FA0DB',
                  defaultMessage: '是',
                })
              : intl.formatMessage({
                  id: 'src.pages.Cluster.Detail.Overview.5DD958C7',
                  defaultMessage: '否',
                })}
          </span>
        );
      },
    },
    {
      title: (
        <IconTip
          tip={intl.formatMessage({
            id: 'src.pages.Cluster.Detail.Overview.0B4A3E74',
            defaultMessage: '只有托管 operator 的参数才有状态',
          })}
          content={intl.formatMessage({
            id: 'src.pages.Cluster.Detail.Overview.6AD01A82',
            defaultMessage: '状态',
          })}
        />
      ),

      dataIndex: 'accordance',
      width: 100,
      filters: accordanceList.map(({ label, value }) => ({
        text: label,
        value,
      })),
      onFilter: (value: any, record) => {
        return record?.accordance === value;
      },
      render: (text) => {
        const tagColor = text ? 'green' : 'gold';
        const tagContent = text
          ? intl.formatMessage({
              id: 'src.pages.Cluster.Detail.Overview.9A3A4407',
              defaultMessage: '已匹配',
            })
          : intl.formatMessage({
              id: 'src.pages.Cluster.Detail.Overview.D6588C55',
              defaultMessage: '不匹配',
            });

        return text === 'null' ? '/' : <Tag color={tagColor}>{tagContent}</Tag>;
      },
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Cluster.Detail.Overview.1B9EA477',
        defaultMessage: '操作',
      }),
      dataIndex: 'controlParameter',
      align: 'center',
      render: (text, record) => {
        const disableUnescrow = [
          'memory_limit',
          'datafile_maxsize',
          'datafile_next',
          'enable_syslog_recycle',
          'max_syslog_file_count',
        ];

        const valueContent =
          parameters?.find((item) => item.name === record.name)?.value ||
          record?.value;

        return (
          <Space size={1}>
            <Button
              type="link"
              onClick={() => {
                setIsDrawerOpen(true);
                setParametersRecord({
                  ...record,
                  value: valueContent,
                });
              }}
            >
              {intl.formatMessage({
                id: 'src.pages.Cluster.Detail.Overview.F5A088FB',
                defaultMessage: '编辑',
              })}
            </Button>
            {text && (
              <Button
                type="link"
                disabled={disableUnescrow.some((item) => item === record.name)}
                loading={patchOBClusterloading}
                onClick={() => {
                  patchOBCluster(ns, name, {
                    deletedParameters: [record.name],
                  });
                }}
              >
                {intl.formatMessage({
                  id: 'src.pages.Cluster.Detail.Overview.5FACF7C0',
                  defaultMessage: '解除托管',
                })}
              </Button>
            )}
          </Space>
        );
      },
    },
  ];

  return (
    <PageContainer>
      <Row>
        <Col span={24}>
          <Card
            title={
              <h2 style={{ marginBottom: 0 }}>
                {intl.formatMessage({
                  id: 'src.pages.Cluster.Detail.Overview.BFE7CA02',
                  defaultMessage: '集群参数',
                })}
              </h2>
            }
          >
            <Table
              rowKey="name"
              pagination={{ simple: true }}
              columns={columns}
              loading={loading}
              dataSource={parametersData}
            />
          </Card>
        </Col>
      </Row>
      <ParametersModal
        visible={isDrawerOpen}
        onCancel={() => setIsDrawerOpen(false)}
        onSuccess={() => {
          setIsDrawerOpen(false);
          clusterDetailRefresh();
          refresh();
        }}
        initialValues={parametersRecord}
        name={name}
        namespace={ns}
      />
    </PageContainer>
  );
}
