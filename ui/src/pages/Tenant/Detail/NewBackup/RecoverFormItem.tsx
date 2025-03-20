import { TZ_NAME_REG } from '@/constants';
import { resourceNameRule } from '@/constants/rules';
import {
  getEssentialParameters as getEssentialParametersReq,
  getSimpleClusterList,
} from '@/services';
import { intl } from '@/utils/intl';
import { useDeepCompareEffect, useRequest, useUpdateEffect } from 'ahooks';
import {
  Button,
  Col,
  DatePicker,
  Form,
  Input,
  Row,
  Select,
  TimePicker,
} from 'antd';
import dayjs from 'dayjs';
import ResourcePools from '../../New/ResourcePools';

export default function RecoverFormItem({
  form,
  clusterList,
  selectClusterId,
  setSelectClusterId,
  setClusterList,
  onFinish,
  type,
}) {
  useRequest(getSimpleClusterList, {
    onSuccess: ({ successful, data }) => {
      if (successful) {
        data?.forEach((cluster) => {
          cluster?.topology?.forEach((zone) => {
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

  const essentialParameter = essentialParameterRes?.data;

  useUpdateEffect(() => {
    const { name, namespace } =
      clusterList.find((cluster) => cluster.id === selectClusterId) || {};
    if (name && namespace) {
      getEssentialParameters({
        ns: namespace,
        name,
      });
    }
  }, [selectClusterId]);

  useDeepCompareEffect(() => {
    if (clusterList) {
      const cluster = clusterList.find(
        (cluster) => cluster.id === selectClusterId,
      );
      cluster?.topology.forEach((zone) => {
        form.setFieldValue(['pools', zone.zone, 'checked'], zone.checked);
      });
    }
  }, [clusterList]);

  const clusterOptions = clusterList
    .filter((cluster) => cluster.status !== 'failed')
    .map((cluster) => ({
      value: cluster.id,
      label: cluster.name,
      status: cluster.status,
    }));

  return (
    <Row gutter={[16, 32]}>
      <Col span={12}>
        <Form.Item
          name={['source', 'restore', 'until', 'date']}
          label={'恢复日期'}
          rules={[
            {
              required: true,
              message: '请选择恢复日期',
            },
          ]}
        >
          <DatePicker style={{ width: '100%' }} placeholder={'请选择'} />
        </Form.Item>
      </Col>
      <Col span={12}>
        <Form.Item
          name={['source', 'restore', 'until', 'time']}
          label={'时分秒'}
          rules={[
            {
              required: true,
              message: '请选择时间',
            },
          ]}
        >
          <TimePicker
            style={{ width: '100%' }}
            defaultOpenValue={dayjs('00:00:00', 'HH:mm:ss')}
          />
        </Form.Item>
      </Col>
      <Col span={8}>
        <Form.Item
          name={['obcluster']}
          label={'集群'}
          rules={[
            {
              required: true,
              message: '请选择集群',
            },
          ]}
        >
          <Select
            placeholder={intl.formatMessage({
              id: 'Dashboard.Detail.NewBackup.PleaseSelect',
              defaultMessage: '请选择',
            })}
            onChange={(value) => setSelectClusterId(value)}
            optionLabelProp="selectLabel"
            options={clusterOptions.map((option) => ({
              value: option.value,
              selectLabel: option.label,
              disabled: option.status !== 'running',
              label: (
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                  }}
                >
                  <span>{option.label}</span>
                  <span>{option.status}</span>
                </div>
              ),
            }))}
          />
        </Form.Item>
      </Col>
      <Col span={8}>
        <Form.Item
          name={['name']}
          label={'资源名'}
          validateFirst
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'Dashboard.Tenant.New.BasicInfo.EnterAResourceName',
                defaultMessage: '请输入资源名',
              }),
            },
            {
              pattern: /\D/,
              message: intl.formatMessage({
                id: 'Dashboard.Tenant.New.BasicInfo.ResourceNamesCannotUsePure',
                defaultMessage: '资源名不能使用纯数字',
              }),
            },
            resourceNameRule,
          ]}
        >
          <Input placeholder="请输入" />
        </Form.Item>
      </Col>
      <Col span={8}>
        <Form.Item
          name={['tenantName']}
          label={'租户名'}
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'Dashboard.Tenant.New.BasicInfo.EnterATenantName',
                defaultMessage: '请输入租户名',
              }),
            },
            {
              pattern: TZ_NAME_REG,
              message: intl.formatMessage({
                id: 'Dashboard.Tenant.New.BasicInfo.TheFirstCharacterMustBe',
                defaultMessage: '首字符必须是字母或者下划线，不能包含 -',
              }),
            },
          ]}
        >
          <Input placeholder="请输入" />
        </Form.Item>
      </Col>
      <Col span={24}>
        <ResourcePools
          form={form}
          type={'tenantBackup'}
          selectClusterId={selectClusterId}
          essentialParameter={essentialParameter}
          clusterList={clusterList}
          setClusterList={setClusterList}
        />
      </Col>
      {type === 'detail' && (
        <>
          <Col span={22}></Col>
          <Col span={2}>
            <Button type="primary" onClick={() => onFinish()}>
              提交
            </Button>
          </Col>
        </>
      )}
    </Row>
  );
}
