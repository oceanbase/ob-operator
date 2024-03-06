import { InputNumber as AntdInputNumber } from 'antd';

export default function InputNumber(props: any) {
  return <AntdInputNumber min={0} {...props} />;
}
