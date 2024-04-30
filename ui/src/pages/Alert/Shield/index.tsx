import { Card, Form } from 'antd';
import AlarmFilter from '../AlarmFilter';

export default function Shield() {
  const [form] = Form.useForm();
  return (
    <div>
      <Card>
        <AlarmFilter form={form} type="shield" />
      </Card>
    </div>
  );
}
