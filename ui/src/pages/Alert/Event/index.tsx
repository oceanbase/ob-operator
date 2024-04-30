import { Card,Form } from 'antd';

import AlarmFilter from '../AlarmFilter';

export default function Event() {
  const [form] = Form.useForm();
  return (
    <div>
      <Card>
        <AlarmFilter form={form} type='event' />
      </Card>
    </div>
  );
}
