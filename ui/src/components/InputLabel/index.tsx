import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import { Button, Checkbox, Col, Form, Input, Popconfirm, Row } from 'antd';

interface InputLabelProps {
  wrapFormName: string;
  labelFormName: string;
  valueFormName: string;
  showRegBox?: boolean;
  showDelIcon?: boolean;
}

export default function InputLabel({
  wrapFormName,
  labelFormName,
  valueFormName,
  showRegBox = false,
  showDelIcon = true,
}: InputLabelProps) {
  return (
    <Form.List name={wrapFormName}>
      {(fields, { add, remove }) => {
        return (
          <div>
            {fields.map(({ key, name }, index) => (
              <Row gutter={8} style={{ marginBottom: 8 }} key={key}>
                <Col span={12}>
                  <Form.Item name={[name, labelFormName]} noStyle>
                    <Input placeholder="请输入标签名" />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name={[name, valueFormName]} noStyle>
                    <Input placeholder="请输入标签值" />
                  </Form.Item>
                </Col>
                {showRegBox && (
                  <Col span={4}>
                    <Form.Item name={[name, 'isRegex']} noStyle>
                      <Checkbox />
                      正则
                    </Form.Item>
                  </Col>
                )}
                {showDelIcon && fields.length > 1 && (
                  <Form.Item
                    label={index === 0 && ' '}
                    style={{ marginBottom: 8 }}
                    name={[name, ' ']}
                  >
                    <Popconfirm
                      placement="left"
                      title="确定要删除该配置项吗？"
                      onConfirm={() => {
                        remove(index);
                      }}
                      okText="删除"
                      cancelText="取消"
                      okButtonProps={{
                        danger: true,
                        ghost: true,
                      }}
                    >
                      <DeleteOutlined style={{ color: 'rgba(0, 0, 0, .45)' }} />
                    </Popconfirm>
                  </Form.Item>
                )}
              </Row>
            ))}
            <Row>
              <Col span={20}>
                <Form.Item>
                  <Button
                    type="dashed"
                    block
                    onClick={() => {
                      const temp = { labelFormName: '', valueFormName: '' };
                      if (showRegBox) temp.isRegex = false;
                      add(temp);
                    }}
                    style={{ color: 'rgba(0,0,0,0.65)' }}
                  >
                    <PlusOutlined />
                    添加
                  </Button>
                </Form.Item>
              </Col>
            </Row>
          </div>
        );
      }}
    </Form.List>
  );
}
