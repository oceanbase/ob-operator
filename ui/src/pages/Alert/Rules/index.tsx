import { Card, Form } from 'antd';
import AlarmFilter from '../AlarmFilter';

export default function Rules() {
  const [form] = Form.useForm();
  return (
    <div>
      <Card>
        <AlarmFilter form={form} type="rules" />
      </Card>
    </div>
  );
}
