import { usePublicKey } from '@/hook/usePublicKey';
import {
  getEssentialParameters as getEssentialParametersReq,
  getSimpleClusterList,
} from '@/services';
import { createTenant } from '@/services/tenant';
import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { useNavigate } from '@umijs/max';
import { useRequest, useUpdateEffect } from 'ahooks';
import { Button, Col, Form, Row, message } from 'antd';
import { useEffect, useState } from 'react';
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
  const [clusterList, setClusterList] = useState<API.SimpleClusterList>([]);
  useRequest(getSimpleClusterList, {
    onSuccess: ({ successful, data }) => {
      if (successful) {
        data.forEach((cluster) => {
          cluster.topology.forEach((zone) => {
            zone.checked = false;
          });
        });

        setClusterList(data);
      }
    },
  });
  const { data: essentialParameterRes, run: getEssentialParameters } =
    useRequest(getEssentialParametersReq, {
      manual: true,
    });
  //Selected cluster resource name
  const clusterName = clusterList.filter(
    (cluster) => cluster.clusterId === selectClusterId,
  )[0]?.name;
  const essentialParameter = essentialParameterRes?.data;

  const onFinish = async (values: any) => {
    const reqData = formatNewTenantForm(values, clusterName, publicKey);
    if (!reqData.pools?.length) {
      message.warning(
        intl.formatMessage({
          id: 'Dashboard.Tenant.New.SelectAtLeastOneZone',
          defaultMessage: '至少选择一个Zone',
        }),
      );
      return;
    }
    
    const ns = clusterList.filter(
      (cluster) => cluster.clusterId === selectClusterId,
    )[0]?.namespace;
    const res = await createTenant({
      namespace: ns,
      ...reqData,
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

  useUpdateEffect(() => {
    const { name, namespace } = clusterList.find(
      (cluster) => cluster.clusterId === selectClusterId,
    );
    if (name && namespace) {
      getEssentialParameters({
        ns: namespace,
        name,
      });
    }
  }, [selectClusterId]);

  useEffect(() => {
    if (clusterList) {
      const cluster = clusterList.find(
        (cluster) => cluster.clusterId === selectClusterId,
      );
      cluster?.topology.forEach((zone) => {
        form.setFieldValue(['pools', zone.zone, 'checked'], zone.checked);
      });
    }
  }, [clusterList]);

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
        <Button type="primary" key="submit" onClick={() => form.submit()}>
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
              essentialParameter={essentialParameter}
              clusterList={clusterList}
              setClusterList={setClusterList}
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
