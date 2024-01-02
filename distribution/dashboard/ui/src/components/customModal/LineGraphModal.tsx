import { Modal } from 'antd';
import React from 'react';

interface LineGraphModalProps {
  visible: boolean;
  setVisible: (prop: boolean) => void;
  width?:number;
  title?:React.ReactNode
}

export default function LineGraphModal(
  props: LineGraphModalProps & React.PropsWithChildren,
) {
  const { visible,setVisible,width,title } = props;
  return <Modal title={title} width={width} footer={false} onCancel={() => setVisible(false) }open={visible}>{props.children}</Modal>;
}
