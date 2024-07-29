import HandleAccountModal from '@/components/customModal/HandleAccountModal';
import HandleRoleModal from '@/components/customModal/HandleRoleModal';
import { PageContainer } from '@ant-design/pro-components';
import type { TabsProps } from 'antd';
import { Button, Tabs } from 'antd';
import { useState } from 'react';
import Accounts from './Accounts';
import Roles from './Roles';
import { ActiveKey, Type } from './type';

export default function Access() {
  const [activeKey, setActiveKey] = useState<ActiveKey>(ActiveKey.ACCOUNTS);
  const [modalVisible, setModalVisible] = useState<boolean>(false);
  const [accountVisible, setAccountVisible] = useState<boolean>(false);

  const onChange = (key: ActiveKey) => {
    setActiveKey(key);
  };

  const items: TabsProps['items'] = [
    {
      key: ActiveKey.ACCOUNTS,
      label: '用户',
      children: <Accounts />,
    },
    {
      key: ActiveKey.ROLES,
      label: '角色',
      children: <Roles />,
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
          <Button type="primary" onClick={() => create(activeKey)}>
            {activeKey === ActiveKey.ACCOUNTS ? '创建账户' : '创建角色'}
          </Button>
        }
        activeKey={activeKey}
        items={items}
        onChange={onChange}
      />
      <HandleAccountModal
        visible={accountVisible}
        setVisible={setAccountVisible}
        type={Type.CREATE}
      />
      <HandleRoleModal
        visible={modalVisible}
        setVisible={setModalVisible}
        type={Type.CREATE}
      />
    </PageContainer>
  );
}
