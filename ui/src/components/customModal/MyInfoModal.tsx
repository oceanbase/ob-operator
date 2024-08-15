import { useModel } from '@umijs/max';
import type { DescriptionsProps } from 'antd';
import { Button, Descriptions } from 'antd';
import CustomModal from '.';

export default function MyInfoModal({
  visible,
  setVisible,
}: API.CommonModalType) {
  const { initialState = {} } = useModel('@@initialState');
  const { accountInfo } = initialState;

  const items: DescriptionsProps['items'] = [
    {
      key: 'nickname',
      label: '昵称',
      children: `${accountInfo?.nickname || '-'}`,
    },
    {
      key: 'username',
      label: '用户名',
      children: `${accountInfo?.username || '-'}`,
    },
    {
      key: 'roles',
      label: '角色',
      children: `${
        accountInfo?.roles.map((role) => role.name).join(',') || '-'
      } `,
    },
    {
      key: 'lastLoginAt',
      label: '最近一次登录',
      children: `${accountInfo?.lastLoginAt || '-'}`,
    },
  ];
  return (
    <CustomModal
      isOpen={visible}
      handleCancel={() => {
        setVisible(false);
      }}
      footer={
        <Button
          type="primary"
          onClick={() => {
            setVisible(false);
          }}
        >
          确定
        </Button>
      }
    >
      <Descriptions title="我的信息" items={items} />
    </CustomModal>
  );
}
