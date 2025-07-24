import { alert } from '@/api';
import { AlarmMatcher } from '@/api/generated';
import AlertDrawer from '@/components/AlertDrawer';
import IconTip from '@/components/IconTip';
import InputLabelComp from '@/components/InputLabelComp';
import { VALIDATE_DEBOUNCE } from '@/constants';
import { DATE_TIME_FORMAT } from '@/constants/datetime';
import { LABEL_NAME_RULE } from '@/constants/rules';
import { Alert } from '@/type/alert';
import { intl } from '@/utils/intl';
import { useModel } from '@umijs/max';
import { useRequest } from 'ahooks';
import type { DrawerProps } from 'antd';
import {
  Button,
  Col,
  DatePicker,
  Form,
  Input,
  Radio,
  Row,
  Select,
  message,
} from 'antd';
import dayjs from 'dayjs';
import { useEffect } from 'react';
import {
  formatShieldSubmitData,
  getInstancesFromRes,
  getSelectList,
  validateLabelValues,
} from '../helper';
import ClusterSelect from './ClusterSelect';
import ServerSelect from './ServerSelect';
import TenantSelect from './TenantSelect';

interface ShieldDrawerProps extends DrawerProps {
  id?: string;
  initialValues?: Alert.ShieldDrawerInitialValues;
  onClose: () => void;
  submitCallback?: () => void;
}

const { TextArea } = Input;

export default function ShieldDrawerForm({
  id,
  onClose,
  initialValues,
  submitCallback,
  ...props
}: ShieldDrawerProps) {
  const [form] = Form.useForm<Alert.ShieldDrawerForm>();
  const { clusterList, tenantList } = useModel('alarm');
  const isEdit = !!id;

  const newInitialValues = {
    matchers: [
      {
        name: '',
        value: '',
        isRegex: false,
      },
    ],

    instances: {
      type: 'obcluster',
    },
    ...initialValues,
  };
  const { data: listRulesRes } = useRequest(alert.listRules);
  const listRules = listRulesRes?.data?.map((rule) => ({
    label: rule.name,
    value: rule.name,
  }));
  const fieldEndTimeChange = (time: number | Date) => {
    if (typeof time === 'number') {
      form.setFieldValue('endsAt', dayjs(new Date().valueOf() + time));
    } else {
      form.setFieldValue('endsAt', dayjs(time));
    }
    form.validateFields(['endsAt']);
  };

  const submit = (values: Alert.ShieldDrawerForm) => {
    if (!values.matchers) values.matchers = [];
    values.matchers = values.matchers.filter(
      (matcher) => matcher.name && matcher.value,
    );
    const _clusterList = getSelectList(
      clusterList!,
      values.instances.type,
      tenantList,
    );
    if (isEdit) values.id = id;
    alert
      .createOrUpdateSilencer(formatShieldSubmitData(values, _clusterList))
      .then(({ successful }) => {
        if (successful) {
          message.success(
            intl.formatMessage({
              id: 'src.pages.Alert.Shield.04F03990',
              defaultMessage: '操作成功!',
            }),
          );
          submitCallback && submitCallback();
          onClose();
        }
      });
  };

  useEffect(() => {
    if (isEdit) {
      alert.getSilencer(id).then(({ successful, data }) => {
        if (successful) {
          form.setFieldsValue({
            comment: data.comment,
            matchers: data.matchers,
            endsAt: dayjs(data.endsAt * 1000),
            instances: getInstancesFromRes(data.instances),
            rules: data.rules,
          });
        }
      });
    }
  }, [id]);

  return (
    <AlertDrawer
      onClose={() => {
        onClose();
      }}
      destroyOnClose={true}
      onSubmit={() => form.submit()}
      title={intl.formatMessage({
        id: 'src.pages.Alert.Shield.BBB1C040',
        defaultMessage: '屏蔽条件',
      })}
      {...props}
    >
      <Form
        form={form}
        preserve={false}
        onFinish={submit}
        layout="vertical"
        initialValues={newInitialValues}
      >
        <Form.Item
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'src.pages.Alert.Shield.1EFE1B33',
                defaultMessage: '请选择',
              }),
            },
          ]}
          name={['instances', 'type']}
          label={intl.formatMessage({
            id: 'src.pages.Alert.Shield.3F7E9AA4',
            defaultMessage: '屏蔽对象类型',
          })}
        >
          <Radio.Group
            onChange={() => {
              form.setFieldsValue({
                instances: {
                  obcluster: undefined,
                  obtenant: undefined,
                  observer: undefined,
                },
              });
            }}
          >
            <Radio value="obcluster">
              {intl.formatMessage({
                id: 'src.pages.Alert.Shield.A4A5A44C',
                defaultMessage: '集群',
              })}
            </Radio>
            <Radio value="obtenant">
              {intl.formatMessage({
                id: 'src.pages.Alert.Shield.3B8541B0',
                defaultMessage: '租户',
              })}
            </Radio>
            <Radio value="observer"> OBServer </Radio>
          </Radio.Group>
        </Form.Item>
        <Form.Item noStyle dependencies={['instances', 'type']}>
          {({ getFieldValue }) => {
            const type = getFieldValue(['instances', 'type']);
            return (
              <Form.Item
                label={intl.formatMessage({
                  id: 'src.pages.Alert.Shield.68A6FF3F',
                  defaultMessage: '屏蔽对象',
                })}
                name={['instances']}
                required
                rules={[
                  {
                    validator: (_, instances: Alert.InstancesType) => {
                      if (
                        !instances.obcluster ||
                        !instances.obcluster.length ||
                        !instances[instances.type] ||
                        !instances[instances.type]?.length
                      ) {
                        return Promise.reject(
                          new Error(
                            intl.formatMessage({
                              id: 'src.pages.Alert.Shield.FA536583',
                              defaultMessage: '请选择',
                            }),
                          ),
                        );
                      }
                      return Promise.resolve();
                    },
                  },
                ]}
              >
                {type === 'obcluster' && <ClusterSelect />}
                {type === 'obtenant' && <TenantSelect />}
                {type === 'observer' && <ServerSelect />}
              </Form.Item>
            );
          }}
        </Form.Item>

        <Form.Item
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'src.pages.Alert.Shield.248B977C',
                defaultMessage: '请选择',
              }),
            },
          ]}
          name={'rules'}
          label={intl.formatMessage({
            id: 'src.pages.Alert.Shield.376304F5',
            defaultMessage: '屏蔽告警规则',
          })}
        >
          <Select
            mode="multiple"
            allowClear
            style={{ width: '100%' }}
            placeholder={intl.formatMessage({
              id: 'src.pages.Alert.Shield.0A4B5A6C',
              defaultMessage: '请选择',
            })}
            options={listRules}
          />
        </Form.Item>
        <Form.Item
          name={'matchers'}
          validateFirst
          validateDebounce={VALIDATE_DEBOUNCE}
          rules={[
            {
              validator: (_, value: AlarmMatcher[]) => {
                if (!validateLabelValues(value)) {
                  return Promise.reject(
                    intl.formatMessage({
                      id: 'src.pages.Alert.Shield.ED437281',
                      defaultMessage: '请检查标签是否完整输入',
                    }),
                  );
                }
                return Promise.resolve();
              },
            },
            LABEL_NAME_RULE,
          ]}
          label={
            <IconTip
              tip={intl.formatMessage({
                id: 'src.pages.Alert.Shield.43E9A1E1',
                defaultMessage:
                  '按照标签匹配条件屏蔽告警，支持值匹配或者正则表达式，当所有条件都满足时告警才会被屏蔽',
              })}
              content={intl.formatMessage({
                id: 'src.pages.Alert.Shield.C4C1B8A3',
                defaultMessage: '标签',
              })}
            />
          }
        >
          <InputLabelComp regex={true} maxLength={8} defaulLabelName="name" />
        </Form.Item>
        <Row style={{ alignItems: 'center' }}>
          <Col>
            <Form.Item
              rules={[
                {
                  required: true,
                  message: intl.formatMessage({
                    id: 'src.pages.Alert.Shield.3D1B719E',
                    defaultMessage: '请选择',
                  }),
                },
              ]}
              name="endsAt"
              label={intl.formatMessage({
                id: 'src.pages.Alert.Shield.11B30410',
                defaultMessage: '屏蔽结束时间',
              })}
            >
              <DatePicker showTime format={DATE_TIME_FORMAT} />
            </Form.Item>
          </Col>
          <Col>
            <Button
              type="link"
              onClick={() => fieldEndTimeChange(6 * 3600 * 1000)}
            >
              {intl.formatMessage({
                id: 'src.pages.Alert.Shield.0CCBF1B7',
                defaultMessage: '6小时',
              })}
            </Button>
            <Button
              type="link"
              onClick={() => fieldEndTimeChange(12 * 3600 * 1000)}
            >
              {intl.formatMessage({
                id: 'src.pages.Alert.Shield.00AF8CB8',
                defaultMessage: '12小时',
              })}
            </Button>
            <Button
              type="link"
              onClick={() => fieldEndTimeChange(24 * 3600 * 1000)}
            >
              {intl.formatMessage({
                id: 'src.pages.Alert.Shield.42A3A94B',
                defaultMessage: '1天',
              })}
            </Button>
            <Button
              onClick={() =>
                fieldEndTimeChange(new Date('2099-12-31 23:59:59'))
              }
              type="link"
            >
              {intl.formatMessage({
                id: 'src.pages.Alert.Shield.FDAD010D',
                defaultMessage: '永久',
              })}
            </Button>
          </Col>
        </Row>
        <Form.Item
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'src.pages.Alert.Shield.1AC4B3C9',
                defaultMessage: '请输入',
              }),
            },
          ]}
          name={'comment'}
          label={intl.formatMessage({
            id: 'src.pages.Alert.Shield.01368558',
            defaultMessage: '备注信息',
          })}
        >
          <TextArea
            rows={4}
            placeholder={intl.formatMessage({
              id: 'src.pages.Alert.Shield.E49D589A',
              defaultMessage: '请输入',
            })}
          />
        </Form.Item>
      </Form>
    </AlertDrawer>
  );
}
