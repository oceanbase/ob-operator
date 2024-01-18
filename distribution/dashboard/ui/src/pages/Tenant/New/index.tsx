import { usePublicKey } from '@/hook/usePublicKey';
import { getSimpleClusterList } from '@/services';
import { createTenant } from '@/services/tenant';
import { PageContainer } from '@ant-design/pro-components';
import { useNavigate } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Button, Col, Form, Row, message } from 'antd';
import { useState } from 'react';
import { formatNewTenantForm } from '../helper';
import BasicInfo from './BasicInfo';
import ResourcePools from './ResourcePools';
import TenantSource from './TenantSource';
// 新建租户页
export default function New() {
  const navigate = useNavigate();
  const publicKey = usePublicKey();
  const [form] = Form.useForm();
  const [passwordVal, setPasswordVal] = useState('');
  const [selectClusterId, setSelectClusterId] = useState<number>();

  const { data: clusterList = [] } = useRequest(getSimpleClusterList);

  //选中的集群资源名
  const clusterName = clusterList.filter(
    (cluster) => cluster.clusterId === selectClusterId,
  )[0]?.name;
  //namespace最后提交的时候确认
  const onFinish = async (values: any) => {
    console.log('values', values);
    const ns = clusterList.filter(
      (cluster) => cluster.clusterId === selectClusterId,
    )[0]?.namespace;
    const res = await createTenant({
      ns,
      ...formatNewTenantForm(values, clusterName, publicKey),
    });
    if (res.successful) {
      message.success('创建租户成功', 3);
      form.resetFields();
      history.back()
    }
  };

  const initialValues = {
    connectWhiteList: ['%'],
  };
  return (
    <PageContainer
      header={{
        title: '创建租户',
        onBack: () => {
          navigate('/tenant');
        },
      }}
      footer={[
        <Button onClick={() => navigate('/tenant')} key="cancel">
          取消
        </Button>,
        <Button key="submit" onClick={() => form.submit()}>
          提交
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
