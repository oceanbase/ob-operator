import { obcluster } from '@/api';
import { CustomFormItem } from '@/components/CustomFormItem';
import InputNumber from '@/components/InputNumber';
import SelectWithTooltip from '@/components/SelectWithTooltip';
import { MINIMAL_CONFIG, SUFFIX_UNIT } from '@/constants';
import { getStorageClasses } from '@/services';
import { intl } from '@/utils/intl';
import { useRequest } from 'ahooks';
import { Button, Col, Drawer, Form, Row, Space, message } from 'antd';
import React, { useEffect } from 'react';

export interface ParametersModalProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
  initialValues: any[];
  name: string;
  namespace: string;
}

export const TooltipItemContent = ({ item }) => {
  return (
    <ul style={{ margin: 0, padding: '10px' }}>
      {item.toolTipData.map((data: any) => {
        const key = Object.keys(data)[0];
        if (typeof data[key] === 'string') {
          return (
            <li style={{ listStyle: 'none' }} key={key}>
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                }}
              >
                <p>{key}：</p>
                <p>{data[key]}</p>
              </div>
            </li>
          );
        } else {
          const value = JSON.stringify(data[key]) || String(data[key]);
          return (
            <li style={{ listStyle: 'none' }} key={key}>
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                }}
              >
                <p>{key}：</p>
                <p>{value}</p>
              </div>
            </li>
          );
        }
      })}
    </ul>
  );
};

const ResourceDrawer: React.FC<ParametersModalProps> = ({
  visible,
  onCancel,
  initialValues,
  name,
  namespace,
  onSuccess,
}) => {
  const [form] = Form.useForm<API.CreateClusterData>();
  const { validateFields } = form;

  useEffect(() => {
    const data = {};
    const log = {};
    const redoLog = {};

    initialValues?.forEach((item) => {
      if (item.type === 'data') {
        data[item.label] = item.value;
      }
      if (item.type === 'log') {
        log[item.label] = item.value;
      }
      if (item.type === 'redoLog') {
        redoLog[item.label] = item.value;
      }
    });

    form.setFieldValue(['storage'], {
      data,
      log,
      redoLog,
    });
  }, [initialValues]);

  const { data: storageClassesRes } = useRequest(getStorageClasses, {});

  const storageClasses = storageClassesRes?.data;

  const { runAsync: patchOBCluster, loading } = useRequest(
    obcluster.patchOBCluster,
    {
      manual: true,
      onSuccess: (res) => {
        if (res.successful) {
          message.success(
            intl.formatMessage({
              id: 'src.pages.Cluster.Detail.Overview.E908AA54',
              defaultMessage: '编辑参数已成功',
            }),
          );
          onSuccess();
        }
      },
    },
  );

  const fontStyle: React.CSSProperties = {
    fontWeight: 600,
  };
  return (
    <Drawer
      title={'存储资源编辑'}
      open={visible}
      destroyOnClose
      onClose={() => onCancel()}
      width={520}
      footer={
        <Space>
          <Button onClick={onCancel}>取消</Button>
          <Button
            type="primary"
            loading={loading}
            onClick={() => {
              validateFields().then((value) => {
                patchOBCluster(name, namespace, value, `存储资源编辑成功`);
              });
            }}
          >
            确定
          </Button>
        </Space>
      }
    >
      <Form form={form} layout="vertical">
        <Row gutter={16}>
          <Col span={24}>
            <p style={fontStyle}>数据</p>
            <div style={{ display: 'flex', justifyContent: 'flex-start' }}>
              <CustomFormItem
                style={{ marginRight: '8px' }}
                label="size"
                name={['storage', 'data', 'size']}
              >
                <InputNumber
                  min={MINIMAL_CONFIG.data}
                  addonAfter={SUFFIX_UNIT}
                  placeholder={intl.formatMessage({
                    id: 'OBDashboard.Cluster.New.Observer.PleaseEnter',
                    defaultMessage: '请输入',
                  })}
                />
              </CustomFormItem>
              <CustomFormItem
                label="storageClass"
                name={['storage', 'data', 'storageClass']}
              >
                {storageClasses && (
                  <SelectWithTooltip
                    name={['storage', 'data', 'storageClass']}
                    form={form}
                    selectList={storageClasses}
                    TooltipItemContent={TooltipItemContent}
                  />
                )}
              </CustomFormItem>
            </div>
          </Col>
          <Col span={24}>
            <p style={fontStyle}>日志</p>
            <div style={{ display: 'flex', justifyContent: 'flex-start' }}>
              <CustomFormItem
                style={{ marginRight: '8px' }}
                label="size"
                name={['storage', 'log', 'size']}
              >
                <InputNumber
                  min={MINIMAL_CONFIG.log}
                  addonAfter={SUFFIX_UNIT}
                  placeholder={intl.formatMessage({
                    id: 'OBDashboard.Cluster.New.Observer.PleaseEnter',
                    defaultMessage: '请输入',
                  })}
                />
              </CustomFormItem>
              <CustomFormItem
                label="storageClass"
                name={['storage', 'log', 'storageClass']}
              >
                {storageClasses && (
                  <SelectWithTooltip
                    form={form}
                    name={['storage', 'log', 'storageClass']}
                    selectList={storageClasses}
                    TooltipItemContent={TooltipItemContent}
                  />
                )}
              </CustomFormItem>
            </div>
          </Col>
          <Col span={24}>
            <p style={fontStyle}>redoLog</p>
            <div style={{ display: 'flex', justifyContent: 'flex-start' }}>
              <CustomFormItem
                style={{ marginRight: '8px' }}
                label="size"
                name={['storage', 'redoLog', 'size']}
              >
                <InputNumber
                  min={MINIMAL_CONFIG.redoLog}
                  addonAfter={SUFFIX_UNIT}
                  placeholder={intl.formatMessage({
                    id: 'OBDashboard.Cluster.New.Observer.PleaseEnter',
                    defaultMessage: '请输入',
                  })}
                />
              </CustomFormItem>
              <CustomFormItem
                label="storageClass"
                validateTrigger="onBlur"
                name={['storage', 'redoLog', 'storageClass']}
              >
                {storageClasses && (
                  <SelectWithTooltip
                    form={form}
                    name={['storage', 'redoLog', 'storageClass']}
                    selectList={storageClasses}
                    TooltipItemContent={TooltipItemContent}
                  />
                )}
              </CustomFormItem>
            </div>
          </Col>
        </Row>
      </Form>
    </Drawer>
  );
};

export default ResourceDrawer;
