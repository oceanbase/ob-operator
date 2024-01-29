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
        message.success('激活备租户成功');
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
      title={'激活备租户'}
      isOpen={visible}
      handleOk={handleSubmit}
      handleCancel={handleCancel}
    >
      <p>备租户升主之后可接受外界读写请求，确定要激活吗？</p>
    </CustomModal>
  );
}
