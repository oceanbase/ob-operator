import { obproxy } from '@/api';
import type { CommonKVPair, CommonResourceSpec } from '@/api/generated';
import { ObproxyPatchOBProxyParam } from '@/api/generated';
import AlertDrawer from '@/components/AlertDrawer';
import { CustomFormItem } from '@/components/CustomFormItem';
import { SERVICE_TYPE, SUFFIX_UNIT } from '@/constants';
import { MIRROR_OBPROXY } from '@/constants/doc';
import { intl } from '@/utils/intl';
import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';
import { useRequest } from 'ahooks';
import type { DrawerProps } from 'antd';
import {
  Button,
  Col,
  Form,
  Input,
  InputNumber,
  Row,
  Select,
  Typography,
} from 'antd';
import { useEffect, useRef } from 'react';
import { isDifferentParams } from '../../helper';

interface ConfigDrawerProps extends DrawerProps {
  onClose: () => void;
  name: string;
  namespace: string;
  image?: string;
  parameters?: CommonKVPair[];
  resource?: CommonResourceSpec;
  serviceType?: string;
  replicas?: number;
}

type FormValue = {
  parameters?: { key: string; value: string }[];
} & ObproxyPatchOBProxyParam;

const { Text } = Typography;

export default function ConfigDrawer({
  onClose,
  name,
  namespace,
  ...props
}: ConfigDrawerProps) {
  const [form] = Form.useForm();
  const preParameters = useRef<CommonKVPair[] | undefined>();
  const { data: listParametersRes } = useRequest(
    obproxy.listOBProxyParameters,
    {
      defaultParams: [namespace, name],
    },
  );
  const listParametersOptions = listParametersRes?.data.map((item) => ({
    label: item.name,
    value: item.name,
    info: item.info,
  }));

  const submit = (values: FormValue) => {
    if (
      !isDifferentParams(values.parameters || [], preParameters.current || [])
    ) {
      delete values.parameters;
    }
  };
  const titleStyle = { fontSize: 14, fontWeight: 600 };

  const labelChange = (label: string, name: number) => {
    const value = listParametersRes?.data?.find(
      (parameter) => parameter.name === label,
    )?.value;
    value && form.setFieldValue(['parameters', name, 'value'], value);
  };
  useEffect(() => {
    preParameters.current = props.parameters;
  }, [props.parameters]);
  return (
    <AlertDrawer
      title="详细配置"
      onSubmit={() => form.submit()}
      destroyOnClose={true}
      onClose={() => onClose()}
      {...props}
    >
      <Form
        initialValues={props}
        form={form}
        onFinish={submit}
        preserve={false}
        layout="vertical"
      >
        <p style={titleStyle}>资源设置</p>
        <CustomFormItem
          label={
            <>
              {intl.formatMessage({
                id: 'Dashboard.Cluster.New.Observer.Image',
                defaultMessage: '镜像',
              })}{' '}
              <a href={MIRROR_OBPROXY} rel="noreferrer" target="_blank">
                {intl.formatMessage({
                  id: 'Dashboard.Cluster.New.Observer.ImageList',
                  defaultMessage: '（镜像列表）',
                })}
              </a>
            </>
          }
          name="image"
          message={intl.formatMessage({
            id: 'Dashboard.Cluster.New.Observer.EnterAnImage',
            defaultMessage: '请输入镜像',
          })}
        >
          <Input
            placeholder={intl.formatMessage({
              id: 'OBDashboard.Cluster.New.Observer.EnterAnImage',
              defaultMessage: '请输入镜像',
            })}
          />
        </CustomFormItem>
        <CustomFormItem name="serviceType" label="服务类型">
          <Select placeholder="请选择" options={SERVICE_TYPE} />
        </CustomFormItem>
        <CustomFormItem name="replicas" label="副本数">
          <InputNumber placeholder="请输入" min={1} />
        </CustomFormItem>
        <CustomFormItem name={['resource', 'cpu']} label="CPU 核数">
          <InputNumber placeholder="请输入" min={1} />
        </CustomFormItem>
        <CustomFormItem name={['resource', 'memory']} label="内存大小">
          <InputNumber placeholder="请输入" min={1} addonAfter={SUFFIX_UNIT} />
        </CustomFormItem>
        <p style={titleStyle}>参数设置</p>
        <Form.List name={'parameters'}>
          {(fields, { add, remove }) => (
            <>
              {fields.map(({ name }) => (
                <Row key={name} gutter={[12, 0]}>
                  <Col span={12}>
                    <CustomFormItem name={[name, 'key']}>
                      <Select
                        placeholder="请选择"
                        onChange={(val) => labelChange(val, name)}
                        optionLabelProp="selectLabel"
                        options={listParametersOptions?.map((item) => ({
                          value: item.value,
                          selectLabel: item.label,
                          label: (
                            <div
                              style={{
                                display: 'flex',
                                justifyContent: 'space-between',
                              }}
                            >
                              <Text>{item.label}</Text>
                              <Text
                                ellipsis={{ tooltip: item.info }}
                                style={{ color: '#8592AD', width: 120 }}
                              >
                                {item.info}
                              </Text>
                            </div>
                          ),
                        }))}
                      />
                    </CustomFormItem>
                  </Col>
                  <Col span={11}>
                    <CustomFormItem name={[name, 'value']}>
                      <Input />
                    </CustomFormItem>
                  </Col>
                  <Col
                    style={{
                      lineHeight: '32px',
                      textAlign: 'center',
                    }}
                    span={1}
                  >
                    <MinusCircleOutlined onClick={() => remove(name)} />
                  </Col>
                </Row>
              ))}
              <CustomFormItem>
                <Button
                  type="dashed"
                  onClick={() => add()}
                  block
                  icon={<PlusOutlined />}
                >
                  {intl.formatMessage({
                    id: 'OBDashboard.components.NodeSelector.AddNodeselector',
                    defaultMessage: '添加 nodeSelector',
                  })}
                </Button>
              </CustomFormItem>
            </>
          )}
        </Form.List>
      </Form>
    </AlertDrawer>
  );
}
