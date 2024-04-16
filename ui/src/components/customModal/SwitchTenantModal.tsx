import { intl } from '@/utils/intl';
import { useRequest } from 'ahooks';
import { message } from 'antd';

import { changeTenantRole } from '@/services/tenant';
import { useParams } from '@umijs/max';

import CustomModal from '.';

export default function SwitchTenantModal({
  visible,
  setVisible,
  successCallback,
}: API.CommonModalType) {
  const { ns, name } = useParams();
  const { run: activateTenant } = useRequest(changeTenantRole, {
    manual: true,
    onSuccess: ({ successful }) => {
      if (successful) {
        message.success(
          intl.formatMessage({
            id: 'Dashboard.components.customModal.SwitchTenantModal.OperationSucceeded',
            defaultMessage: '操作成功',
          }),
        );
        setVisible(false);
        if (successCallback) successCallback();
      }
    },
  });
  const handleSubmit = async () => {
    await activateTenant({ ns, name, switchover: true });
  };
  const handleCancel = () => setVisible(false);

  return (
    <CustomModal
      title={intl.formatMessage({
        id: 'Dashboard.components.customModal.SwitchTenantModal.ActiveStandbySwitchover',
        defaultMessage: '主备切换',
      })}
      isOpen={visible}
      handleOk={handleSubmit}
      handleCancel={handleCancel}
    >
      <p>
        {intl.formatMessage({
          id: 'Dashboard.components.customModal.SwitchTenantModal.AfterThePrimaryAndSecondary',
          defaultMessage:
            '主备切换后，原备租户将成为新的主租户，原主租户将成为新的备租户，确定要切换吗？',
        })}
      </p>
    </CustomModal>
  );
}
