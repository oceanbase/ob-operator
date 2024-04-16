import { intl } from '@/utils/intl';
import { Modal } from 'antd';
import { ReactNode } from 'react';

interface CustomModalProps {
  isOpen: boolean;
  title: string;
  handleOk: () => void;
  handleCancel: () => void;
  children: ReactNode;
  width?: number;
}

export default function CustomModal(props: CustomModalProps) {
  const { isOpen, handleOk, handleCancel, title, width = 520 } = props;
  return (
    <Modal
      width={width}
      title={title}
      open={isOpen}
      onOk={handleOk}
      okText={intl.formatMessage({
        id: 'Dashboard.components.customModal.Ok',
        defaultMessage: '确定',
      })}
      cancelText={intl.formatMessage({
        id: 'Dashboard.components.customModal.Cancel',
        defaultMessage: '取消',
      })}
      onCancel={handleCancel}
    >
      {props.children}
    </Modal>
  );
}
