import { access as accessReq } from '@/api';
import type { AcAccount, AcRole } from '@/api/generated';
import HandleAccountModal from '@/components/customModal/HandleAccountModal';
import showDeleteConfirm from '@/components/customModal/showDeleteConfirm';
import { useAccess } from '@umijs/max';
import { useRequest } from 'ahooks';
import type { TableProps } from 'antd';
import { Button, Space, Table, message } from 'antd';
import { useState } from 'react';
import { Type } from './type';

interface AccountsProps {
  allAccounts: AcAccount[] | undefined;
  refreshAccounts: () => void;
}

export default function Accounts({
  allAccounts,
  refreshAccounts,
}: AccountsProps) {
  const access = useAccess();
  const [editData, setEditData] = useState<AcAccount>();
  const [modalVisible, setModalVisible] = useState<boolean>(false);
  const { run: deleteAccount } = useRequest(accessReq.deleteAccount, {
    manual: true,
    onSuccess: ({ successful }) => {
      if (successful) {
        message.success('删除成功');
        refreshAccounts();
      }
    },
  });
  const handleEdit = (editData: AcAccount) => {
    setEditData(editData);
    setModalVisible(true);
  };
  const columns: TableProps<AcAccount>['columns'] = [
    {
      title: '用户名',
      key: 'username',
      dataIndex: 'username',
    },
    {
      title: '昵称',
      key: 'nickname',
      dataIndex: 'nickname',
    },
    {
      title: '描述',
      key: 'description',
      dataIndex: 'description',
    },
    {
      title: '角色',
      key: 'roles',
      dataIndex: 'roles',
      render: (roles: AcRole[]) => (
        <span>{roles.map((role) => role.name).join(',')}</span>
      ),
    },
    {
      title: '最近一次登陆时间',
      key: 'lastLoginAt',
      dataIndex: 'lastLoginAt',
    },
    {
      title: '操作',
      key: 'actinon',
      render: (_, record) => {
        const disabled =
          record.roles.some((role) => role.name === 'admin') || !access.acwrite;
        return (
          <Space>
            <Button disabled={disabled} type="link">
              重置
            </Button>
            <Button
              onClick={() => handleEdit(record)}
              disabled={disabled}
              type="link"
            >
              编辑
            </Button>
            <Button
              disabled={disabled}
              style={disabled ? {} : { color: '#ff4b4b' }}
              type="link"
              onClick={() =>
                showDeleteConfirm({
                  title: '你确定要删除该用户吗？',
                  onOk: () => deleteAccount(record.username),
                })
              }
            >
              删除
            </Button>
          </Space>
        );
      },
    },
  ];
  return (
    <div>
      <Table dataSource={allAccounts} columns={columns} />
      <HandleAccountModal
        setVisible={setModalVisible}
        editValue={editData}
        visible={modalVisible}
        type={Type.EDIT}
      />
    </div>
  );
}
