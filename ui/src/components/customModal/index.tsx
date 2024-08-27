import { intl } from '@/utils/intl';
import { Modal, ModalProps } from 'antd';

/**
 * By default, you are not allowed to click on the mask to close the pop-up window.
 */
export default function CustomModal({
  width = 520,
  children,
  maskClosable = false,
  ...props
}: ModalProps) {
  return (
    <Modal
      width={width}
      okText={intl.formatMessage({
        id: 'Dashboard.components.customModal.Ok',
        defaultMessage: '确定',
      })}
      maskClosable={maskClosable}
      cancelText={intl.formatMessage({
        id: 'Dashboard.components.customModal.Cancel',
        defaultMessage: '取消',
      })}
      {...props}
    >
      {children}
    </Modal>
  );
}
