import { usePublicKey } from '@/hook/usePublicKey';
import { getSimpleClusterList } from '@/services';
import { createTenant } from '@/services/tenant';
import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { useNavigate } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Button, Col, Form, Row, message } from 'antd';
import { useState } from 'react';
import { formatNewTenantForm } from '../helper';
import BasicInfo from './BasicInfo';
import ResourcePools from './ResourcePools';
import TenantSource from './TenantSource';
// New tenant page
export default function New() {
  const navigate = useNavigate();
  const publicKey = usePublicKey();
  const [form] = Form.useForm();
  const [passwordVal, setPasswordVal] = useState('');
  const [selectClusterId, setSelectClusterId] = useState<number>();

  const { data: clusterList = [] } = useRequest(getSimpleClusterList);

  //Selected cluster resource name
  const clusterName = clusterList.filter(
    (cluster) => cluster.clusterId === selectClusterId,
  )[0]?.name;

  const onFinish = async (values: any) => {
    const ns = clusterList.filter(
      (cluster) => cluster.clusterId === selectClusterId,
    )[0]?.namespace;
    const res = await createTenant({
      ns,
      ...formatNewTenantForm(values, clusterName, publicKey),
    });
    if (res.successful) {
      message.success(
        intl.formatMessage({
          id: 'Dashboard.Tenant.New.TenantCreatedSuccessfully',
          defaultMessage: '创建租户成功',
        }),
        3,
      );
      form.resetFields();
      history.back();
    }
  };

  const initialValues = {
    connectWhiteList: ['%'],
  };
  return (
    <PageContainer
      header={{
        title: intl.formatMessage({
          id: 'Dashboard.Tenant.New.CreateATenant',
          defaultMessage: '创建租户',
        }),
        onBack: () => {
          history.back();
        },
      }}
      footer={[
        <Button onClick={() => navigate('/tenant')} key="cancel">
          {intl.formatMessage({
            id: 'Dashboard.Tenant.New.Cancel',
            defaultMessage: '取消',
          })}
        </Button>,
        <Button key="submit" onClick={() => form.submit()}>
          {intl.formatMessage({
            id: 'Dashboard.Tenant.New.Submit',
            defaultMessage: '提交',
          })}
        </Button>,
      ]}
    >
      <Form
        form={form}
        onFinish={onFinish}
        layout="vertical"
        initialValues={initialValues}
        style={{ marginBottom: 56 }}
      >
        <Row gutter={[16, 16]}>
          <Col span={24}>
            <BasicInfo
              passwordVal={passwordVal}
              clusterList={clusterList}
              setSelectClusterId={setSelectClusterId}
              setPasswordVal={setPasswordVal}
              form={form}
            />
          </Col>
          <Col span={24}>
            <ResourcePools
              form={form}
              selectClusterId={selectClusterId}
              clusterList={clusterList}
            />
          </Col>
          <Col span={24}>
            <TenantSource form={form} clusterName={clusterName} />
          </Col>
        </Row>
      </Form>
    </PageContainer>
  );
}
