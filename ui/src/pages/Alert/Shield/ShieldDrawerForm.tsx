import { alert } from '@/api';
import AlertDrawer from '@/components/AlertDrawer';
import IconTip from '@/components/IconTip';
import InputLabel from '@/components/InputLabel';
import { Alert } from '@/type/alert';
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
  filterLabel
} from '../helper';
import ShieldObjInput from './ShieldObjInput';

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
  const shieldObjType = Form.useWatch(['instances', 'type'], form);
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
    values.matchers = filterLabel(values.matchers);
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
          message.success('操作成功!');
          submitCallback && submitCallback();
          onClose();
        }
      });
  };

  useEffect(() => {
    if (isEdit) {
      alert.getSilencer(id).then(({ successful, data }) => {
        if (successful) {
          console.log(getInstancesFromRes(data.instances));

          form.setFieldsValue({
            comment: data.comment,
            matchers: data.matchers,
            endsAt: dayjs(data.endsAt),
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
      title="屏蔽条件"
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
              message: '请选择',
            },
          ]}
          name={['instances', 'type']}
          label="屏蔽对象类型"
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
            <Radio value="obcluster"> 集群 </Radio>
            <Radio value="obtenant"> 租户 </Radio>
            <Radio value="observer"> OBServer </Radio>
          </Radio.Group>
        </Form.Item>
        <Form.Item style={{ marginBottom: 0 }} label="屏蔽对象">
          <ShieldObjInput shieldObjType={shieldObjType} form={form} />
        </Form.Item>
        <Form.Item
          rules={[
            {
              required: true,
              message: '请选择',
            },
          ]}
          name={'rules'}
          label="屏蔽告警规则"
        >
          <Select
            mode="multiple"
            allowClear
            style={{ width: '100%' }}
            placeholder="请选择"
            options={listRules}
          />
        </Form.Item>
        <Form.Item
          label={
            <IconTip
              tip="支持对指定的指标进行屏蔽，如慢 SQL告警，支持对 SQLID 进行过滤，支持正则表达式"
              content={'标签'}
            />
          }
        >
          <InputLabel
            wrapFormName="matchers"
            labelFormName="name"
            valueFormName="value"
            regBoxFormName="isRegex"
            form={form}
            maxCount={8}
          />
        </Form.Item>
        <Row style={{ alignItems: 'center' }}>
          <Col>
            <Form.Item
              rules={[
                {
                  required: true,
                  message: '请选择',
                },
              ]}
              name="endsAt"
              label="屏蔽结束时间"
            >
              <DatePicker showTime format="YYYY-MM-DD HH:mm:ss" />
            </Form.Item>
          </Col>
          <Col>
            <Button
              type="link"
              onClick={() => fieldEndTimeChange(6 * 3600 * 1000)}
            >
              6小时
            </Button>
            <Button
              type="link"
              onClick={() => fieldEndTimeChange(12 * 3600 * 1000)}
            >
              12小时
            </Button>
            <Button
              type="link"
              onClick={() => fieldEndTimeChange(24 * 3600 * 1000)}
            >
              1天
            </Button>
            <Button
              onClick={() =>
                fieldEndTimeChange(new Date('2099-12-31 23:59:59'))
              }
              type="link"
            >
              永久
            </Button>
          </Col>
        </Row>
        <Form.Item
          rules={[
            {
              required: true,
              message: '请输入',
            },
          ]}
          name={'comment'}
          label="备注信息"
        >
          <TextArea rows={4} placeholder="请输入" />
        </Form.Item>
      </Form>
    </AlertDrawer>
  );
}
