import { obcluster } from '@/api';
import type { ParamPatchOBClusterParam } from '@/api/generated';
import { CustomFormItem } from '@/components/CustomFormItem';
import InputNumber from '@/components/InputNumber';
import SelectWithTooltip from '@/components/SelectWithTooltip';
import { SUFFIX_UNIT } from '@/constants';
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
  deploymentMode?: string;
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

const ResourceDrawer: React.FC<
  ParametersModalProps & { resource?: { cpu?: number; memory?: number } }
> = ({
  visible,
  onCancel,
  initialValues,
  name,
  namespace,
  onSuccess,
  resource,
  deploymentMode,
}) => {
  const isSharedStorage = deploymentMode === 'shared_storage';
  const [form] = Form.useForm<ParamPatchOBClusterParam>();
  const { validateFields, setFieldValue, resetFields } = form;

  useEffect(() => {
    const data: Record<string, any> = {};
    const log: Record<string, any> = {};
    const redoLog: Record<string, any> = {};

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

    setFieldValue(['storage'], {
      data,
      log,
      ...(isSharedStorage ? {} : { redoLog }),
    });

    // 设置 CPU 和 Memory 的默认值
    if (resource) {
      if (resource.cpu) {
        setFieldValue(['resource', 'cpu'], resource.cpu);
      }
      if (resource.memory) {
        // memory 可能是以字节为单位，需要转换为 GB
        const memoryInGB =
          typeof resource.memory === 'number'
            ? resource.memory / (1 << 30)
            : resource.memory;
        setFieldValue(['resource', 'memory'], memoryInGB);
      }
    }
  }, [initialValues, resource, setFieldValue, isSharedStorage]);

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
      title={intl.formatMessage({
        id: 'src.pages.Cluster.Detail.Overview.41F76901',
        defaultMessage: '节点资源编辑',
      })}
      open={visible}
      destroyOnClose
      onClose={() => {
        onCancel();
        resetFields();
      }}
      width={520}
      footer={
        <Space>
          <Button
            onClick={() => {
              onCancel();
              resetFields();
            }}
          >
            {intl.formatMessage({
              id: 'src.pages.Cluster.Detail.Overview.3B8C3AE9',
              defaultMessage: '取消',
            })}
          </Button>
          <Button
            type="primary"
            loading={loading}
            onClick={() => {
              validateFields().then((value) => {
                if (isSharedStorage && value.storage)
                  delete value.storage.redoLog;
                patchOBCluster(
                  namespace,
                  name,
                  value,
                  intl.formatMessage({
                    id: 'src.pages.Cluster.Detail.Overview.DBF1120A',
                    defaultMessage: '节点资源编辑成功',
                  }),
                );
              });
            }}
          >
            {intl.formatMessage({
              id: 'src.pages.Cluster.Detail.Overview.AC4C9FB4',
              defaultMessage: '确定',
            })}
          </Button>
        </Space>
      }
    >
      <Form form={form} layout="vertical">
        <Row gutter={16}>
          <Col span={24}>
            <p style={fontStyle}>
              {intl.formatMessage({
                id: 'src.pages.Cluster.Detail.Overview.70C825D8',
                defaultMessage: '计算资源',
              })}
            </p>
            <CustomFormItem label="CPU" name={['resource', 'cpu']}>
              <InputNumber
                style={{ width: '180px' }}
                placeholder={intl.formatMessage({
                  id: 'OBDashboard.Cluster.New.Observer.PleaseEnter',
                  defaultMessage: '请输入',
                })}
              />
            </CustomFormItem>
            <CustomFormItem label="Memory" name={['resource', 'memory']}>
              <InputNumber
                addonAfter={SUFFIX_UNIT}
                placeholder={intl.formatMessage({
                  id: 'OBDashboard.Cluster.New.Observer.PleaseEnter',
                  defaultMessage: '请输入',
                })}
              />
            </CustomFormItem>
            <p style={fontStyle}>
              {intl.formatMessage({
                id: 'src.pages.Cluster.Detail.Overview.77C825D8',
                defaultMessage: '存储资源',
              })}
            </p>
            <div style={{ display: 'flex', justifyContent: 'flex-start' }}>
              <CustomFormItem
                style={{ marginRight: '8px' }}
                label="size"
                name={['storage', 'data', 'size']}
              >
                <InputNumber
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
            <p style={fontStyle}>
              {intl.formatMessage({
                id: 'src.pages.Cluster.Detail.Overview.BB0D5386',
                defaultMessage: '日志',
              })}
            </p>
            <div style={{ display: 'flex', justifyContent: 'flex-start' }}>
              <CustomFormItem
                style={{ marginRight: '8px' }}
                label="size"
                name={['storage', 'log', 'size']}
              >
                <InputNumber
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
          {!isSharedStorage && (
            <Col span={24}>
              <p style={fontStyle}>redoLog</p>
              <div style={{ display: 'flex', justifyContent: 'flex-start' }}>
                <CustomFormItem
                  style={{ marginRight: '8px' }}
                  label="size"
                  name={['storage', 'redoLog', 'size']}
                >
                  <InputNumber
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
          )}
        </Row>
      </Form>
    </Drawer>
  );
};

export default ResourceDrawer;
