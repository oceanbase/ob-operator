import { getStorageClasses } from '@/services';
import { createClusterReportWrap } from '@/services/reportRequest/clusterReportReq';
import { strTrim } from '@/utils/helper';
import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { useNavigate } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Button, Col, Form, Row, message } from 'antd';
import { useState } from 'react';

import { encryptText, usePublicKey } from '@/hook/usePublicKey';
import BackUp from './BackUp';
import BasicInfo from './BasicInfo';
import Monitor from './Monitor';
import Observer from './Observer';
import Parameters from './Parameters';
import Topo from './Topo';

export default function New() {
  const navigate = useNavigate();
  const [form] = Form.useForm<API.CreateClusterData>();
  const [passwordVal, setPasswordVal] = useState('');
  const { data: storageClassesRes } = useRequest(getStorageClasses, {
    onSuccess: ({ successful, data }) => {
      if (successful && data.length === 1) {
        const { value } = data[0];
        form.setFieldValue(['observer', 'storage'], {
          data: {
            storageClass: value,
          },
          log: {
            storageClass: value,
          },
          redoLog: {
            storageClass: value,
          },
        });
      }
    },
  });
  const publicKey = usePublicKey();
  const storageClasses = storageClassesRes?.data;
  const onFinish = async (values: API.CreateClusterData) => {
    values.clusterId = new Date().getTime() % 4294901759;
    values.rootPassword = encryptText(values.rootPassword, publicKey) as string;

    const res = await createClusterReportWrap({ ...strTrim(values) });
    if (res.successful) {
      message.success(res.message, 3);
      form.resetFields();
      setPasswordVal('');
      history.back();
    }
  };
  const initialValues = {
    mode: 'NORMAL',
    topology: [
      {
        zone: 'zone1',
        replicas: 1,
      },
      {
        zone: 'zone2',
        replicas: 1,
      },
      {
        zone: 'zone3',
        replicas: 1,
      },
    ],
  };
  return (
    <PageContainer
      header={{
        title: intl.formatMessage({
          id: 'dashboard.Cluster.New.CreateACluster',
          defaultMessage: '创建集群',
        }),
        onBack: () => {
          navigate('/cluster');
        },
      }}
      footer={[
        <Button onClick={() => navigate('/cluster')} key="cancel">
          {intl.formatMessage({
            id: 'dashboard.Cluster.New.Cancel',
            defaultMessage: '取消',
          })}
        </Button>,
        <Button type="primary" key="submit" onClick={() => form.submit()}>
          {intl.formatMessage({
            id: 'dashboard.Cluster.New.Submit',
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
              setPasswordVal={setPasswordVal}
              form={form}
            />
          </Col>
          <Topo form={form} />
          <Observer storageClasses={storageClasses} form={form} />
          <Monitor />
          <Parameters />
          <BackUp />
        </Row>
      </Form>
    </PageContainer>
  );
}
