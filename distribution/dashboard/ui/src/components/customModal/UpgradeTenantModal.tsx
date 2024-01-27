import { useRequest } from 'ahooks';
import { message } from 'antd';

import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { changeTenantRole } from '@/services/tenant';
import type { CommonModalType } from '.';
import CustomModal from '.';

export default function UpgradeTenantModal({
  visible,
  setVisible,
  successCallback,
}: CommonModalType) {
  const { run: activateTenant } = useRequest(changeTenantRole, {
    manual: true,
    onSuccess: ({ successful }) => {
      if (successful) {
        message.success('租户版本升级成功');
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
      title="租户版本升级"
      isOpen={visible}
      handleOk={handleSubmit}
      handleCancel={handleCancel}
    >
      <p>当前租户的版本为 xxxx，集群的版本为xxx，确定升级租户吗？</p>
    </CustomModal>
  );
}
