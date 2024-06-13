import { intl } from '@/utils/intl';
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
        title: intl.formatMessage({
          id: 'src.pages.OBProxy.New.00532253',
          defaultMessage: '创建 OBProxy 集群',
        }),
        onBack: () => {
          navigate('/obproxy');
        },
      }}
      footer={[
        <Button onClick={() => navigate('/obproxy')}>
          {intl.formatMessage({
            id: 'src.pages.OBProxy.New.3A4BDC5D',
            defaultMessage: '取消',
          })}
        </Button>,
        <Button>
          {intl.formatMessage({
            id: 'src.pages.OBProxy.New.E2247569',
            defaultMessage: '提交',
          })}
        </Button>,
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
