import { access } from '@/api';
import type { AcAccount, AcCreateAccountParam } from '@/api/generated';
import { encryptText, usePublicKey } from '@/hook/usePublicKey';
import { Type } from '@/pages/Access/type';
import { useRequest } from 'ahooks';
import { Form, Input, Select, message } from 'antd';
import { omit } from 'lodash';
import { useEffect, useMemo } from 'react';
import CustomModal from '.';

interface HandleRoleModalProps {
  visible: boolean;
  setVisible: (visible: boolean) => void;
  successCallback?: () => void;
  editValue?: AcAccount;
  type: Type;
}

export default function HandleAccountModal({
  visible,
  setVisible,
  successCallback,
  editValue,
  type,
}: HandleRoleModalProps) {
  const [form] = Form.useForm();
  const publicKey = usePublicKey();
  const handleSubmit = async () => {
    try {
      await form.validateFields();
      form.submit();
    } catch (err) {}
  };
  const { data: allRolesRes } = useRequest(access.listAllRoles);
  const rolesOption = useMemo(() => {
    return (
      allRolesRes?.data.map((role) => ({
        label: role.name,
        value: role.name,
      })) || []
    );
  }, [allRolesRes]);
  const onFinish = async (
    formData: AcCreateAccountParam & { confirmPassword: string },
  ) => {
    formData.password = encryptText(formData.password, publicKey) as string;
    const res =
      type === Type.CREATE
        ? await access.createAccount(omit(formData, ['confirmPassword']))
        : await access.patchAccount(
            formData.username,
            omit(formData, ['confirmPassword', 'username']),
          );
    if (res.successful) {
      message.success('操作成功！');
      if (successCallback) successCallback();
      form.resetFields();
      setVisible(false);
    }
  };

  useEffect(() => {
    if (type === Type.EDIT) {
      form.setFieldsValue({
        description: editValue?.description,
        nickname: editValue?.nickname,
        roles: editValue?.roles.map((item) => item.name),
      });
    }
  }, [type, editValue]);
  return (
    <CustomModal
      title={`${type === Type.EDIT ? '编辑' : '创建'}用户`}
      isOpen={visible}
      handleOk={handleSubmit}
      handleCancel={() => {
        form.resetFields();
        setVisible(false);
      }}
    >
      <Form form={form} labelCol={{ span: 4 }} onFinish={onFinish}>
        {type === Type.CREATE && (
          <Form.Item
            rules={[{ required: true, message: '请输入用户名' }]}
            label="用户名"
            name={'username'}
          >
            <Input placeholder="请输入" />
          </Form.Item>
        )}

        <Form.Item label="描述" name={'description'}>
          <Input placeholder="请输入" />
        </Form.Item>
        <Form.Item
          rules={[{ required: true, message: '请输入昵称' }]}
          label="昵称"
          name={'nickname'}
        >
          <Input placeholder="请输入" />
        </Form.Item>
        <Form.Item
          rules={[{ required: true, message: '请选择角色' }]}
          label="角色"
          name={'roles'}
        >
          <Select mode="multiple" options={rolesOption} placeholder="请选择" />
        </Form.Item>
        {type === Type.CREATE && (
          <Form.Item
            rules={[
              {
                required: true,
                message: '请输入',
              },
            ]}
            label="密码"
            name={'password'}
          >
            <Input type="password" placeholder="请输入" />
          </Form.Item>
        )}
        {type === Type.CREATE && (
          <Form.Item
            validateFirst
            rules={[
              {
                required: true,
                message: '请输入',
              },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (getFieldValue('password') !== value) {
                    return Promise.reject(new Error('两次密码输入不一致'));
                  }
                  return Promise.resolve();
                },
              }),
            ]}
            label="确认密码"
            name={'confirmPassword'}
          >
            <Input type="password" placeholder="请输入" />
          </Form.Item>
        )}
      </Form>
    </CustomModal>
  );
}
