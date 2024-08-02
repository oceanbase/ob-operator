import { access as accessReq } from '@/api';
import type { AcPolicy, AcRole } from '@/api/generated';
import HandleRoleModal from '@/components/customModal/HandleRoleModal';
import showDeleteConfirm from '@/components/customModal/showDeleteConfirm';
import { useAccess } from '@umijs/max';
import { useRequest } from 'ahooks';
import type { TableProps } from 'antd';
import { Button, Space, Table, message } from 'antd';
import { useState } from 'react';
import { Type } from './type';

interface RolesProps {
  allRoles: AcRole[] | undefined;
  refreshRoles: () => void;
}

export default function Roles({ allRoles, refreshRoles }: RolesProps) {
  const [modalVisible, setModalVisible] = useState<boolean>(false);
  const access = useAccess();
  const [editData, setEditData] = useState<AcRole>();

  const { run: deleteRole } = useRequest(accessReq.deleteRole, {
    manual: true,
    onSuccess: ({ successful }) => {
      if (successful) {
        message.success('删除成功！');
        refreshRoles();
      }
    },
  });
  const columns: TableProps<AcRole>['columns'] = [
    {
      title: '角色',
      key: 'name',
      dataIndex: 'name',
    },
    {
      title: '描述',
      key: 'description',
      dataIndex: 'description',
    },
    {
      title: '权限',
      key: 'policies',
      dataIndex: 'policies',
      render: (permission) => {
        return (
          <Space size={[8, 16]} wrap>
            {permission.map((item: AcPolicy) => (
              <span>
                {item.object}：{item.action}
              </span>
            ))}
          </Space>
        );
      },
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => {
        const disabled = record.name === 'admin' || !access.acwrite;
        return (
          <Space>
            <Button
              onClick={() => handleEdit(record)}
              disabled={disabled}
              type="link"
            >
              编辑
            </Button>
            <Button
              disabled={disabled}
              type="link"
              style={disabled ? {} : { color: '#ff4b4b' }}
              onClick={() =>
                showDeleteConfirm({
                  title: '你确定要删除该角色吗？',
                  onOk: () => deleteRole(record.name),
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

  const handleEdit = (editData: AcRole) => {
    setEditData(editData);
    setModalVisible(true);
  };

  return (
    <div>
      <Table dataSource={allRoles} rowKey={'name'} columns={columns} />
      <HandleRoleModal
        visible={modalVisible}
        editValue={editData}
        setVisible={setModalVisible}
        type={Type.EDIT}
      />
    </div>
  );
}
