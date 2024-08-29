import { access } from '@/api';
import type { AcCreateRoleParam, AcPolicy, AcRole } from '@/api/generated';
import { Type } from '@/pages/Access/type';
import { intl } from '@/utils/intl';
import { useModel } from '@umijs/max';
import type { CheckboxProps } from 'antd';
import { Checkbox, Col, Form, Input, Row, message } from 'antd';
import { pick, uniqBy } from 'lodash';
import { useEffect, useState } from 'react';
import CustomModal from '.';

interface HandleRoleModalProps {
  visible: boolean;
  setVisible: (visible: boolean) => void;
  successCallback?: () => void;
  editValue?: AcRole;
  createdRoles?: string[];
  type: Type;
}

interface PermissionSelectProps {
  fetchData: AcPolicy[];
  onChange?: (val: AcPolicy[]) => void;
  defaultValue?: AcPolicy[];
  value?: AcPolicy[];
}

type CheckedList = {
  domain: string;
  checked: string[];
}[];

function PermissionSelect({
  fetchData,
  onChange,
  defaultValue,
  value = [],
}: PermissionSelectProps) {
  const indeterminate = value.length < fetchData.length && value.length > 0;
  const [checkedList, setCheckedList] = useState<CheckedList>([]);
  const checkAll = !checkedList.some(
    (item) =>
      !(item.checked.includes('read') && item.checked.includes('write')),
  );
  const options = [
    {
      label: intl.formatMessage({
        id: 'src.components.customModal.BFC9AB05',
        defaultMessage: '读',
      }),
      value: 'read',
    },
    {
      label: intl.formatMessage({
        id: 'src.components.customModal.6FC754B4',
        defaultMessage: '写',
      }),
      value: 'write',
    },
  ];

  const onCheckAllChange: CheckboxProps['onChange'] = (e) => {
    if (e.target.checked) {
      setCheckedList((preCheckedList) =>
        preCheckedList.map((item) => ({ ...item, checked: ['read', 'write'] })),
      );
    } else {
      setCheckedList((preCheckedList) =>
        preCheckedList.map((item) => ({ ...item, checked: [] })),
      );
    }
  };
  const handleSelected = (val: string[], target: string) => {
    setCheckedList((preCheckedList) => {
      const newList = [...preCheckedList];
      newList.forEach((item) => {
        if (item.domain === target) {
          item.checked = val;
        }
      });
      return newList;
    });
  };
  useEffect(() => {
    if (defaultValue) {
      const newCheckedList: CheckedList = [];
      for (const item of defaultValue) {
        newCheckedList.push({
          domain: item.domain,
          checked: item.action === 'write' ? ['write', 'read'] : [item.action],
        });
      }
      setCheckedList(newCheckedList);
    }
  }, [defaultValue]);

  useEffect(() => {
    const newValue = [];
    for (const item of checkedList) {
      if (!item.checked.includes('write') && !item.checked.includes('read'))
        continue;
      newValue.push({
        action: item.checked.includes('write') ? 'write' : 'read',
        domain: item.domain,
        object: '*',
      });
    }
    newValue.length && onChange?.(newValue);
  }, [checkedList]);
  return (
    <div style={{ padding: 5 }}>
      <Checkbox
        style={{ marginBottom: 16 }}
        indeterminate={indeterminate}
        onChange={onCheckAllChange}
        checked={checkAll}
      >
        {intl.formatMessage({
          id: 'src.components.customModal.1E76C4F3',
          defaultMessage: '所有权限',
        })}
      </Checkbox>
      {fetchData.map((item, index) => (
        <div key={index}>
          <Row gutter={[8, 16]}>
            <Col span={8}> {item.domain}</Col>
            <Col span={16}>
              <Checkbox.Group
                options={options}
                value={
                  checkedList.find(
                    (item) => item.domain === fetchData[index].domain,
                  )?.checked
                }
                onChange={(val) => handleSelected(val, fetchData[index].domain)}
              />
            </Col>
          </Row>
        </div>
      ))}
    </div>
  );
}

export default function HandleRoleModal({
  visible,
  setVisible,
  successCallback,
  editValue,
  createdRoles,
  type,
}: HandleRoleModalProps) {
  const [form] = Form.useForm();
  const { initialState, refresh } = useModel('@@initialState');
  const handleSubmit = async () => {
    try {
      await form.validateFields();
      form.submit();
    } catch (err) {}
  };
  const allPolicies = uniqBy(initialState?.policies, 'domain') || [];
  const defaultValue = allPolicies.map((item) => {
    if (type === Type.CREATE) {
      return { ...item, action: '' };
    } else {
      const editItem = editValue?.policies.find(
        (policy) => policy.domain === item.domain,
      );
      return editItem ? { ...editItem } : { ...item, action: '' };
    }
  });
  const onFinish = async (formData: AcCreateRoleParam) => {
    if (type === Type.CREATE && createdRoles?.includes(formData.name)) {
      message.warning(
        intl.formatMessage({
          id: 'src.components.customModal.08DC5F92',
          defaultMessage: '角色已存在',
        }),
      );
      return;
    }
    const res =
      type === Type.CREATE
        ? await access.createRole(formData)
        : await access.patchRole(
            editValue!.name,
            pick(formData, ['description', 'permissions']),
          );
    if (res.successful) {
      if (type == Type.EDIT) {
        await refresh();
      }
      message.success(
        intl.formatMessage({
          id: 'src.components.customModal.66C28C4A',
          defaultMessage: '操作成功！',
        }),
      );
      if (successCallback) successCallback();
      setVisible(false);
    }
  };

  useEffect(() => {
    if (type === Type.EDIT && visible) {
      form.setFieldsValue({
        description: editValue?.description,
        permissions: editValue?.policies,
      });
    }
  }, [type, editValue, visible]);

  return (
    <CustomModal
      title={intl.formatMessage(
        {
          id: 'src.components.customModal.1F81961E',
          defaultMessage: '{ConditionalExpression0}角色',
        },
        {
          ConditionalExpression0:
            type === Type.EDIT
              ? intl.formatMessage({
                  id: 'src.components.customModal.1920004B',
                  defaultMessage: '编辑',
                })
              : intl.formatMessage({
                  id: 'src.components.customModal.D7653D14',
                  defaultMessage: '创建',
                }),
        },
      )}
      open={visible}
      onOk={handleSubmit}
      onCancel={() => {
        form.resetFields();
        setVisible(false);
      }}
    >
      <Form form={form} onFinish={onFinish} preserve={false}>
        {type === Type.CREATE && (
          <Form.Item
            rules={[
              {
                required: true,
                message: intl.formatMessage({
                  id: 'src.components.customModal.FECE2219',
                  defaultMessage: '请输入角色名称',
                }),
              },
            ]}
            label={intl.formatMessage({
              id: 'src.components.customModal.0515E4FE',
              defaultMessage: '名称',
            })}
            name={'name'}
          >
            <Input
              placeholder={intl.formatMessage({
                id: 'src.components.customModal.5CDA23D6',
                defaultMessage: '请输入',
              })}
            />
          </Form.Item>
        )}

        <Form.Item
          label={intl.formatMessage({
            id: 'src.components.customModal.75E8B57A',
            defaultMessage: '描述',
          })}
          name={'description'}
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'src.components.customModal.00AA38BD',
                defaultMessage: '请输入描述',
              }),
            },
          ]}
        >
          <Input
            placeholder={intl.formatMessage({
              id: 'src.components.customModal.465FF0F1',
              defaultMessage: '请输入',
            })}
          />
        </Form.Item>
        <Form.Item
          required
          label={intl.formatMessage({
            id: 'src.components.customModal.D24B8F5C',
            defaultMessage: '权限',
          })}
          name={'permissions'}
        >
          <PermissionSelect
            fetchData={allPolicies}
            defaultValue={defaultValue}
          />
        </Form.Item>
      </Form>
    </CustomModal>
  );
}
