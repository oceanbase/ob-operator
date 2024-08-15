import { access as accessReq } from '@/api';
import type { AcPolicy, AcRole } from '@/api/generated';
import HandleRoleModal from '@/components/customModal/HandleRoleModal';
import showDeleteConfirm from '@/components/customModal/showDeleteConfirm';
import { intl } from '@/utils/intl';
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
        message.success(
          intl.formatMessage({
            id: 'src.pages.Access.70280437',
            defaultMessage: '删除成功！',
          }),
        );
        refreshRoles();
      }
    },
  });
  const columns: TableProps<AcRole>['columns'] = [
    {
      title: intl.formatMessage({
        id: 'src.pages.Access.3D6FBFA4',
        defaultMessage: '角色',
      }),
      key: 'name',
      dataIndex: 'name',
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Access.14CE430B',
        defaultMessage: '描述',
      }),
      key: 'description',
      dataIndex: 'description',
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Access.C04CE1B8',
        defaultMessage: '权限',
      }),
      key: 'policies',
      dataIndex: 'policies',
      render: (permission) => {
        return (
          <Space size={[8, 16]} wrap>
            {permission.map((item: AcPolicy) => (
              <span>
                {item.domain}:{item.action}
              </span>
            ))}
          </Space>
        );
      },
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Access.3CF8CEC0',
        defaultMessage: '操作',
      }),
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
              {intl.formatMessage({
                id: 'src.pages.Access.D2244128',
                defaultMessage: '编辑',
              })}
            </Button>
            <Button
              disabled={disabled}
              type="link"
              style={disabled ? {} : { color: '#ff4b4b' }}
              onClick={() =>
                showDeleteConfirm({
                  title: intl.formatMessage({
                    id: 'src.pages.Access.CF6370DC',
                    defaultMessage: '你确定要删除该角色吗？',
                  }),
                  onOk: () => deleteRole(record.name),
                })
              }
            >
              {intl.formatMessage({
                id: 'src.pages.Access.F3487256',
                defaultMessage: '删除',
              })}
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
        successCallback={refreshRoles}
        type={Type.EDIT}
      />
    </div>
  );
}
