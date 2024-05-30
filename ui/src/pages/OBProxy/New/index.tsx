import { PageContainer } from '@ant-design/pro-components';
import { useNavigate } from '@umijs/max';
import { Button, Col, Form, Row } from 'antd';
import BasicConfig from './BasicConfig';

export default function New() {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  return (
    <PageContainer
      header={{
        title: '创建 OBProxy 集群',
        onBack: () => {
          navigate('/obproxy');
        },
      }}
      footer={[
        <Button onClick={() => navigate('/obproxy')}>取消</Button>,
        <Button>提交</Button>,
      ]}
    >
      <Form form={form}>
        <Row gutter={[16, 16]}>
          <Col span={24}>
            <BasicConfig form={form} />
          </Col>
        </Row>
      </Form>
    </PageContainer>
  );
}
