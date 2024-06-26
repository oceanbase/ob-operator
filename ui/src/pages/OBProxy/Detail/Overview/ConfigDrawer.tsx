import { obproxy } from '@/api';
import type { CommonKVPair } from '@/api/generated';
import { ObproxyPatchOBProxyParam } from '@/api/generated';
import AlertDrawer from '@/components/AlertDrawer';
import { CustomFormItem } from '@/components/CustomFormItem';
import IconTip from '@/components/IconTip';
import { SERVICE_TYPE, SUFFIX_UNIT } from '@/constants';
import { MIRROR_OBPROXY } from '@/constants/doc';
import { OBProxy } from '@/type/obproxy';
import { intl } from '@/utils/intl';
import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';
import { useDebounceFn, useRequest } from 'ahooks';
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
  message,
} from 'antd';
import { useEffect, useRef } from 'react';
import { isDifferentParams } from '../../helper';

type ConfigDrawerProps = {
  onClose: () => void;
  submitCallback?: () => void;
} & OBProxy.CommonProxyDetail &
  DrawerProps;

type FormValue = {
  parameters?: { key: string; value: string }[];
} & ObproxyPatchOBProxyParam;

const { Text } = Typography;

export default function ConfigDrawer({
  onClose,
  name,
  namespace,
  submitCallback,
  ...props
}: ConfigDrawerProps) {
  const [form] = Form.useForm();

  const preParameters = useRef<CommonKVPair[] | undefined>();
  const { data: listParametersRes } = useRequest(
    obproxy.listOBProxyParameters,
    {
      defaultParams: [namespace!, name!],
    },
  );
  const listParametersOptions = listParametersRes?.data.map((item) => ({
    label: item.name,
    value: item.name,
    info: item.info,
  }));

  const submit = async (values: FormValue) => {
    if (
      !isDifferentParams(values.parameters || [], preParameters.current || [])
    ) {
      delete values.parameters;
    }
    const res = await obproxy.patchOBProxy(namespace!, name!, values);
    if (res.successful) {
      message.success(
        intl.formatMessage({
          id: 'src.pages.OBProxy.Detail.Overview.A9BB34D6',
          defaultMessage: '操作成功！',
        }),
      );
      submitCallback && submitCallback();
      onClose();
    }
  };
  const { run: debounceSubmit } = useDebounceFn(submit, { wait: 300 });
  const titleStyle = { fontSize: 14, fontWeight: 600 };

  const labelChange = (label: string, name: number) => {
    const value = listParametersRes?.data?.find(
      (parameter) => parameter.name === label,
    )?.value;
    if (typeof value !== 'undefined') {
      form.setFieldValue(['parameters', name, 'value'], value);
    }
  };
  useEffect(() => {
    preParameters.current = props.parameters;
  }, [props.parameters]);

  return (
    <AlertDrawer
      title={intl.formatMessage({
        id: 'src.pages.OBProxy.Detail.Overview.680A1826',
        defaultMessage: '详细配置',
      })}
      onSubmit={() => form.submit()}
      destroyOnClose={true}
      onClose={() => {
        form.resetFields();
        onClose();
      }}
      {...props}
    >
      <Form
        initialValues={props}
        form={form}
        onFinish={debounceSubmit}
        layout="vertical"
      >
        <p style={titleStyle}>
          {intl.formatMessage({
            id: 'src.pages.OBProxy.Detail.Overview.F9D66FC0',
            defaultMessage: '资源设置',
          })}
        </p>
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
        <CustomFormItem
          name="serviceType"
          label={intl.formatMessage({
            id: 'src.pages.OBProxy.Detail.Overview.DBE1B6C0',
            defaultMessage: '服务类型',
          })}
        >
          <Select
            placeholder={intl.formatMessage({
              id: 'src.pages.OBProxy.Detail.Overview.E7B1B575',
              defaultMessage: '请选择',
            })}
            options={SERVICE_TYPE}
          />
        </CustomFormItem>
        <CustomFormItem
          name="replicas"
          label={intl.formatMessage({
            id: 'src.pages.OBProxy.Detail.Overview.2DA6A0A7',
            defaultMessage: '副本数',
          })}
        >
          <InputNumber
            placeholder={intl.formatMessage({
              id: 'src.pages.OBProxy.Detail.Overview.52C1F09A',
              defaultMessage: '请输入',
            })}
            min={1}
          />
        </CustomFormItem>
        <CustomFormItem
          name={['resource', 'cpu']}
          label={intl.formatMessage({
            id: 'src.pages.OBProxy.Detail.Overview.A4448DF2',
            defaultMessage: 'CPU 核数',
          })}
        >
          <InputNumber
            placeholder={intl.formatMessage({
              id: 'src.pages.OBProxy.Detail.Overview.2FA13720',
              defaultMessage: '请输入',
            })}
            min={1}
          />
        </CustomFormItem>
        <CustomFormItem
          name={['resource', 'memory']}
          label={intl.formatMessage({
            id: 'src.pages.OBProxy.Detail.Overview.0075C2B3',
            defaultMessage: '内存大小',
          })}
        >
          <InputNumber
            placeholder={intl.formatMessage({
              id: 'src.pages.OBProxy.Detail.Overview.3F3066D5',
              defaultMessage: '请输入',
            })}
            min={1}
            addonAfter={SUFFIX_UNIT}
          />
        </CustomFormItem>
        <p style={titleStyle}>
          <IconTip
            tip={intl.formatMessage({
              id: 'src.pages.OBProxy.Detail.Overview.FEBB24FF',
              defaultMessage:
                '删除列表中的参数并不会将原来已经设置的参数复原，如果需要参数立即生效，请覆盖参数为原值或者重启集群',
            })}
            content={intl.formatMessage({
              id: 'src.pages.OBProxy.Detail.Overview.D537DD35',
              defaultMessage: '参数设置',
            })}
          />
        </p>
        <Form.List name={'parameters'}>
          {(fields, { add, remove }) => (
            <>
              {fields.map(({ name }) => (
                <Row key={name} gutter={[12, 0]}>
                  <Col span={12}>
                    <CustomFormItem name={[name, 'key']}>
                      <Select
                        placeholder={intl.formatMessage({
                          id: 'src.pages.OBProxy.Detail.Overview.8A75D872',
                          defaultMessage: '请选择',
                        })}
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
                      <Input
                        placeholder={intl.formatMessage({
                          id: 'src.pages.OBProxy.Detail.Overview.A6701513',
                          defaultMessage: '请输入',
                        })}
                      />
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
                    id: 'src.pages.OBProxy.Detail.Overview.8E87D135',
                    defaultMessage: '添加',
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
