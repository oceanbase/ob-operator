import { intl } from '@/utils/intl';
import { getLocale, useModel } from '@umijs/max';
import type { DescriptionsProps } from 'antd';
import { Button, Descriptions } from 'antd';
import CustomModal from '.';

export default function MyInfoModal({
  visible,
  setVisible,
}: API.CommonModalType) {
  const { initialState = {} } = useModel('@@initialState');
  const { accountInfo } = initialState;
  const locale = getLocale();
  const items: DescriptionsProps['items'] = [
    {
      key: 'nickname',
      label: intl.formatMessage({
        id: 'src.components.customModal.550B1F12',
        defaultMessage: '昵称',
      }),
      children: `${accountInfo?.nickname || '-'}`,
    },
    {
      key: 'username',
      label: intl.formatMessage({
        id: 'src.components.customModal.C5B6CB95',
        defaultMessage: '用户名',
      }),
      children: `${accountInfo?.username || '-'}`,
    },
    {
      key: 'roles',
      label: intl.formatMessage({
        id: 'src.components.customModal.23C67DC1',
        defaultMessage: '角色',
      }),
      children: `${
        accountInfo?.roles.map((role) => role.name).join(',') || '-'
      } `,
    },
    {
      key: 'description',
      label: intl.formatMessage({
        id: 'src.components.customModal.60C77BE0',
        defaultMessage: '描述',
      }),
      children: `${accountInfo?.description || '-'}`,
    },
    {
      key: 'lastLoginAt',
      label: intl.formatMessage({
        id: 'src.pages.Access.B52ED09C',
        defaultMessage: '上次登录于',
      }),
      children: `${accountInfo?.lastLoginAt || '-'}`,
    },
  ];
  const labelWidth = locale === 'en-US' ? 110 : 84;
  return (
    <CustomModal
      open={visible}
      onCancel={() => {
        setVisible(false);
      }}
      maskClosable={true}
      footer={
        <Button
          type="primary"
          onClick={() => {
            setVisible(false);
          }}
        >
          {intl.formatMessage({
            id: 'src.components.customModal.D1F5DD19',
            defaultMessage: '确定',
          })}
        </Button>
      }
    >
      <Descriptions
        title={intl.formatMessage({
          id: 'src.components.customModal.FCCAF16C',
          defaultMessage: '我的信息',
        })}
        labelStyle={{ width: labelWidth, justifyContent: 'end' }}
        column={1}
        items={items}
      />
    </CustomModal>
  );
}
