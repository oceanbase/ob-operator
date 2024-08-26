import { access as accessReq } from '@/api';
import HandleAccountModal from '@/components/customModal/HandleAccountModal';
import HandleRoleModal from '@/components/customModal/HandleRoleModal';
import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { useAccess } from '@umijs/max';
import { useRequest } from 'ahooks';
import type { TabsProps } from 'antd';
import { Button, Tabs } from 'antd';
import { uniq } from 'lodash';

import { useMemo, useState } from 'react';
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
      label: intl.formatMessage({
        id: 'src.pages.Access.D6457915',
        defaultMessage: '用户',
      }),
      children: (
        <Accounts allAccounts={allAccounts} refreshAccounts={refreshAccounts} />
      ),
    },
    {
      key: ActiveKey.ROLES,
      label: intl.formatMessage({
        id: 'src.pages.Access.FB4D558D',
        defaultMessage: '角色',
      }),
      children: (
        <Roles
          allRoles={allRoles}
          allAccounts={allAccounts}
          refreshRoles={refreshRoles}
        />
      ),
    },
  ];

  const create = (type: ActiveKey) => {
    if (type === ActiveKey.ACCOUNTS) {
      setAccountVisible(true);
    } else {
      setModalVisible(true);
    }
  };

  const existingRole = useMemo(() => {
    return (
      uniq(
        allAccounts
          ?.map((account) => account.roles.map((role) => role.name))
          .flat(),
      ) || []
    );
  }, [allAccounts]);

  return (
    <PageContainer
      title={intl.formatMessage({
        id: 'src.pages.Access.DE5C1B7D',
        defaultMessage: '权限控制',
      })}
    >
      <Tabs
        tabBarExtraContent={
          access.acwrite ? (
            <Button type="primary" onClick={() => create(activeKey)}>
              {activeKey === ActiveKey.ACCOUNTS
                ? intl.formatMessage({
                    id: 'src.pages.Access.2FC6252B',
                    defaultMessage: '创建账户',
                  })
                : intl.formatMessage({
                    id: 'src.pages.Access.8D14D739',
                    defaultMessage: '创建角色',
                  })}
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
        existingRole={existingRole}
        type={Type.CREATE}
        successCallback={refreshRoles}
      />
    </PageContainer>
  );
}
