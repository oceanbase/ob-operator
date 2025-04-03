import { obcluster } from '@/api';
import CustomTooltip from '@/components/CustomTooltip';
import IconTip from '@/components/IconTip';
import { getColumnSearchProps } from '@/utils/component';
import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Button, Card, Col, message, Row, Space, Table, Tag } from 'antd';
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

  const {
    data: listOBClusterParameters,
    loading,
    refresh,
  } = useRequest(obcluster.listOBClusterParameters, {
    defaultParams: [ns, name],
  });

  const parametersData = listOBClusterParameters?.data;
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

  const statusList = [
    {
      label: (
        <Tag color={'green'}>
          {intl.formatMessage({
            id: 'src.pages.Cluster.Detail.Overview.D5CCD27D',
            defaultMessage: '已匹配',
          })}
        </Tag>
      ),

      value: 'matched',
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

      value: 'notMatched',
    },
    {
      label: '/',
      value: '',
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
      title: intl.formatMessage({
        id: 'src.pages.Cluster.Detail.Parameters.B46C4AD6',
        defaultMessage: '当前值',
      }),
      dataIndex: 'value',
      width: 160,
      render: (text: string, record) => {
        const values = record?.values;

        const singleValue = values?.map((item) => item.value);
        const MultipleValue = values?.map(
          (item) => `${item.value} {${item.metasStr}}`,
        );
        const content = values?.length !== 1 ? MultipleValue : singleValue;
        const tooltip = values?.map((item) => (
          <div>{`${item.value} {${item.metasStr}}`}</div>
        ));
        return (
          <>
            {content?.join('') ? (
              <CustomTooltip
                text={content}
                tooltipTitle={values?.length !== 1 ? tooltip : content}
                width={150}
              />
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
      width: 300,
      render: (text) => {
        return <CustomTooltip text={text} width={290} />;
      },
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Cluster.Detail.Overview.4FCF90AF',
        defaultMessage: '托管 operator',
      }),
      width: 140,
      dataIndex: 'isManagedByOperator',
      filters: controlParameters.map(({ label, value }) => ({
        text: label,
        value,
      })),
      onFilter: (value: any, record) => {
        return record?.isManagedByOperator === value;
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

      dataIndex: 'status',
      width: 100,
      filters: statusList.map(({ label, value }) => ({
        text: label,
        value,
      })),
      onFilter: (value: any, record) => {
        return record?.status === value;
      },
      render: (text) => {
        const content = statusList?.find((item) => item.value === text)?.label;

        return !text ? '/' : <span>{content}</span>;
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

        return (
          <Space size={1}>
            <Button
              type="link"
              onClick={() => {
                setIsDrawerOpen(true);
                setParametersRecord({
                  ...record,
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
          refresh();
        }}
        initialValues={parametersRecord}
        name={name}
        namespace={ns}
      />
    </PageContainer>
  );
}
