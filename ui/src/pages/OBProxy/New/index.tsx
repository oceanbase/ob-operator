import { obproxy } from '@/api';
import { ObproxyCreateOBProxyParam } from '@/api/generated';
import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { useNavigate } from '@umijs/max';
import { Button, Col, Form, Row, message } from 'antd';
import BasicConfig from './BasicConfig';
import DetailConfig from './DetailConfig';

type FormValues = {
  obCluster: string;
} & ObproxyCreateOBProxyParam;

export default function New() {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const submit = async (values: FormValues) => {
    try {
      const res = await obproxy.createOBPROXY({
        ...values,
        obCluster: JSON.parse(values.obCluster),
      });
      if (res.successful) {
        message.success('创建成功！', 3);
        form.resetFields();
        history.back();
      }
    } catch (err) {
      console.error('err:', err);
    }
  };
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
        <Button onClick={() => form.submit()} type="primary">
          {intl.formatMessage({
            id: 'src.pages.OBProxy.New.E2247569',
            defaultMessage: '提交',
          })}
        </Button>,
      ]}
    >
      <Form onFinish={submit} form={form}>
        <Row gutter={[16, 16]}>
          <Col span={24}>
            <BasicConfig form={form} />
          </Col>
          <Col span={24}>
            <DetailConfig form={form} />
          </Col>
        </Row>
      </Form>
    </PageContainer>
  );
}
