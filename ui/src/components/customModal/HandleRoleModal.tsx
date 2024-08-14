import { access } from '@/api';
import type { AcCreateRoleParam, AcPolicy, AcRole } from '@/api/generated';
import { Type } from '@/pages/Access/type';
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
    { label: '读', value: 'read' },
    { label: '写', value: 'write' },
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
        所有权限
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
  type,
}: HandleRoleModalProps) {
  const [form] = Form.useForm();
  const { initialState } = useModel('@@initialState');
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
    const res =
      type === Type.CREATE
        ? await access.createRole(formData)
        : await access.patchRole(
            editValue!.name,
            pick(formData, ['description', 'permissions']),
          );
    if (res.successful) {
      message.success('操作成功！');
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
      title={`${type === Type.EDIT ? '编辑' : '创建'}角色`}
      isOpen={visible}
      handleOk={handleSubmit}
      handleCancel={() => {
        form.resetFields();
        setVisible(false);
      }}
    >
      <Form form={form} onFinish={onFinish} preserve={false}>
        {type === Type.CREATE && (
          <Form.Item
            rules={[{ required: true, message: '请输入角色名称' }]}
            label="名称"
            name={'name'}
          >
            <Input placeholder="请输入" />
          </Form.Item>
        )}
        <Form.Item
          label="描述"
          name={'description'}
          rules={[{ required: true, message: '请输入描述' }]}
        >
          <Input placeholder="请输入" />
        </Form.Item>
        <Form.Item required label="权限" name={'permissions'}>
          <PermissionSelect
            fetchData={allPolicies}
            defaultValue={defaultValue}
          />
        </Form.Item>
      </Form>
    </CustomModal>
  );
}
