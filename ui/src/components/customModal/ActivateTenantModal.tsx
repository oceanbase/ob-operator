import { intl } from '@/utils/intl';
import { useRequest } from 'ahooks';
import { message } from 'antd';

import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { changeTenantRole } from '@/services/tenant';
import type { CommonModalType } from '.';
import CustomModal from '.';

export default function ActivateTenantModal({
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
            id: 'Dashboard.components.customModal.ActivateTenantModal.TheStandbyTenantHasBeen',
            defaultMessage: '激活备租户成功',
          }),
        );
        setVisible(false);
        if(successCallback)successCallback();
      }
    },
  });
  const handleSubmit = async () => {
    const [ns, name] = getNSName();
    await activateTenant({ ns, name, failover: true });
  };
  const handleCancel = () => setVisible(false);

  return (
    <CustomModal
      title={intl.formatMessage({
        id: 'Dashboard.components.customModal.ActivateTenantModal.ActivateASecondaryTenant',
        defaultMessage: '激活备租户',
      })}
      isOpen={visible}
      handleOk={handleSubmit}
      handleCancel={handleCancel}
    >
      <p>
        {intl.formatMessage({
          id: 'Dashboard.components.customModal.ActivateTenantModal.AfterTheStandbyTenantIs',
          defaultMessage: '备租户升主之后可接受外界读写请求，确定要激活吗？',
        })}
      </p>
    </CustomModal>
  );
}
