import { access as accessReq } from '@/api';
import HandleAccountModal from '@/components/customModal/HandleAccountModal';
import HandleRoleModal from '@/components/customModal/HandleRoleModal';
import { PageContainer } from '@ant-design/pro-components';
import { useAccess } from '@umijs/max';
import { useRequest } from 'ahooks';
import type { TabsProps } from 'antd';
import { Button, Tabs } from 'antd';
import { useState } from 'react';
import Accounts from './Accounts';
import Roles from './Roles';
import { ActiveKey, Type } from './type';

export default function Access() {
  const [activeKey, setActiveKey] = useState<ActiveKey>(ActiveKey.ACCOUNTS);
  const access = useAccess();
  const [modalVisible, setModalVisible] = useState<boolean>(false);
  const [accountVisible, setAccountVisible] = useState<boolean>(false);
  const { data: allRolesRes, refresh: refreshRoles } = useRequest(
    accessReq.listAllRoles,
  );
  const { data: allAccountsRes, refresh: refreshAccounts } = useRequest(
    accessReq.listAllAccounts,
  );
  const allRoles = allRolesRes?.data;
  const allAccounts = allAccountsRes?.data;
  const onChange = (key: ActiveKey) => {
    setActiveKey(key);
  };

  const items: TabsProps['items'] = [
    {
      key: ActiveKey.ACCOUNTS,
      label: '用户',
      children: (
        <Accounts allAccounts={allAccounts} refreshAccounts={refreshAccounts} />
      ),
    },
    {
      key: ActiveKey.ROLES,
      label: '角色',
      children: <Roles allRoles={allRoles} refreshRoles={refreshRoles} />,
    },
  ];

  const create = (type: ActiveKey) => {
    if (type === ActiveKey.ACCOUNTS) {
      setAccountVisible(true);
    } else {
      setModalVisible(true);
    }
  };

  return (
    <PageContainer title="权限控制">
      <Tabs
        tabBarExtraContent={
          access.acwrite ? (
            <Button type="primary" onClick={() => create(activeKey)}>
              {activeKey === ActiveKey.ACCOUNTS ? '创建账户' : '创建角色'}
            </Button>
          ) : null
        }
        activeKey={activeKey}
        items={items}
        onChange={onChange}
      />
      <HandleAccountModal
        visible={accountVisible}
        setVisible={setAccountVisible}
        type={Type.CREATE}
        successCallback={refreshAccounts}
      />
      <HandleRoleModal
        visible={modalVisible}
        setVisible={setModalVisible}
        type={Type.CREATE}
        successCallback={refreshRoles}
      />
    </PageContainer>
  );
}
