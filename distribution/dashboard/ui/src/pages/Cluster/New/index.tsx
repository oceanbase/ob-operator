import { getStorageClasses } from '@/services';
import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { useNavigate,history } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Button, Form, Row,message } from 'antd';
import { useState } from 'react';

import BackUp from './BackUp';
import BasicInfo from './BasicInfo';
import Monitor from './Monitor';
import Observer from './Observer';
import Parameters from './Parameters';
import Topo from './Topo';
import { createObclusterReq } from '@/services';
import { encryptText, usePublicKey } from '@/hook/usePublicKey'

export const SuffixSelector = <div>GB</div>;
export default function New() {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [passwordVal, setPasswordVal] = useState('');
  const { data: storageClasses } = useRequest(getStorageClasses);
  const publicKey = usePublicKey();

  const onFinish = async (values: any) => {
    values.clusterId = new Date().getTime() % 4294901759;
    values.rootPassword = encryptText(values.rootPassword, publicKey) as string;
    const res = await createObclusterReq(values);
    if (res.successful) {
      message.success(res.message, 3);
      form.resetFields();
      setPasswordVal('');
      history.back()
    }
  };
  const initialValues = {
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
    // observer:{
    //   resource:{
    //     cpu:2,
    //     memory:10
    //   },
    //   storage:{
    //     data:{
    //       size:30
    //     },
    //     log:{
    //       size:30
    //     },
    //     redoLog:{
    //       size:30
    //     }
    //   }
    // },
    // monitor:{
    //   resource:{
    //     cpu:1,
    //     memory:1
    //   }
    // }
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
        <Button key="submit" onClick={() => form.submit()}>
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
          <BasicInfo
            passwordVal={passwordVal}
            setPasswordVal={setPasswordVal}
            form={form}
          />
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
