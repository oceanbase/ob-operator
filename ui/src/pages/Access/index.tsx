import HandleRoleModal from '@/components/customModal/HandleRoleModal';
import { PageContainer } from '@ant-design/pro-components';
import type { TabsProps } from 'antd';
import { Button, Tabs } from 'antd';
import { useState } from 'react';
import Accounts from './Accounts';
import Roles from './Roles';

type ActiveKey = 'accounts' | 'roles';

export default function Access() {
  const [activeKey, setActiveKey] = useState<ActiveKey>('roles');
  const [modalVisible, setModalVisible] = useState<boolean>(false);

  const onChange = (key: ActiveKey) => {
    setActiveKey(key);
  };

  const items: TabsProps['items'] = [
    {
      key: 'accounts',
      label: '用户',
      children: <Accounts />,
    },
    {
      key: 'roles',
      label: '角色',
      children: <Roles />,
    },
  ];

  const create = (type: ActiveKey) => {
    if (type === 'accounts') {
    } else {
      setModalVisible(true);
    }
  };

  return (
    <PageContainer title="权限控制">
      <Tabs
        defaultActiveKey="roles"
        tabBarExtraContent={
          <Button type="primary" onClick={() => create(activeKey)}>
            {activeKey === 'accounts' ? '创建账户' : '创建角色'}
          </Button>
        }
        activeKey={activeKey}
        items={items}
        onChange={onChange}
      />
      <HandleRoleModal
        visible={modalVisible}
        setVisible={setModalVisible}
        type="create"
      />
    </PageContainer>
  );
}
