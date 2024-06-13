import { obproxy } from '@/api';
import type { CommonKVPair, CommonResourceSpec } from '@/api/generated';
import { ObproxyPatchOBProxyParam } from '@/api/generated';
import AlertDrawer from '@/components/AlertDrawer';
import { CustomFormItem } from '@/components/CustomFormItem';
import InputLabelComp from '@/components/InputLabelComp';
import { SERVICE_TYPE, SUFFIX_UNIT } from '@/constants';
import { MIRROR_OBPROXY } from '@/constants/doc';
import { intl } from '@/utils/intl';
import type { DrawerProps } from 'antd';
import { Button, Form, Input, InputNumber, Select, Space } from 'antd';

interface ConfigDrawerProps extends DrawerProps {
  onClose: () => void;
  name?: string;
  namespace?: string;
  image?: string;
  parameters?: CommonKVPair[];
  resource?: CommonResourceSpec;
  serviceType?: string;
  replicas?: number;
}

type FormValue = {
  parameters: { key: string; value: string }[];
} & ObproxyPatchOBProxyParam;

export default function ConfigDrawer({
  onClose,
  name,
  namespace,
  ...props
}: ConfigDrawerProps) {
  const [form] = Form.useForm();
  const submit = (values: FormValue) => {
    if (namespace && name) obproxy.patchOBProxy(namespace, name, values);
  };
  const Footer = () => {
    return (
      <div>
        <Space>
          <Button onClick={() => {}} type="primary">
            提交
          </Button>
          <Button onClick={onClose}>取消</Button>
        </Space>
      </div>
    );
  };
  const titleStyle = { fontSize: 14, fontWeight: 600 };
  return (
    <AlertDrawer
      title="详细配置"
      footer={<Footer />}
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
        <CustomFormItem name={'parameters'}>
          <InputLabelComp />
        </CustomFormItem>
      </Form>
    </AlertDrawer>
  );
}
