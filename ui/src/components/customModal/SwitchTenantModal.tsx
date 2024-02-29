import { intl } from '@/utils/intl';
import { useRequest } from 'ahooks';
import { message } from 'antd';

import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { changeTenantRole } from '@/services/tenant';
import type { CommonModalType } from '.';
import CustomModal from '.';

export default function SwitchTenantModal({
  visible,
  setVisible,
  successCallback,
}: CommonModalType) {
  const { run: activateTenant } = useRequest(changeTenantRole, {
    manual: true,
    onSuccess: ({ successful }) => {
      if (successful) {
        message.success(
          intl.formatMessage({
            id: 'Dashboard.components.customModal.SwitchTenantModal.TheStandbyTenantHasBeen',
            defaultMessage: '激活备租户成功',
          }),
        );
        setVisible(false);
        successCallback();
      }
    },
  });
  const handleSubmit = async () => {
    const [ns, name] = getNSName();
    await activateTenant({ ns, name });
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
