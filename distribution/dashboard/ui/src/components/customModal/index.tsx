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

export type CommonModalType = {
  visible: boolean;
  setVisible: (prop: boolean) => void;
  successCallback: () => void;
};

export default function CustomModal(props: CustomModalProps) {
  const { isOpen, handleOk, handleCancel, title, width = 520 } = props;
  return (
    <Modal
      width={width}
      title={title}
      open={isOpen}
      onOk={handleOk}
      onCancel={handleCancel}
    >
      {props.children}
    </Modal>
  );
}
