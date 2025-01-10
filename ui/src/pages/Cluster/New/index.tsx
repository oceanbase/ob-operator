import { getStorageClasses } from '@/services';
import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { useNavigate } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Button, Col, Form, message, Row } from 'antd';
import { useEffect, useState } from 'react';

import { encryptText, usePublicKey } from '@/hook/usePublicKey';
import { createClusterReportWrap } from '@/services/reportRequest/clusterReportReq';
import { strTrim } from '@/utils/helper';
import BackUp from './BackUp';
import BasicInfo from './BasicInfo';
import Monitor from './Monitor';
import Observer from './Observer';
import Parameters from './Parameters';
import Topo from './Topo';

export default function New() {
  const navigate = useNavigate();
  const [form] = Form.useForm<API.CreateClusterData>();
  const [passwordVal, setPasswordVal] = useState<string>('');
  const [pvcValue, setPvcValue] = useState<boolean>(false);
  const [deleteValue, setDeleteValue] = useState<boolean>(false);
  const { data: storageClassesRes, run: fetchStorageClasses } = useRequest(
    getStorageClasses,
    {
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
      manual: true,
    },
  );
  const publicKey = usePublicKey();
  const storageClasses = storageClassesRes?.data;
  const onFinish = async (values: API.CreateClusterData) => {
    values.clusterId = new Date().getTime() % 4294901759;
    values.rootPassword = encryptText(values.rootPassword, publicKey) as string;
    values.deletionProtection = deleteValue;
    values.pvcIndependent = pvcValue;

    const topologyValue = strTrim(values)?.topology?.map((item) => ({
      ...item,
      nodeSelector: undefined,
      affinities:
        item?.affinities?.concat(item?.nodeSelector) ||
        item?.nodeSelector ||
        item?.affinities,
    }));

    console.log('strTrim values', strTrim(values));
    values.topology = topologyValue;
    const res = await createClusterReportWrap({ ...strTrim(values) });
    if (res.successful) {
      message.success(res.message, 3);
      form.resetFields();
      setPasswordVal('');
      setPvcValue(false);
      setDeleteValue(true);
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

  useEffect(() => {
    fetchStorageClasses();
  }, []);
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
              deleteValue={deleteValue}
              setPasswordVal={setPasswordVal}
              setDeleteValue={setDeleteValue}
              form={form}
            />
          </Col>
          <Topo form={form} />
          <Observer
            storageClasses={storageClasses}
            form={form}
            pvcValue={pvcValue}
            setPvcValue={setPvcValue}
          />
          <Monitor />
          <Parameters />
          <BackUp />
        </Row>
      </Form>
    </PageContainer>
  );
}
