import { intl } from '@/utils/intl';
import { useRequest } from 'ahooks';
import { message } from 'antd';

import { useParams } from '@umijs/max';
import { upgradeTenantCompatibilityVersion } from '@/services/tenant';
import type { CommonModalType } from '.';
import CustomModal from '.';

export default function UpgradeTenantModal({
  visible,
  setVisible,
  successCallback,
}: CommonModalType) {
  const { ns, name } = useParams();
  const { run: upgradeTenant } = useRequest(upgradeTenantCompatibilityVersion, {
    manual: true,
    onSuccess: ({ successful }) => {
      if (successful) {
        message.success(
          intl.formatMessage({
            id: 'Dashboard.components.customModal.UpgradeTenantModal.TheTenantVersionHasBeen',
            defaultMessage: '租户版本升级成功',
          }),
        );
        setVisible(false);
        if(successCallback) successCallback();
      }
    },
  });
  const handleSubmit = async () => {
    await upgradeTenant({ ns, name });
  };
  const handleCancel = () => setVisible(false);

  return (
    <CustomModal
      title={intl.formatMessage({
        id: 'Dashboard.components.customModal.UpgradeTenantModal.TenantVersionUpgrade',
        defaultMessage: '租户版本升级',
      })}
      isOpen={visible}
      handleOk={handleSubmit}
      handleCancel={handleCancel}
    >
      <p>
        {intl.formatMessage({
          id: 'Dashboard.components.customModal.UpgradeTenantModal.TheCurrentTenantVersionIs',
          defaultMessage:
            '当前租户的版本为 xxxx，集群的版本为xxx，确定升级租户吗？',
        })}
      </p>
    </CustomModal>
  );
}
