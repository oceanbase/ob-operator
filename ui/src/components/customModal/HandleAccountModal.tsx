import { access } from '@/api';
import type { AcAccount, AcCreateAccountParam } from '@/api/generated';
import { encryptText, usePublicKey } from '@/hook/usePublicKey';
import { Type } from '@/pages/Access/type';
import { intl } from '@/utils/intl';
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
  const { data: allRolesRes } = useRequest(access.listAllRoles, {
    refreshDeps: [visible],
  });
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
            editValue!.username,
            omit(formData, ['confirmPassword', 'username', 'password']),
          );
    if (res.successful) {
      message.success(
        intl.formatMessage({
          id: 'src.components.customModal.8EA35AF0',
          defaultMessage: '操作成功！',
        }),
      );
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
  }, [type, editValue, visible]);

  return (
    <CustomModal
      title={intl.formatMessage(
        {
          id: 'src.components.customModal.3487AEC1',
          defaultMessage: '{ConditionalExpression0}用户',
        },
        {
          ConditionalExpression0:
            type === Type.EDIT
              ? intl.formatMessage({
                  id: 'src.components.customModal.F4E9F1AB',
                  defaultMessage: '编辑',
                })
              : intl.formatMessage({
                  id: 'src.components.customModal.2EDC3613',
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
      <Form form={form} labelCol={{ span: 4 }} onFinish={onFinish}>
        {type === Type.CREATE && (
          <Form.Item
            rules={[
              {
                required: true,
                message: intl.formatMessage({
                  id: 'src.components.customModal.7DB0CF1B',
                  defaultMessage: '请输入用户名',
                }),
              },
            ]}
            label={intl.formatMessage({
              id: 'src.components.customModal.94E51D06',
              defaultMessage: '用户名',
            })}
            name={'username'}
          >
            <Input
              placeholder={intl.formatMessage({
                id: 'src.components.customModal.05BFF296',
                defaultMessage: '请输入',
              })}
            />
          </Form.Item>
        )}
        <Form.Item
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'src.components.customModal.449E2A83',
                defaultMessage: '请输入昵称',
              }),
            },
          ]}
          label={intl.formatMessage({
            id: 'src.components.customModal.7374EF1F',
            defaultMessage: '昵称',
          })}
          name={'nickname'}
        >
          <Input
            placeholder={intl.formatMessage({
              id: 'src.components.customModal.DA5DFFAE',
              defaultMessage: '请输入',
            })}
          />
        </Form.Item>
        <Form.Item
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'src.components.customModal.806DEA1C',
                defaultMessage: '请选择角色',
              }),
            },
          ]}
          label={intl.formatMessage({
            id: 'src.components.customModal.A9A06567',
            defaultMessage: '角色',
          })}
          name={'roles'}
        >
          <Select
            mode="multiple"
            options={rolesOption}
            placeholder={intl.formatMessage({
              id: 'src.components.customModal.DED9610F',
              defaultMessage: '请选择',
            })}
          />
        </Form.Item>
        <Form.Item
          label={intl.formatMessage({
            id: 'src.components.customModal.CC59F523',
            defaultMessage: '描述',
          })}
          name={'description'}
        >
          <Input
            placeholder={intl.formatMessage({
              id: 'src.components.customModal.80E5AC42',
              defaultMessage: '请输入',
            })}
          />
        </Form.Item>

        {type === Type.CREATE && (
          <Form.Item
            rules={[
              {
                required: true,
                message: intl.formatMessage({
                  id: 'src.components.customModal.5B73A2BF',
                  defaultMessage: '请输入',
                }),
              },
            ]}
            label={intl.formatMessage({
              id: 'src.components.customModal.966B1EBB',
              defaultMessage: '密码',
            })}
            name={'password'}
          >
            <Input
              type="password"
              placeholder={intl.formatMessage({
                id: 'src.components.customModal.263979F5',
                defaultMessage: '请输入',
              })}
            />
          </Form.Item>
        )}

        {type === Type.CREATE && (
          <Form.Item
            validateFirst
            rules={[
              {
                required: true,
                message: intl.formatMessage({
                  id: 'src.components.customModal.13D0A14C',
                  defaultMessage: '请输入',
                }),
              },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (getFieldValue('password') !== value) {
                    return Promise.reject(
                      new Error(
                        intl.formatMessage({
                          id: 'src.components.customModal.A1607740',
                          defaultMessage: '两次密码输入不一致',
                        }),
                      ),
                    );
                  }
                  return Promise.resolve();
                },
              }),
            ]}
            label={intl.formatMessage({
              id: 'src.components.customModal.4EF7C449',
              defaultMessage: '确认密码',
            })}
            name={'confirmPassword'}
          >
            <Input
              type="password"
              placeholder={intl.formatMessage({
                id: 'src.components.customModal.E903F734',
                defaultMessage: '请输入',
              })}
            />
          </Form.Item>
        )}
      </Form>
    </CustomModal>
  );
}
