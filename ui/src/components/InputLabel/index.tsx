import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import { Button, Checkbox, Col, Form, Input, Popconfirm, Row } from 'antd';
import styles from './index.less';

interface InputLabelProps {
  wrapFormName: string;
  labelFormName: string;
  valueFormName: string;
  regBoxFormName?: string;
  showDelete?: boolean;
  maxCount?: number;
}

export default function InputLabel({
  wrapFormName,
  labelFormName,
  valueFormName,
  regBoxFormName,
  showDelete = true,
  maxCount,
}: InputLabelProps) {
  return (
    <Form.List name={wrapFormName}>
      {(fields, { add, remove }) => {
        return (
          <div>
            {fields.map(({ key, name }, index) => (
              <Row gutter={8} style={{ marginBottom: 8 }} key={key}>
                <Col span={11}>
                  <Form.Item name={[name, labelFormName]} noStyle>
                    <Input placeholder="请输入标签名" />
                  </Form.Item>
                </Col>
                <Col span={10}>
                  <Form.Item name={[name, valueFormName]} noStyle>
                    <Input placeholder="请输入标签值" />
                  </Form.Item>
                </Col>
                {regBoxFormName && (
                  <Col span={2}>
                    <Form.Item name={[name, regBoxFormName]} noStyle>
                      <Checkbox style={{ marginRight: 4 }} />
                      正则
                    </Form.Item>
                  </Col>
                )}
                {showDelete && fields.length > 1 && (
                  <Col span={1}>
                    <Form.Item className={styles.delContent} name={[name, ' ']}>
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
                        <DeleteOutlined
                          style={{ color: 'rgba(0, 0, 0, .45)' }}
                        />
                      </Popconfirm>
                    </Form.Item>
                  </Col>
                )}
              </Row>
            ))}
            {!maxCount || fields.length < maxCount ? (
              <Row>
                <Col span={24}>
                  <Form.Item>
                    <Button
                      type="dashed"
                      block
                      onClick={() => {
                        const temp = { labelFormName: '', valueFormName: '' };
                        if (regBoxFormName) temp[regBoxFormName] = false;
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
            ) : null}
          </div>
        );
      }}
    </Form.List>
  );
}
